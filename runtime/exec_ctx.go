// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/xerr"
)

var ErrIllegalFactInjection = fmt.Errorf("fact injection not allowed in child context")

type injectedFact struct {
	value     box.Value
	typeRef   ast.TypeRef
	isDefault bool
}

// ExecutionContext holds ephemeral state for a single rule execution.
// It owns an arena and is disposed immediately after the run.
type ExecutionContext struct {
	rwmu sync.RWMutex

	policy *index.Policy

	createdAt time.Time

	parent *ExecutionContext

	refStack []string // reference stack for cycle detection

	// callDepth counts the lambda and derive invocations enclosing this
	// context (see MaxCallDepth). Child contexts inherit it; an invocation
	// sets its child's depth to the call site's plus one. It is written only
	// before the context is shared, so it needs no lock.
	callDepth int

	// evalDepth counts the eval frames active in this evaluation (see
	// MaxEvalDepth). One counter is shared by every context of a root
	// evaluation, imported decisions included; a lambda invocation uses its
	// call site's counter.
	evalDepth *evalDepthCounter

	facts map[string]injectedFact        // injected via WITH
	lets  map[string]*ast.VarDeclaration // policy-scoped lets

	locals map[string]box.Value // evaluated local values

	modules map[string]*ModuleBinding // alias -> module binding (for `use`)

	executor Executor

	// evalDerive marks evaluation inside a derive body: blocks fact reads, rule dispatch by
	// identifier, TypeScript module calls, and impure builtins (see getTarget / evalIdent).
	// Set only on the detached child EC during invokeDerive; not guarded by rwmu because
	// derive evaluation is single-goroutine per caller EC (no concurrent derive on one EC).
	evalDerive *index.Derive
}

func (ec *ExecutionContext) IsLetInjected(name string) bool {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	_, ok := ec.lets[name]
	return ok
}

func NewExecutionContext(policy *index.Policy, executor Executor) *ExecutionContext {
	return &ExecutionContext{
		parent:    nil,
		createdAt: time.Now(),
		policy:    policy,
		evalDepth: &evalDepthCounter{},
		refStack:  make([]string, 0), // reference stack
		facts:     make(map[string]injectedFact),
		locals:    make(map[string]box.Value),
		lets:      make(map[string]*ast.VarDeclaration),
		modules:   make(map[string]*ModuleBinding),
		executor:  executor,
	}
}

// DetachedChildContext returns an isolated execution context for derive evaluation:
// no parent bubble, no facts/modules/lets (derive bodies use locals and DefineShort/FQN only),
// fresh locals map, same policy and executor, refStack clone (derive→derive cycle detection
// across the detached boundary), and inherited createdAt (so now() stays policy-stable).
func (ec *ExecutionContext) DetachedChildContext() *ExecutionContext {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()

	return &ExecutionContext{
		policy:    ec.policy,
		executor:  ec.executor,
		createdAt: ec.createdAt,
		refStack:  slices.Clone(ec.refStack),
		callDepth: ec.callDepth,
		evalDepth: ec.evalDepth,
		locals:    make(map[string]box.Value),
	}
}

// Dispose frees the arena immediately. Do NOT reuse an EC after Dispose.
func (ec *ExecutionContext) Dispose() {}

// AttachedChildContext creates a child context. All lookups will be
// performed in the child context first, then the parent context.
func (ec *ExecutionContext) AttachedChildContext() *ExecutionContext {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()

	return &ExecutionContext{
		parent:     ec,
		createdAt:  ec.createdAt,
		refStack:   slices.Clone(ec.refStack),            // inherit the call stack from the parent
		callDepth:  ec.callDepth,                         // inherit the invocation depth from the parent
		evalDepth:  ec.evalDepth,                         // share the evaluation's frame counter
		policy:     ec.policy,                            // inherit the policy from the parent
		modules:    ec.modules,                           // inherit the module bindings from the parent
		executor:   ec.executor,                          // inherit the executor from the parent
		evalDerive: ec.evalDerive,                        // inherit derive evaluation mode (evalBlock uses AttachedChildContext)
		facts:      nil,                                  // a child context should not have facts at all
		locals:     make(map[string]box.Value),           // local values
		lets:       make(map[string]*ast.VarDeclaration), // local let declarations
	}
}

func (ec *ExecutionContext) CreatedAt() time.Time {
	if ec.parent != nil {
		return ec.parent.CreatedAt()
	}
	return ec.createdAt
}

// Inject facts into the current context.
// It is illegal to inject facts into a child context.
func (ec *ExecutionContext) InjectFact(ctx context.Context, name string, v box.Value, isDefault bool, typeRef ast.TypeRef) error {
	ec.rwmu.Lock()
	defer ec.rwmu.Unlock()

	if ec.parent != nil {
		return fmt.Errorf("%s: %w", name, ErrIllegalFactInjection)
	}

	ec.facts[name] = injectedFact{
		// an imported decision bound to a bool or trinary fact, whether
		// injected or a default, holds its state (#136)
		value:     bindTruthValue(v, typeRef),
		isDefault: isDefault,
		typeRef:   typeRef,
	}
	return nil
}

func (ec *ExecutionContext) IsFactInjected(name string) bool {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	_, ok := ec.facts[name]
	return ok
}

// Inject local let declarations into the current context.
// Let declarations are always injected into the current context - NEVER in the parent.
func (ec *ExecutionContext) InjectLet(name string, v *ast.VarDeclaration) error {
	ec.rwmu.Lock()
	defer ec.rwmu.Unlock()

	if other, ok := ec.lets[name]; ok {
		return xerr.ErrConflict("let declaration", v.Span(), other.Span())
	}
	ec.lets[name] = v
	return nil
}

// SetLocal sets a local value in the current context if and only if the current context supplied an identifier
// with that name.
func (ec *ExecutionContext) SetLocal(name string, value box.Value, force bool) {
	// Take a read lock first to check if we need to actually write
	ec.rwmu.RLock()

	// Force always upgrades to write lock and writes
	if force {
		ec.rwmu.RUnlock()
		ec.rwmu.Lock()
		defer ec.rwmu.Unlock()
		ec.locals[name] = value
		return
	}

	// Check presence for fact, let, and rule
	_, hasFact := ec.facts[name]
	_, hasLet := ec.lets[name]
	_, hasRule := ec.policy.Rules[name]
	parent := ec.parent
	ec.rwmu.RUnlock()

	if hasFact || hasLet || hasRule {
		ec.rwmu.Lock()
		defer ec.rwmu.Unlock()
		ec.locals[name] = value
		return
	}

	// Otherwise, forward to parent if one exists
	if parent != nil {
		parent.SetLocal(name, value, false)
	}
}

// GetLocal gets a local value from the current context if present - otherwise the parent context is checked.
func (ec *ExecutionContext) GetLocal(name string) (box.Value, bool) {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	v, ok := ec.locals[name]
	if !ok && ec.parent != nil {
		// if we have a parent, we need to get the local from the parent
		return ec.parent.GetLocal(name)
	}
	return v, ok
}

func (ec *ExecutionContext) GetFact(name string) (box.Value, bool) {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	if ec.parent != nil {
		// if we have a parent, we need to get the fact from the parent
		return ec.parent.GetFact(name)
	}
	v, ok := ec.facts[name]
	if !ok {
		return box.Undefined(), false
	}
	return v.value, ok
}

func (ec *ExecutionContext) GetLet(name string) (*ast.VarDeclaration, bool) {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	v, ok := ec.lets[name]
	if !ok && ec.parent != nil {
		// if we have a parent, we need to get the let from the parent
		return ec.parent.GetLet(name)
	}
	return v, ok
}

func (ec *ExecutionContext) BindModule(alias string, m *ModuleBinding) {
	ec.rwmu.Lock()
	defer ec.rwmu.Unlock()
	ec.modules[alias] = m
}

func (ec *ExecutionContext) Module(alias string) (binding *ModuleBinding, found bool) {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	m, ok := ec.modules[alias]
	return m, ok
}

// PushRefStack adds an item to the reference stack for cycle detection
func (ec *ExecutionContext) PushRefStack(uniqueID string) error {
	ec.rwmu.Lock()
	defer ec.rwmu.Unlock()

	// Check if this rule is already in the stack (cycle detection)
	if slices.Contains(ec.refStack, uniqueID) {
		return fmt.Errorf("'%s' references itself: %w", uniqueID, xerr.ErrInfiniteRecursion(append(ec.refStack, uniqueID)))
	}

	ec.refStack = append(ec.refStack, uniqueID)
	return nil
}

// PopRefStack removes the last item from the call stack
func (ec *ExecutionContext) PopRefStack() {
	ec.rwmu.Lock()
	defer ec.rwmu.Unlock()

	if len(ec.refStack) > 0 {
		ec.refStack = ec.refStack[:len(ec.refStack)-1]
	}
}

// GetCallStack returns a copy of the current reference stack
func (ec *ExecutionContext) GetRefStack() []string {
	ec.rwmu.RLock()
	defer ec.rwmu.RUnlock()
	return slices.Clone(ec.refStack)
}

// MaxCallDepth bounds nested lambda and derive invocations in one evaluation.
// A let-bound lambda can call itself (#161), directly or by being passed to
// itself, and neither form can be ruled out statically, so unbounded recursion
// is stopped here with an ordinary error instead of a fatal Go stack overflow
// that would take a server down. MaxCallDepth alone does not bound the Go
// stack, because one level costs stack in proportion to how deeply the
// recursive call is nested in the body; MaxEvalDepth bounds that.
const MaxCallDepth = 1000

// MaxEvalDepth bounds the eval frames active at once in one evaluation, which
// is what Go stack use is proportional to. A runaway recursion stopped at
// this limit used under 128 MB of Go stack (64 MB was not enough), with and
// without -race, whether its body was trivial or nested 2000 terms deep:
// eight times below Go's default 1 GB maximum stack. A trivial
// self-recursive lambda uses about a dozen frames per level, so MaxCallDepth
// trips first for it. A non-recursive expression hits the limit only when
// nested about 25,000 levels deep.
const MaxEvalDepth = 50000

// ErrEvalDepthExceeded is wrapped by the error eval returns when an
// evaluation would exceed MaxEvalDepth. Like ErrCallDepthExceeded it is
// passed through evalCall unwrapped.
var ErrEvalDepthExceeded = errors.New("evaluation depth exceeded")

// evalDepthCounter is one evaluation's count of active eval frames. An
// evaluation runs on one goroutine; the counter is atomic only so that a
// callable that escaped into another evaluation cannot race on it.
type evalDepthCounter struct {
	n atomic.Int64
}

func (c *evalDepthCounter) enter() error {
	if c.n.Add(1) > MaxEvalDepth {
		c.n.Add(-1)
		return fmt.Errorf("%w (limit %d nested evaluations)", ErrEvalDepthExceeded, MaxEvalDepth)
	}
	return nil
}

func (c *evalDepthCounter) leave() { c.n.Add(-1) }

// inheritDepth carries a calling evaluation's depth budget into ec, a fresh
// root context for an imported decision, so a recursion through imports is
// bounded by the same limits.
func (ec *ExecutionContext) inheritDepth(caller *ExecutionContext) {
	if caller == nil {
		return
	}
	ec.callDepth = caller.callDepth
	if caller.evalDepth != nil {
		ec.evalDepth = caller.evalDepth
	}
}

// ErrCallDepthExceeded is wrapped by the error an invocation returns when it
// would exceed MaxCallDepth. evalCall passes it through without adding a
// "failed to call function" layer per level.
var ErrCallDepthExceeded = errors.New("call depth exceeded")

// enterCall returns the depth of an invocation made from the context site,
// or an error when it would exceed MaxCallDepth. kind names the callee in the
// message ("lambda" or "derive").
func enterCall(site *ExecutionContext, kind string) (int, error) {
	depth := site.callDepth + 1
	if depth > MaxCallDepth {
		return 0, fmt.Errorf("%s %w (limit %d)", kind, ErrCallDepthExceeded, MaxCallDepth)
	}
	return depth, nil
}
