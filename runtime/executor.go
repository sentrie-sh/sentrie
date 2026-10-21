// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	stdErr "errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/binaek/perch"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/js"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

type NewExecutorOption func(*executorImpl)

const (
	// DefaultEvalTimeout bounds one ExecRule or ExecPolicy call unless
	// overridden with WithEvalTimeout (#142). When it passes while a module call
	// or load is running, that JavaScript is interrupted and the evaluation fails
	// with context.DeadlineExceeded. Sentrie's own expression evaluation does not
	// check it (#225).
	DefaultEvalTimeout = 10 * time.Second

	// defaultCallMemoizeCacheEntries bounds the call memoization cache unless
	// overridden with WithCallMemoizeCacheSize.
	defaultCallMemoizeCacheEntries = 10_000
)

// WithEvalTimeout sets the deadline for each ExecRule or ExecPolicy call.
// A value <= 0 disables the executor's own deadline; the caller's context
// still applies.
func WithEvalTimeout(d time.Duration) NewExecutorOption {
	return func(e *executorImpl) {
		e.evalTimeout = d
	}
}

// WithCallMemoizeCacheSize sets the maximum number of entries in the call memoization cache.
func WithCallMemoizeCacheSize(entries int) NewExecutorOption {
	return func(e *executorImpl) {
		e.callMemoizePerch = perch.New[any](entries)
	}
}

type ExecutorOutput struct {
	PolicyName  string              `json:"policy"`
	Namespace   string              `json:"namespace"`
	RuleName    string              `json:"rule"`
	Decision    *Decision           `json:"decision"`
	Attachments DecisionAttachments `json:"attachments"`
	RuleNode    *trace.Node         `json:"trace"`
}

func (e *ExecutorOutput) ToTrinary() trinary.Value {
	return e.Decision.State
}

type Executor interface {
	ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) ([]*ExecutorOutput, error)
	ExecRule(ctx context.Context, namespace, policy, rule string, facts map[string]any) (*ExecutorOutput, error)
	Index() *index.Index
	// Close releases every loaded module. Evaluations started afterwards fail
	// with ErrExecutorClosed once they need a module. Close is idempotent.
	Close() error
}

// ErrExecutorClosed is returned when an evaluation needs a module after Close.
var ErrExecutorClosed = stdErr.New("executor is closed")

// moduleLoad is one module's entry in the executor's lifetime module map. done
// is closed once module and err are set, so concurrent first users wait on a
// single load (singleflight) rather than each loading the module.
type moduleLoad struct {
	done   chan struct{}
	module *jsModule
	err    error
}

// executorImpl ties together the index, JS loader, and evaluation.
type executorImpl struct {
	index            *index.Index
	jsRegistry       *js.Registry
	callMemoizePerch *perch.Perch[any]
	evalTimeout      time.Duration

	// modules holds every loaded module for the executor's lifetime, keyed by
	// module path. Its size is bounded by the pack's distinct `use` targets, so
	// it needs no eviction (#191).
	modulesMu     sync.Mutex
	modules       map[string]*moduleLoad
	modulesClosed bool
	// moduleLoads counts module loads started, for tests and diagnostics.
	moduleLoads atomic.Int64
}

// NewExecutor builds an Executor with built-in @sentra/* modules registered.
func NewExecutor(idx *index.Index, opts ...NewExecutorOption) (Executor, error) {
	exec := &executorImpl{
		index:            idx,
		jsRegistry:       js.NewRegistry(idx.Pack.Location),
		callMemoizePerch: perch.New[any](defaultCallMemoizeCacheEntries),
		evalTimeout:      DefaultEvalTimeout,
		modules:          map[string]*moduleLoad{},
	}

	exec.jsRegistry.RegisterGoBuiltin("uuid", js.BuiltinUuidGo)
	exec.jsRegistry.RegisterGoBuiltin("crypto", js.BuiltinCryptoGo)
	exec.jsRegistry.RegisterGoBuiltin("time", js.BuiltinTimeGo)
	exec.jsRegistry.RegisterGoBuiltin("encoding", js.BuiltinEncodingGo)
	exec.jsRegistry.RegisterGoBuiltin("collection", js.BuiltinCollectionGo)
	exec.jsRegistry.RegisterGoBuiltin("jwt", js.BuiltinJwtGo)
	exec.jsRegistry.RegisterGoBuiltin("regex", js.BuiltinRegexGo)
	exec.jsRegistry.RegisterGoBuiltin("net", js.BuiltinNetGo)
	exec.jsRegistry.RegisterGoBuiltin("hash", js.BuiltinHashGo)
	exec.jsRegistry.RegisterGoBuiltin("url", js.BuiltinUrlGo)
	exec.jsRegistry.RegisterGoBuiltin("string", js.BuiltinStringGo)
	exec.jsRegistry.RegisterGoBuiltin("json", js.BuiltinJsonGo)
	exec.jsRegistry.RegisterGoBuiltin("semver", js.BuiltinSemverGo)
	exec.jsRegistry.RegisterGoBuiltin("math", js.BuiltinMathGo)

	// Register TypeScript builtin module for JavaScript globals
	exec.jsRegistry.RegisterTSBuiltin("js", string(js.BuiltinJSTS))

	for _, opt := range opts {
		opt(exec)
	}

	return exec, nil
}

func (e *executorImpl) Index() *index.Index {
	return e.index
}

// Close drops every loaded module and rejects later module loads. A module call
// already running keeps its own VM and finishes normally; nothing it holds is
// shared with the modules being dropped.
func (e *executorImpl) Close() error {
	e.modulesMu.Lock()
	defer e.modulesMu.Unlock()
	if e.modulesClosed {
		return nil
	}
	e.modulesClosed = true
	for _, l := range e.modules {
		select {
		case <-l.done:
			if l.module != nil {
				l.module.close()
			}
		default:
			// Still loading: the loader sees modulesClosed and closes it.
		}
	}
	e.modules = nil
	return nil
}

// withEvalDeadline derives the per-evaluation deadline from ctx. The returned
// done releases it and, when this deadline (not the caller's) is what ended the
// evaluation, says so in the error. A nested ExecRule (a cross-policy import)
// inherits whichever deadline is earlier.
func (e *executorImpl) withEvalDeadline(parent context.Context) (context.Context, func(error) error) {
	if e.evalTimeout <= 0 {
		return parent, func(err error) error { return err }
	}
	ctx, cancel := context.WithTimeout(parent, e.evalTimeout)
	return ctx, func(err error) error {
		defer cancel()
		if err != nil && parent.Err() == nil && stdErr.Is(ctx.Err(), context.DeadlineExceeded) && stdErr.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("evaluation exceeded its %s deadline: %w", e.evalTimeout, err)
		}
		return err
	}
}

// ExecPolicy executes all exported rules and returns the results
func (e *executorImpl) ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) (_ []*ExecutorOutput, err error) {
	ctx, done := e.withEvalDeadline(ctx)
	defer func() { err = done(err) }()

	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, err
	}

	theLock := &sync.Mutex{}
	var compositeErr error
	outputs := make([]*ExecutorOutput, 0, len(p.RuleExports))
	wg := &sync.WaitGroup{}
	for exportedName, ruleExport := range p.RuleExports {
		ruleName := exportedName
		if ruleExport != nil && ruleExport.RuleName != "" {
			ruleName = ruleExport.RuleName
		}
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					theLock.Lock()
					defer theLock.Unlock()
					compositeErr = stdErr.Join(compositeErr, fmt.Errorf("panic in ExecRule %q: %v", ruleName, r))
				}
			}()

			output, err := e.ExecRule(ctx, namespace, policy, ruleName, facts)

			// now that we have the output, we can add it to the outputs slice,
			// but we need to lock the mutex to avoid race conditions
			theLock.Lock()
			defer theLock.Unlock()
			if err != nil {
				compositeErr = stdErr.Join(compositeErr, err)
				return
			}

			// add the output to the outputs slice
			outputs = append(outputs, output)
		})
	}
	wg.Wait()

	return outputs, compositeErr
}

// ExecRule executes an exported rule and returns the result
func (e *executorImpl) ExecRule(ctx context.Context, namespace, policy, rule string, injectedFacts map[string]any) (_ *ExecutorOutput, err error) {
	ctx, done := e.withEvalDeadline(ctx)
	defer func() { err = done(err) }()

	// Validate exported
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, err
	}
	if err := p.VerifyRuleExported(rule); err != nil {
		return nil, err
	}

	ec := NewExecutionContext(p, e)
	defer ec.Dispose()

	for factName, factStatement := range p.Facts {
		// look for a value for this fact in the passed in facts map
		factValue, ok := injectedFacts[factName]

		// we do not have a value for this fact, and it is required - error
		if !ok && !factStatement.Optional {
			return nil, xerr.ErrRequiredFact(factName)
		}

		if ok {
			decodedFactValue := box.FromBoundaryAny(factValue)
			if decodedFactValue.IsNull() && !ast.IsNullableTypeRef(factStatement.Type) {
				return nil, fmt.Errorf("fact '%s' cannot be null: %w", factName, xerr.ErrInvalidInvocation(""))
			}
			err := ec.InjectFact(ctx, factName, decodedFactValue, false, factStatement.Type)
			if err != nil {
				return nil, err
			}
			continue // move on to the next fact
		}

		// if the fact has a default value, evaluate it and inject it into the context
		if factStatement.Default != nil {
			// evaluate the default value, this will be injected into the context
			val, _, err := eval(ctx, ec, e, p, factStatement.Default)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", err.Error(), xerr.ErrUnresolvableFact(factName))
			}

			if val.IsNull() && !ast.IsNullableTypeRef(factStatement.Type) {
				return nil, fmt.Errorf("fact '%s' cannot have null default value: %w", factName, xerr.ErrInvalidInvocation(""))
			}

			// inject the default value
			if err := ec.InjectFact(ctx, factName, val, true, factStatement.Type); err != nil {
				return nil, err
			}
		}
	}

	// bind lets
	for k, v := range p.Lets {
		if err := ec.InjectLet(k, v); err != nil {
			return nil, err
		}
	}

	// Bind `use` modules
	if err := e.bindUses(ctx, ec, p); err != nil {
		return nil, err
	}

	decision, attachments, ruleNode, err := e.execRule(ctx, ec, namespace, policy, rule)
	if err != nil && decision == nil {
		decision = DecisionOf(box.Trinary(trinary.Unknown))
	}
	return &ExecutorOutput{
		PolicyName:  policy,
		Namespace:   namespace,
		RuleName:    rule,
		Decision:    decision,
		Attachments: attachments,
		RuleNode:    ruleNode,
	}, err
}

func (e *executorImpl) execRule(ctx context.Context, ec *ExecutionContext, namespace, policy, rule string) (*Decision, DecisionAttachments, *trace.Node, error) {
	thePolicy, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, nil, nil, err
	}

	theRule, ok := thePolicy.Rules[rule]
	if !ok {
		return nil, nil, nil, xerr.ErrRuleNotFound(index.RuleFQN(namespace, policy, rule))
	}

	// Check for infinite recursion before evaluating the rule
	if err := ec.PushRefStack(theRule.FQN.String()); err != nil {
		return nil, nil, nil, err
	}
	defer ec.PopRefStack()

	// Wrap rule evaluation in a decision node
	ctx, ruleNode, done := trace.New(ctx, theRule.Node, "rule-outcome", map[string]any{
		"namespace": namespace,
		"policy":    policy,
		"rule":      rule,
	})
	defer done()

	// validate the facts against the type
	for name, fact := range ec.facts {
		if fact.typeRef == nil {
			// if there's no shape indication, we skip validation
			continue
		}
		stmt := thePolicy.Facts[name]
		// validate the value against the type
		if err := validateValueAgainstTypeRef(ctx, ec, e, thePolicy, fact.value, fact.typeRef, stmt.Span()); err != nil {
			return nil, nil, nil, err
		}
	}

	d, node, err := evaluateRuleOutcome(ctx, ec, e, thePolicy, theRule)
	ruleNode.Attach(node)
	ruleNode.SetResult(box.Object(d))
	ruleNode.SetErr(err)
	if err != nil {
		return d, nil, ruleNode, err
	}

	// Compute attachment values if exported
	attachments := map[string]box.Value{}
	if ex, ok := thePolicy.RuleExports[rule]; ok {
		for _, attachment := range ex.Attachments {
			ctx, attachmentNode, done := trace.New(ctx, attachment.Value, "attachment", map[string]any{
				"name": attachment.Name,
			})
			defer done()

			v, node, err := eval(ctx, ec, e, thePolicy, attachment.Value)
			attachmentNode.Attach(node)
			if err != nil {
				attachmentNode.SetErr(err)
				return d, attachments, ruleNode, err
			}
			attachments[attachment.Name] = v
			attachmentNode.SetResult(v)
			ruleNode.Attach(attachmentNode)
			continue
		}
	}

	return d, attachments, ruleNode, nil
}

// loadJSInstance builds a fresh VM and loads the module into it. The instance
// carries the module's whole exports object; restriction to the names a `use`
// asked for happens per binding (see ModuleBinding.lookupExport).
func (e *executorImpl) loadJSInstance(ctx context.Context, ms *js.ModuleSpec, start time.Time) (*JSInstance, error) {
	// Fresh VM with its own require cache; the compiled program is reused.
	ar := js.NewAliasRuntime(e.jsRegistry, ms.Dir)
	// Pin time before any module code runs, so top-level code and calls agree.
	js.SetExecutionStart(ar.VM, start)
	if err := ar.SetupStdLib(ctx, e.index.Pack); err != nil {
		return nil, err
	}

	// Load the top-level module (exports object)
	exObj, err := ar.Require(ctx, ms.Dir, ms.KeyOrPath())
	if err != nil {
		return nil, err
	}
	// Capturing the export keys may run a Proxy ownKeys trap, so it runs under
	// the deadline too.
	var instance *JSInstance
	if err := runGuardedJS(ctx, ar.VM, func() error {
		instance = newJSInstance(ar.VM, exObj)
		return nil
	}); err != nil {
		return nil, err
	}
	return instance, nil
}

func (e *executorImpl) bindUses(ctx context.Context, ec *ExecutionContext, p *index.Policy) error {
	fileDir, err := filepath.Abs(filepath.Dir(p.FilePath))
	if err != nil {
		return err
	}

	for alias, use := range p.Uses {
		ms, err := e.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, fileDir)
		if err != nil {
			return err
		}

		// Resolve and ensure program exists
		binding, err := e.getModuleBinding(ctx, use, ms)
		if err != nil {
			return err
		}

		ec.BindModule(alias, binding)
	}
	return nil
}

// getModuleBinding returns the binding for one `use` statement. The loaded
// module is shared by every `use` of the same path; the export
// projection and the "missing required export" check belong to this `use` alone.
func (e *executorImpl) getModuleBinding(ctx context.Context, use *ast.UseStatement, ms *js.ModuleSpec) (*ModuleBinding, error) {
	module, err := e.getModule(ctx, ms)
	if err != nil {
		return nil, err
	}
	if err := module.checkExports(ctx, use.Modules); err != nil {
		return nil, err
	}
	return &ModuleBinding{
		CanonicalKey: module.key,
		Alias:        use.As,
		exports:      slices.Clone(use.Modules),
		module:       module,
	}, nil
}

// getModule returns the shared module for ms, keyed by path alone. The first
// caller loads it (compiling the program once and running it in a probe VM);
// concurrent callers wait for that load, and later callers reuse the result for
// the executor's lifetime. Calls then run in fresh VMs (see jsModule).
func (e *executorImpl) getModule(ctx context.Context, ms *js.ModuleSpec) (*jsModule, error) {
	key := ms.KeyOrPath()
	for {
		e.modulesMu.Lock()
		if e.modulesClosed {
			e.modulesMu.Unlock()
			return nil, ErrExecutorClosed
		}
		if e.modules == nil {
			e.modules = map[string]*moduleLoad{}
		}
		l, ok := e.modules[key]
		if !ok {
			l = &moduleLoad{done: make(chan struct{})}
			e.modules[key] = l
			e.modulesMu.Unlock()
			return e.loadModule(ctx, key, ms, l)
		}
		e.modulesMu.Unlock()

		select {
		case <-l.done:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		// A load that failed only because its own caller's context ended says
		// nothing about this caller; try again with this caller's context.
		if l.err != nil && ctx.Err() == nil && isContextError(l.err) {
			continue
		}
		return l.module, l.err
	}
}

// loadModule performs the single load for l and publishes the result. A failed
// load is not cached, so a later evaluation retries it.
func (e *executorImpl) loadModule(ctx context.Context, key string, ms *js.ModuleSpec, l *moduleLoad) (*jsModule, error) {
	load := func(ctx context.Context, start time.Time) (*JSInstance, error) {
		return e.loadJSInstance(ctx, ms, start)
	}
	e.moduleLoads.Add(1)
	module, err := newJSModule(ctx, key, load)

	e.modulesMu.Lock()
	if err != nil {
		if e.modules[key] == l {
			delete(e.modules, key)
		}
	} else if e.modulesClosed {
		module.close()
		module, err = nil, ErrExecutorClosed
	}
	l.module, l.err = module, err
	close(l.done)
	e.modulesMu.Unlock()
	return module, err
}

func isContextError(err error) bool {
	return stdErr.Is(err, context.Canceled) || stdErr.Is(err, context.DeadlineExceeded)
}

// evaluateRuleOutcome drives rule evaluation and returns (value, node, error).
func evaluateRuleOutcome(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, r *index.Rule) (*Decision, *trace.Node, error) {
	ctx, rn, done := trace.New(ctx, r.Node, "rule", map[string]any{
		"name": r.Name,
	})
	defer done()

	// `when` gate: (`when` is `true` by default)
	whenVal := trinary.True

	// evaluate the when gate
	if r.When != nil {
		ctx, wn, done := trace.New(ctx, r.When, "rule-when", map[string]any{})
		defer done()

		cond, condNode, err := eval(ctx, ec, e, p, r.When)
		wn.Attach(condNode)
		if err != nil {
			wn.SetErr(err)
			return nil, rn, err
		}
		whenVal = box.TrinaryFrom(cond)
		rn.Attach(wn)
	}

	if !whenVal.IsTrue() {
		// the default response is NA
		theDefault := DecisionOf(box.Trinary(trinary.Unknown))

		// we have a default expression
		if r.Default != nil {
			ctx, dn, done := trace.New(ctx, r.Default, "rule-default", map[string]any{})
			defer done()

			// evaluate the default expression
			val, defNode, err := eval(ctx, ec, e, p, r.Default)
			dn.Attach(defNode).SetResult(val).SetErr(err)

			theDefault = DecisionOf(val)
			rn.Attach(dn)
		}
		return theDefault, rn, nil
	}

	ctx, rb, done := trace.New(ctx, r.Body, "rule-body", map[string]any{})
	defer done()

	val, bodyNode, err := eval(ctx, ec, e, p, r.Body)
	rb.Attach(bodyNode).SetResult(val).SetErr(err)
	rn.Attach(rb)

	// Coerce to a *Decision using tristate.From(val)
	return DecisionOf(val), rn, err
}
