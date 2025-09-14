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
	"path/filepath"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/perch"
	"github.com/binaek/sentra/runtime/js"
	"github.com/binaek/sentra/runtime/trace"
	"github.com/binaek/sentra/trinary"
	"github.com/binaek/sentra/xerr"
	"github.com/dop251/goja"
	"github.com/jackc/puddle/v2"
)

type NewExecutorOption func(*executorImpl)

// The number of Megabytes to allocate for the call memoize cache
func WithCallMemoizeCacheSize(size int) NewExecutorOption {
	return func(e *executorImpl) {
		e.callMemoizePerch = perch.New[any](size << 20 /* size in megabytes */)
	}
}

type Executor interface {
	ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) (*Decision, DecisionAttachments, *trace.Node, error)
	ExecRule(ctx context.Context, namespace, policy, rule string, facts map[string]any) (*Decision, DecisionAttachments, *trace.Node, error)
	Index() *index.Index
}

// executorImpl ties together the index, JS loader, and evaluation.
type executorImpl struct {
	index            *index.Index
	jsRegistry       *js.Registry
	callMemoizePerch *perch.Perch[any]
}

// NewExecutor builds an Executor with built-in @sentra/* modules registered.
func NewExecutor(idx *index.Index, opts ...NewExecutorOption) Executor {
	exec := &executorImpl{
		index:            idx,
		jsRegistry:       js.NewRegistry(idx.Pack.Location),
		callMemoizePerch: perch.New[any](10 << 20 /* 10 MB */),
	}

	exec.jsRegistry.RegisterGoBuiltin("uuid", js.BuiltinUuidGo)
	exec.jsRegistry.RegisterGoBuiltin("crypto", js.BuiltinCryptoGo)
	exec.jsRegistry.RegisterGoBuiltin("base64", js.BuiltinBase64Go)

	for _, opt := range opts {
		opt(exec)
	}

	return exec
}

func (e *executorImpl) Index() *index.Index {
	return e.index
}

// ExecPolicy uses policy's `outcome` rule; returns (value, attachments, tree) as a RuleOutcome.
func (e *executorImpl) ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) (*Decision, DecisionAttachments, *trace.Node, error) {
	return e.ExecRule(ctx, namespace, policy, "outcome", facts)
}

// ExecRule executes an exported rule and returns (value, attachments, tree) as a RuleOutcome.
func (e *executorImpl) ExecRule(ctx context.Context, namespace, policy, rule string, facts map[string]any) (*Decision, DecisionAttachments, *trace.Node, error) {
	// Validate exported
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := p.VerifyRuleExported(rule); err != nil {
		return nil, nil, nil, err
	}

	ec := NewExecutionContext(p)
	defer ec.Dispose()

	for factName, factStatement := range p.Facts {
		// look for a value for this fact in the passed in facts map
		if _, ok := facts[factName]; ok {
			if err := ec.InjectFact(ctx, factName, facts[factName], factStatement.Type); err != nil {
				return nil, nil, nil, err
			}
			continue // move on to the next fact
		}

		// no value supplied for this fact, and if the fact has no default value, we error.
		// this is an invalid invocation.
		if factStatement.Default == nil {
			return nil, nil, nil, xerr.ErrInvalidInvocation(fmt.Sprintf("fact %q has no default value", factName))
		}

		// evaluate the default value, this will be injected into the context
		val, _, err := eval(ctx, ec, e, p, factStatement.Default)
		if err != nil {
			return nil, nil, nil, err
		}

		// inject the default value
		if err := ec.InjectFact(ctx, factStatement.Name, val, factStatement.Type); err != nil {
			return nil, nil, nil, err
		}
	}

	// bind lets
	for k, v := range p.Lets {
		ec.InjectLet(k, v)
	}

	// Bind `use` modules
	if err := e.bindUses(ctx, ec, p); err != nil {
		return nil, nil, nil, err
	}

	decision, attachments, ruleNode, err := e.execRule(ctx, ec, namespace, policy, rule)
	return decision, attachments, ruleNode, err
}

func (e *executorImpl) execRule(ctx context.Context, ec *ExecutionContext, namespace, policy, rule string) (*Decision, DecisionAttachments, *trace.Node, error) {
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, nil, nil, err
	}

	r, ok := p.Rules[rule]
	if !ok {
		return nil, nil, nil, xerr.ErrRuleNotFound(index.RuleFQN(namespace, policy, rule))
	}

	// Wrap rule evaluation in a decision node
	ruleNode, done := trace.New("rule-outcome", rule, r, map[string]any{
		"namespace": namespace,
		"policy":    policy,
	})
	defer done()

	d, node, err := evaluateRuleOutcome(ctx, ec, e, p, r)
	ruleNode.Attach(node)
	ruleNode.SetResult(d)
	ruleNode.SetErr(err)
	if err != nil {
		return nil, nil, nil, err
	}

	// Compute attachment values if exported
	attachments := map[string]any{}
	if ex, ok := p.RuleExports[rule]; ok {
		for _, attachment := range ex.Attachments {
			attachmentNode, done := trace.New("attachment", attachment.Name, nil, map[string]any{
				"name":  attachment.Name,
				"alias": attachment.Alias,
			})
			defer done()

			if attachment.Name == "aUnknown" {
				str := "aUnknown"
				_ = str
			}

			v, node, err := evalIdent(ctx, ec, e, p, attachment.Name)
			attachmentNode.Attach(node)
			if err != nil {
				return nil, nil, nil, err
			}
			attachments[attachment.Name] = v
			ruleNode.Attach(attachmentNode)
			continue
		}
	}

	return d, attachments, ruleNode, nil
}

func (e *executorImpl) jsBindingConstructor(ctx context.Context, use *ast.UseStatement, ms *js.ModuleSpec) (*JSInstance, error) {
	// Per-alias VM with require cache
	ar := js.NewAliasRuntime(e.jsRegistry, ms.Dir)

	// Load the top-level module (exports object)
	exObj, err := ar.Require(ctx, ms.Dir, ms.KeyOrPath())
	if err != nil {
		return nil, err
	}

	// Build exports map restricted to requested idents
	exports := map[string]goja.Value{}
	if len(use.Modules) > 0 {
		for _, name := range use.Modules {
			if v := exObj.Get(name); v != nil {
				exports[name] = v
			} else {
				// strict: fail fast if requested fn missing
				return nil, fmt.Errorf("module %s missing required export %q", ms.KeyOrPath(), name)
			}
		}
	} else {
		for _, k := range exObj.Keys() {
			exports[k] = exObj.Get(k)
		}
	}
	return &JSInstance{
		VM:      ar.VM,
		Exports: exports,
	}, nil
}

func (e *executorImpl) bindUses(ctx context.Context, ec *ExecutionContext, p *index.Policy) error {
	fileDir, err := filepath.Abs(filepath.Dir(p.FilePath))
	if err != nil {
		return err
	}
	for _, use := range p.Uses {
		// Resolve and ensure program exists
		ms, err := e.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, fileDir)
		if err != nil {
			return err
		}

		vmPool, err := puddle.NewPool(&puddle.Config[*JSInstance]{
			Constructor: func(ctx context.Context) (*JSInstance, error) {
				return e.jsBindingConstructor(ctx, use, ms)
			},
			Destructor: func(res *JSInstance) {
				res.VM.ClearInterrupt()
			},
			MaxSize: 10,
		})

		if err != nil {
			return err
		}

		// warm up the pool - this will create a couple of VMs - and also verify that we can actually acquire them
		if err := vmPool.CreateResource(ctx); err != nil {
			return err
		}

		ec.BindModule(use.As, ModuleBinding{
			Alias:  use.As,
			VMPool: vmPool,
		})
	}
	return nil
}

// evaluateRuleOutcome drives rule evaluation and returns (value, node, error).
func evaluateRuleOutcome(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, r *index.Rule) (*Decision, *trace.Node, error) {
	rn, done := trace.New("rule", r.Name, r, map[string]any{})
	defer done()

	// `when` gate: (`when` is `true` by default)
	whenVal := trinary.False

	// evaluate the when gate
	if r.When != nil {
		wn, done := trace.New("rule-when", r.Name, r.When, map[string]any{})
		defer done()

		cond, condNode, err := eval(ctx, ec, e, p, r.When)
		wn.Attach(condNode)
		if err != nil {
			wn.SetErr(err)
			return nil, rn, err
		}
		whenVal = trinary.From(cond)
		rn.Attach(wn)
	}

	if !whenVal.IsTrue() {
		// the default response is NA
		theDefault := DecisionOf(trinary.Unknown)

		// we have a default expression
		if r.Default != nil {
			dn, done := trace.New("rule-default", r.Name, r.Default, map[string]any{})
			defer done()

			// evaluate the default expression
			val, defNode, err := eval(ctx, ec, e, p, r.Default)
			dn.Attach(defNode).SetResult(val).SetErr(err)

			theDefault = DecisionOf(val)
			rn.Attach(dn)
		}
		return theDefault, rn, nil
	}

	rb, done := trace.New("rule-body", r.Name, r, map[string]any{})
	defer done()

	val, bodyNode, err := eval(ctx, ec, e, p, r.Body)
	rb.Attach(bodyNode).SetResult(val).SetErr(err)
	rn.Attach(rb)

	// Otherwise, coerce to a *Decision using tristate.From(val)
	return DecisionOf(val), rn, err
}
