// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/xerr"
)

// checkCallTargets statically checks the callee of every call whose callee is
// a bare identifier, resolving it as the runtime's getTarget does (#161):
//
//   - a value binding (localCallee) must not be a rule, must not be memoized,
//     must not have a statically known kind other than callable, and, when it
//     is a let bound directly to a lambda, must get an argument count the
//     lambda accepts;
//   - otherwise a visible derive or a builtin is accepted (builtin calls are
//     checked by checkBuiltinCalls);
//   - anything else is an unknown function, reported with the module alias
//     when the policy imports the name with `use`.
//
// Slash-qualified derive calls and `alias.fn` module calls are not identifier
// callees and are left to their own checks. Derive bodies are walked too, but
// validateDerivePurity runs first and is stricter there.
func (idx *Index) checkCallTargets(ctx context.Context) error {
	var errs []error
	err := idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		call, ok := e.(*ast.CallExpression)
		if !ok {
			return
		}
		id, ok := call.Callee.(*ast.Identifier)
		if !ok {
			return
		}
		if err := k.checkCallTarget(call, id); err != nil {
			errs = append(errs, err)
		}
	})
	if err != nil {
		return err
	}
	return errors.Join(errs...)
}

func (k *kindCheckCtx) checkCallTarget(call *ast.CallExpression, id *ast.Identifier) error {
	name := id.Value
	if info, isRule, ok := k.localCallee(name); ok {
		switch {
		case isRule:
			return xerr.ErrCallTarget(id.Span(), fmt.Sprintf("cannot call rule %q: a rule is not a function", name))
		case call.Memoized:
			return xerr.ErrCallTarget(id.Span(), fmt.Sprintf("%q is a local value and cannot be memoized with '!': memoization applies to derives, builtins and module functions", name))
		case info.lambda != nil:
			return checkLambdaCallArity(name, info.lambda, call)
		}
		if kind, known := k.resolveIdentKind(id); known && kind != box.ValueCallable {
			return xerr.ErrCallTarget(id.Span(), fmt.Sprintf("cannot call %q: it is a %s value, not a function", name, kind))
		}
		return nil
	}
	if k.isShadowed(name) || k.lookupDerive(name) != nil {
		return nil
	}
	if _, ok := builtins.Table[name]; ok {
		return nil
	}
	if k.policy != nil {
		for _, alias := range slices.Sorted(maps.Keys(k.policy.Uses)) {
			if use := k.policy.Uses[alias]; use != nil && slices.Contains(use.Modules, name) {
				return xerr.ErrCallTarget(id.Span(), fmt.Sprintf("%q is imported from module alias %q and must be called as %s.%s(...)", name, alias, alias, name))
			}
		}
	}
	return xerr.ErrCallTarget(id.Span(), fmt.Sprintf("unknown function %q: it is not a let, parameter, fact, derive or builtin in scope", name))
}

// checkLambdaCallArity mirrors padAndValidateCallableArgs in the runtime: at
// most one argument per parameter, and at least one per required parameter.
func checkLambdaCallArity(name string, lam *ast.LambdaExpression, call *ast.CallExpression) error {
	max := len(lam.Params)
	min := ast.RequiredLambdaArity(lam)
	got := len(call.Arguments)
	switch {
	case got > max:
		return xerr.ErrCallTarget(call.Arguments[max].Span(), fmt.Sprintf("%q takes at most %d argument(s), got %d", name, max, got))
	case got < min:
		return xerr.ErrCallTarget(call.Callee.Span(), fmt.Sprintf("%q takes at least %d argument(s), got %d", name, min, got))
	}
	return nil
}
