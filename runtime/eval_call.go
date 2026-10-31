// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
)

func evalCall(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.CallExpression) (response box.Value, traceNode *trace.Node, err error) {
	ctx, n, done := trace.New(ctx, t, "call", map[string]any{
		"target": t.Callee.String(),
		"args":   t.Arguments,
	})
	defer done()

	args := make([]box.Value, 0, len(t.Arguments))
	for _, a := range t.Arguments {
		v, child, err := eval(ctx, ec, exec, p, a)
		n.Attach(child)
		if err != nil {
			return box.Undefined(), n.SetErr(err), err
		}
		args = append(args, v)
	}

	if t.Memoized {
		for _, a := range args {
			if a.IsCallable() {
				err := fmt.Errorf("memoized call cannot take callable arguments")
				return box.Undefined(), n.SetErr(err), err
			}
		}
	}

	target, err := getTarget(ctx, ec, exec, p, t)
	if err != nil {
		return box.Undefined(), n.SetErr(err), err
	}

	// use a thin wrapper around the target to handle the caching
	wrappedTarget := func(ctx context.Context, args ...box.Value) (box.Value, error) {
		if !t.Memoized {
			return target(ctx, args...)
		}

		ttl := 5 * time.Minute // default to 5 minutes
		if t.MemoizeTTL != nil {
			ttl = *t.MemoizeTTL
		}

		hashKey, hashErr := calculateHashKey(t, args)
		if hashErr != nil {
			// Arguments cannot be hashed stably; bypass memoization.
			return target(ctx, args...)
		}
		if isClockDependentCall(ec, p, t) {
			// The memo cache is executor-wide, so without this a cached
			// answer would outlive the execution whose clock produced it.
			hashKey += fmt.Sprintf("@%d", ec.CreatedAt().UnixNano())
		}
		loader := func(ctx context.Context, key string) (any, error) {
			return target(ctx, args...)
		}
		out, _, err := exec.callMemoizePerch.Get(ctx, hashKey, ttl, loader)
		if err != nil {
			return box.Undefined(), err
		}
		v, ok := out.(box.Value)
		if ok {
			return v, nil
		}
		return box.FromAny(out), nil
	}

	// call the target
	out, err := wrappedTarget(ctx, args...)
	if err != nil {
		if errors.Is(err, xerr.InjectedError{}) || errors.Is(err, ErrCallDepthExceeded) || errors.Is(err, ErrEvalDepthExceeded) {
			// An injected error keeps the user's message. A depth error is
			// returned as is, or each of the up to MaxCallDepth levels would
			// add its own "failed to call function" prefix.
			return box.Undefined(), n.SetErr(err), err
		}
		err = fmt.Errorf("failed to call function '%s': %w", t.Callee.String(), err)
		return box.Undefined(), n.SetErr(err), err
	}
	return out, n.SetResult(out), nil
}

// clockDependentBuiltins and clockDependentModuleCalls name the callees whose
// answer depends on the execution start time: the clock itself, and the JWT
// checks that compare exp and nbf against it. Memoizing them across executions
// would return an earlier execution's clock (an expired token still verifying).
// now_ms is listed ahead of the builtin itself (#153 adds it) so that layer
// needs no change here.
var clockDependentBuiltins = map[string]bool{"now": true, "now_ms": true}

var clockDependentModuleCalls = map[string]map[string]bool{
	"@" + constants.APPNAME + "/time": {"now": true},
	"@" + constants.APPNAME + "/jwt":  {"decode": true, "verify": true},
}

// isClockDependentCall reports whether t calls one of the clock-dependent
// builtins or module functions above.
func isClockDependentCall(ec *ExecutionContext, p *index.Policy, t *ast.CallExpression) bool {
	if id, ok := t.Callee.(*ast.Identifier); ok {
		return clockDependentBuiltins[id.Value] && lookupDeriveByIdentifier(ec, p, id.Value) == nil
	}
	alias, fn := splitAliasFn(t.Callee.String())
	if alias == "" || fn == "" {
		return false
	}
	binding, ok := ec.Module(alias)
	if !ok || binding == nil {
		return false
	}
	return clockDependentModuleCalls[binding.CanonicalKey][fn]
}

// Helper to split "alias.fn" if ever needed
func splitAliasFn(s string) (string, string) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return s, ""
	}
	return parts[0], parts[1]
}

// calculateHashKey hashes memoization arguments and prefixes the AST call-site pointer.
// The pointer is stable for the lifetime of the loaded index; the executor holds the index
// and the memoization cache, so keys do not outlive their call sites.
func calculateHashKey(node *ast.CallExpression, args []box.Value) (string, error) {
	hashArgs := make([]any, 0, len(args))
	for _, a := range args {
		// Callables are rejected for memoized calls before we get here.
		h, err := box.TryToBoundaryAny(a)
		if err != nil {
			return "", err
		}
		hashArgs = append(hashArgs, h)
	}
	arghash, err := hashstructure.Hash(hashArgs, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%p:%d", node, arghash), nil
}

func lookupDeriveByIdentifier(ec *ExecutionContext, p *index.Policy, name string) *index.Derive {
	if ec.evalDerive != nil {
		if d, ok := ec.evalDerive.DefineShort[name]; ok {
			return d
		}
		return nil
	}
	if p == nil {
		return nil
	}
	if d, ok := p.Derives[name]; ok {
		return d
	}
	if p.Namespace != nil {
		if d, ok := p.Namespace.Derives[name]; ok {
			return d
		}
	}
	return nil
}

func callerNamespaceFQNForDeriveExport(ec *ExecutionContext, p *index.Policy) string {
	if ec.evalDerive != nil && ec.evalDerive.Namespace != nil {
		return ec.evalDerive.Namespace.FQN.String()
	}
	if p != nil && p.Namespace != nil {
		return p.Namespace.FQN.String()
	}
	return ""
}

func callerPolicyForDeriveScope(ec *ExecutionContext, p *index.Policy) *index.Policy {
	if ec.evalDerive != nil && ec.evalDerive.Policy != nil {
		return ec.evalDerive.Policy
	}
	return p
}

func enforceDerivePolicyScopeForCaller(ec *ExecutionContext, p *index.Policy, d *index.Derive) error {
	return d.VisibleFromPolicy(callerPolicyForDeriveScope(ec, p))
}

func enforceDeriveExportForCaller(ec *ExecutionContext, p *index.Policy, d *index.Derive) error {
	if d == nil || d.Namespace == nil {
		return nil
	}
	if err := enforceDerivePolicyScopeForCaller(ec, p, d); err != nil {
		return err
	}
	caller := callerNamespaceFQNForDeriveExport(ec, p)
	if caller == "" {
		return nil
	}
	if d.Namespace.FQN.String() != caller {
		return d.Namespace.VerifyDeriveExported(d.Name)
	}
	return nil
}

func lookupDeriveBySlashFQ(ec *ExecutionContext, exec *executorImpl, p *index.Policy, fqn string) (*index.Derive, error) {
	parts := strings.Split(fqn, ast.FQNSeparator)
	if len(parts) < 3 {
		return nil, nil
	}
	var d *index.Derive
	if ec.evalDerive != nil {
		if d = ec.evalDerive.DefineFQN[fqn]; d != nil {
			if err := enforceDeriveExportForCaller(ec, p, d); err != nil {
				return nil, err
			}
			return d, nil
		}
		var ok bool
		d, ok = exec.index.DerivesByFQN[fqn]
		if !ok {
			return nil, fmt.Errorf("unknown derive %q", fqn)
		}
		if err := enforceDeriveExportForCaller(ec, p, d); err != nil {
			return nil, err
		}
		return d, nil
	}
	var err error
	d, err = exec.index.ResolveDerive(fqn)
	if err != nil {
		return nil, err
	}
	if err := enforceDeriveExportForCaller(ec, p, d); err != nil {
		return nil, err
	}
	return d, nil
}

// getTarget resolves a call's callee. The precedence is lexical, innermost
// first, and is mirrored by index.kindCheckCtx (isBuiltinCall, localCallee):
//
//  1. a slash-qualified derive FQN of three or more segments;
//  2. for an identifier callee, a value binding: a local (lambda parameter or
//     evaluated let), a let, and outside a derive body a fact or a rule. The
//     binding is evaluated like an identifier reference and must hold a
//     callable (#161), so a let or parameter shadows a derive or builtin of
//     the same name;
//  3. a visible derive by short name;
//  4. a builtin;
//  5. `alias.fn` against the policy's module bindings.
//
// Inside a derive body (ec.evalDerive != nil) every step decides explicitly:
// facts and rules are not bindings there, a bound callable must be pure (see
// bindingCallTarget), impure builtins and module calls are refused, and an
// unresolved callee is an error.
func getTarget(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, c *ast.CallExpression) (func(context.Context, ...box.Value) (box.Value, error), error) {
	if fqn := ast.SlashCalleeFQNS(c.Callee); fqn != "" {
		parts := strings.Split(fqn, ast.FQNSeparator)
		// Require at least three segments (e.g. com/ex/name) so a two-segment slash chain
		// is not mistaken for a derive FQN (see division vs slash-callee ambiguity).
		if len(parts) >= 3 {
			d, err := lookupDeriveBySlashFQ(ec, exec, p, fqn)
			if err != nil {
				return nil, err
			}
			if d != nil {
				return func(ctx context.Context, args ...box.Value) (box.Value, error) {
					return invokeDerive(ctx, ec, exec, p, d, args)
				}, nil
			}
		}
	}

	calleeStr := c.Callee.String()

	var nameForBuiltin string
	if id, ok := c.Callee.(*ast.Identifier); ok {
		if calleeIsBinding(ec, p, id.Value) {
			return bindingCallTarget(ctx, ec, exec, p, c, id)
		}
		if d := lookupDeriveByIdentifier(ec, p, id.Value); d != nil {
			return func(ctx context.Context, args ...box.Value) (box.Value, error) {
				return invokeDerive(ctx, ec, exec, p, d, args)
			}, nil
		}
		nameForBuiltin = id.Value
	} else {
		nameForBuiltin = calleeStr
	}

	if nameForBuiltin != "" {
		if decl, ok := builtins.Table[nameForBuiltin]; ok {
			// Paired with the static check in index.CheckCapabilities.
			if !exec.grant().AllowsBuiltin(nameForBuiltin) {
				return nil, fmt.Errorf("builtin %q: %w", nameForBuiltin, xerr.ErrNotGranted)
			}
			// Paired with the static check in index/derive_purity.go.
			if ec.evalDerive != nil && !decl.PermittedInDerive() {
				if decl.Nondeterministic {
					return nil, fmt.Errorf("builtin %q is nondeterministic and is not permitted inside a derive", nameForBuiltin)
				}
				return nil, fmt.Errorf("builtin %q is not permitted inside a derive - it cannot guarantee deterministic output within a single policy execution", nameForBuiltin)
			}
			return func(ctx context.Context, args ...box.Value) (box.Value, error) {
				site := &CallSite{EC: ec, Exec: exec, Policy: p}
				if handled, v, err := decl.Precheck(site, args); handled || err != nil {
					return v, err
				}
				return decl.Impl(ctx, site, args...)
			}, nil
		}
	}

	module, fn := splitAliasFn(calleeStr)
	if ec.evalDerive != nil && module != "" && fn != "" {
		return nil, fmt.Errorf("TypeScript module calls like %q are not permitted inside a derive - they cannot guarantee deterministic output within a single policy execution", calleeStr)
	}

	if module == "" || fn == "" {
		// Not a binding, derive, builtin or `alias.fn`: an unknown callee,
		// refused alike inside and outside a derive body. The import error is
		// kept for a name the policy imports from a module, which has to be
		// called through its alias.
		return nil, unresolvedCalleeError(p, calleeStr)
	}

	modulebinding, ok := ec.Module(module)
	if !ok {
		e := xerr.ErrModuleInvocation(module, fn)
		return nil, e
	}

	return func(ctx context.Context, args ...box.Value) (box.Value, error) {
		anyArgs := make([]any, 0, len(args))
		for _, a := range args {
			if a.IsCallable() {
				return box.Undefined(), fmt.Errorf("cannot pass callable value to module function %s.%s", module, fn)
			}
			x, err := box.TryToBoundaryAny(a)
			if err != nil {
				return box.Undefined(), fmt.Errorf("module call %s.%s: %w", module, fn, err)
			}
			anyArgs = append(anyArgs, x)
		}
		out, err := modulebinding.Call(ctx, ec, fn, anyArgs...)
		return box.FromBoundaryAny(out), err
	}, nil
}

// calleeIsBinding reports whether an identifier callee names a value binding,
// which takes precedence over a derive or builtin of the same name. The
// lookups are the ones evalIdent makes before it reaches derives: locals and
// lets, and outside a derive body facts (injected or declared) and rules.
// Inside a derive body facts and rules are not bindings, matching evalIdent's
// guards; the detached derive context holds neither anyway.
func calleeIsBinding(ec *ExecutionContext, p *index.Policy, name string) bool {
	if _, ok := ec.GetLocal(name); ok {
		return true
	}
	if _, ok := ec.GetLet(name); ok {
		return true
	}
	if ec.evalDerive != nil {
		return false
	}
	if _, ok := ec.GetFact(name); ok {
		return true
	}
	if p != nil {
		if _, ok := p.Facts[name]; ok {
			return true
		}
		if _, ok := p.Rules[name]; ok {
			return true
		}
	}
	return false
}

// bindingCallTarget calls the callable a value binding holds (#161). The
// binding is evaluated as an identifier reference would be, so a lazy let is
// evaluated, type-checked and cached, and a let that calls itself is reported
// as a self-reference.
//
// Inside a derive body only a pure callable may be called: a derive, or a
// lambda created while evaluating some derive body, whose captured context is
// a detached derive context without facts, rules or modules and whose body the
// purity check walked. A lambda created outside any derive (passed in as an
// argument) captures the caller's full context and is refused. The static
// side (index.checkDeriveCall) is stricter: it accepts only a let bound
// directly to a lambda literal, which always yields such a lambda.
//
// A memoized call through a binding is refused: the memo key is the call site
// and the arguments, so a lambda that captures facts or parameters would be
// answered from another execution's cache.
func bindingCallTarget(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, c *ast.CallExpression, id *ast.Identifier) (func(context.Context, ...box.Value) (box.Value, error), error) {
	name := id.Value
	if c.Memoized {
		return nil, fmt.Errorf("%q is a local value and cannot be memoized with '!': memoization applies to derives, builtins and module functions", name)
	}
	v, _, err := evalIdent(ctx, ec, exec, p, id)
	if err != nil {
		return nil, err
	}
	if !v.IsCallable() {
		return nil, fmt.Errorf("cannot call %q: it is a %s value, not a function", name, v.Kind())
	}
	callable, err := callableFromValue(v)
	if err != nil {
		return nil, err
	}
	if ec.evalDerive != nil {
		switch fn := callable.(type) {
		case *deriveCallable:
			// Derives are pure by construction.
		case *lambdaCallable:
			if fn.capture == nil || fn.capture.evalDerive == nil {
				return nil, fmt.Errorf("calling %q is not permitted inside a derive - the lambda was defined outside the derive and could read facts or call modules", name)
			}
		default:
			return nil, fmt.Errorf("calling %q is not permitted inside a derive - it is not a derive or a lambda defined in a derive", name)
		}
	}
	return func(ctx context.Context, args ...box.Value) (box.Value, error) {
		return callable.Invoke(ctx, &CallSite{EC: ec, Exec: exec, Policy: p}, args)
	}, nil
}

// unresolvedCalleeError reports a callee that names nothing callable. The
// "unable to resolve import" error is reserved for a name the policy imports
// with `use`, which is called through its module alias, not bare.
func unresolvedCalleeError(p *index.Policy, callee string) error {
	if p != nil {
		for _, alias := range slices.Sorted(maps.Keys(p.Uses)) {
			if use := p.Uses[alias]; use != nil && slices.Contains(use.Modules, callee) {
				return fmt.Errorf("%q is imported from module alias %q and must be called as %s.%s(...): %w", callee, alias, alias, callee, xerr.ErrImportResolution(alias, callee))
			}
		}
	}
	return fmt.Errorf("unknown function %q: it is not a let, parameter, fact, derive or builtin in scope", callee)
}
