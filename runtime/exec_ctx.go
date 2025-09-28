// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/pkg/errors"
)

var ErrIllegalFactInjection = fmt.Errorf("fact injection not allowed in child context")

type injectedFact struct {
	value     any
	typeRef   ast.TypeRef
	isDefault bool
}

// ExecutionContext holds ephemeral state for a single rule execution.
// It owns an arena and is disposed immediately after the run.
type ExecutionContext struct {
	mu sync.Mutex

	policy *index.Policy

	parent *ExecutionContext

	facts map[string]injectedFact        // injected via WITH
	lets  map[string]*ast.VarDeclaration // policy-scoped lets

	locals map[string]any // evaluated local values

	modules map[string]*ModuleBinding // alias -> module binding (for `use`)
}

func NewExecutionContext(policy *index.Policy) *ExecutionContext {
	return &ExecutionContext{
		parent:  nil,
		policy:  policy,
		facts:   make(map[string]injectedFact),
		locals:  make(map[string]any),
		lets:    make(map[string]*ast.VarDeclaration),
		modules: make(map[string]*ModuleBinding),
	}
}

// Dispose frees the arena immediately. Do NOT reuse an EC after Dispose.
func (ec *ExecutionContext) Dispose() {}

// AttachedChildContext creates a child context. All lookups will be
// performed in the child context first, then the parent context.
func (ec *ExecutionContext) AttachedChildContext() *ExecutionContext {
	return &ExecutionContext{
		parent:  ec,
		policy:  ec.policy, // inherit the policy from the parent
		facts:   nil,       // a child context should not have facts at all
		locals:  make(map[string]any),
		lets:    make(map[string]*ast.VarDeclaration),
		modules: ec.modules, // inherit the module bindings from the parent
	}
}

// Inject facts into the current context.
// It is illegal to inject facts into a child context.
func (ec *ExecutionContext) InjectFact(ctx context.Context, name string, v any, isDefault bool, typeRef ast.TypeRef) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.parent != nil {
		return errors.Wrap(ErrIllegalFactInjection, name)
	}

	ec.facts[name] = injectedFact{
		value:     v,
		isDefault: isDefault,
		typeRef:   typeRef,
	}
	return nil
}

func (ec *ExecutionContext) IsFactInjected(name string) bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	_, ok := ec.facts[name]
	return ok
}

// Inject local let declarations into the current context.
// Let declarations are always injected into the current context - NEVER in the parent.
func (ec *ExecutionContext) InjectLet(name string, v *ast.VarDeclaration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.lets[name] = v
}

// SetLocal sets a local value in the current context if and only if the current context supplied an identifier
// with that name.
func (ec *ExecutionContext) SetLocal(name string, value any, force bool) {
	if force {
		ec.mu.Lock()
		defer ec.mu.Unlock()
		ec.locals[name] = value
		return
	}

	// Only set if we have a fact, let, or rule with this name in the current context
	if _, ok := ec.GetFact(name); ok {
		ec.locals[name] = value
		return
	}

	if _, ok := ec.GetLet(name); ok {
		ec.locals[name] = value
		return
	}

	if _, ok := ec.policy.Rules[name]; ok {
		ec.mu.Lock()
		defer ec.mu.Unlock()
		ec.locals[name] = value
		return
	}

	if ec.parent != nil {
		ec.parent.SetLocal(name, value, false)
	}
}

// GetLocal gets a local value from the current context if present - otherwise the parent context is checked.
func (ec *ExecutionContext) GetLocal(name string) (any, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	v, ok := ec.locals[name]
	if !ok && ec.parent != nil {
		// if we have a parent, we need to get the local from the parent
		return ec.parent.GetLocal(name)
	}
	return v, ok
}

func (ec *ExecutionContext) GetFact(name string) (any, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if ec.parent != nil {
		// if we have a parent, we need to get the fact from the parent
		return ec.parent.GetFact(name)
	}
	v, ok := ec.facts[name]
	if !ok {
		return Undefined, false
	}
	return v.value, ok
}

func (ec *ExecutionContext) GetLet(name string) (*ast.VarDeclaration, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	v, ok := ec.lets[name]
	if !ok && ec.parent != nil {
		// if we have a parent, we need to get the let from the parent
		return ec.parent.GetLet(name)
	}
	return v, ok
}

func (ec *ExecutionContext) BindModule(alias string, m *ModuleBinding) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.modules[alias] = m
}

func (ec *ExecutionContext) Module(alias string) (*ModuleBinding, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	m, ok := ec.modules[alias]
	return m, ok
}
