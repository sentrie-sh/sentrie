// SPDX-License-Identifier: Apache-2.0

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
	stdErr "errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/binaek/perch"
	"github.com/dop251/goja"
	"github.com/jackc/puddle/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/js"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

type NewExecutorOption func(*executorImpl)

// The number of Megabytes to allocate for the call memoize cache
func WithCallMemoizeCacheSize(size int) NewExecutorOption {
	return func(e *executorImpl) {
		e.callMemoizePerch = perch.New[any](size << 20 /* size in megabytes */)
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
}

// executorImpl ties together the index, JS loader, and evaluation.
type executorImpl struct {
	index              *index.Index
	jsRegistry         *js.Registry
	moduleBindingPerch *perch.Perch[*ModuleBinding] // --> (policy.useAlias) -> module binding
	callMemoizePerch   *perch.Perch[any]
}

// NewExecutor builds an Executor with built-in @sentra/* modules registered.
func NewExecutor(idx *index.Index, opts ...NewExecutorOption) (Executor, error) {
	exec := &executorImpl{
		index:              idx,
		jsRegistry:         js.NewRegistry(idx.Pack.Location),
		moduleBindingPerch: perch.New[*ModuleBinding](100 << 20 /* 100 MB */), // --> (policy.useAlias) -> module binding
		callMemoizePerch:   perch.New[any](10 << 20 /* 10 MB */),
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

	// Reserve the cache slots
	exec.moduleBindingPerch.Reserve()

	exec.callMemoizePerch.Reserve()

	return exec, nil
}

func (e *executorImpl) Index() *index.Index {
	return e.index
}

// ExecPolicy executes all exported rules and returns the results
func (e *executorImpl) ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) ([]*ExecutorOutput, error) {
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, err
	}

	theLock := &sync.Mutex{}
	var compositeErr error
	outputs := make([]*ExecutorOutput, 0, len(p.RuleExports))
	wg := &sync.WaitGroup{}
	for _, ruleExport := range p.RuleExports {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					compositeErr = stdErr.New("panic in ExecRule: " + fmt.Sprintf("%v", r))
				}
			}()

			output, err := e.ExecRule(ctx, namespace, policy, ruleExport.RuleName, facts)

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
func (e *executorImpl) ExecRule(ctx context.Context, namespace, policy, rule string, injectedFacts map[string]any) (*ExecutorOutput, error) {
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
			// Facts are always non-nullable - validate value is not null
			if factValue == nil {
				return nil, errors.Wrapf(xerr.ErrInvalidInvocation(""), "fact '%s' cannot be null", factName)
			}
			err := ec.InjectFact(ctx, factName, factValue, false, factStatement.Type)
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
				return nil, errors.Wrap(xerr.ErrUnresolvableFact(factName), err.Error())
			}

			// Facts are always non-nullable - validate default value is not null
			if val == nil {
				return nil, errors.Wrapf(xerr.ErrInvalidInvocation(""), "fact '%s' cannot have null default value", factName)
			}

			// inject the default value
			if err := ec.InjectFact(ctx, factStatement.Name, val, true, factStatement.Type); err != nil {
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
		decision = DecisionOf(trinary.Unknown)
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
	for name, value := range ec.facts {
		if value.typeRef == nil {
			// if there's no shape indication, we skip validation
			continue
		}
		stmt := thePolicy.Facts[name]
		// validate the value against the type
		if err := validateValueAgainstTypeRef(ctx, ec, e, thePolicy, value.value, value.typeRef, stmt.Span()); err != nil {
			return nil, nil, nil, err
		}
	}

	d, node, err := evaluateRuleOutcome(ctx, ec, e, thePolicy, theRule)
	ruleNode.Attach(node)
	ruleNode.SetResult(d)
	ruleNode.SetErr(err)
	if err != nil {
		return d, nil, ruleNode, err
	}

	// Compute attachment values if exported
	attachments := map[string]any{}
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

func (e *executorImpl) jsBindingConstructor(ctx context.Context, use *ast.UseStatement, ms *js.ModuleSpec) (*JSInstance, error) {
	// Per-alias VM with require cache
	ar := js.NewAliasRuntime(e.jsRegistry, ms.Dir)
	if err := ar.SetupStdLib(ctx, e.index.Pack); err != nil {
		return nil, err
	}

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
		rt:      ar.VM,
		exports: exports,
	}, nil
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
		binding, _, err := e.getModuleBinding(ctx, use, ms)
		if err != nil {
			return err
		}

		ec.BindModule(alias, binding)
	}
	return nil
}

// getModuleBinding resolves and caches a module binding for a given use statement and module spec.
func (e *executorImpl) getModuleBinding(ctx context.Context, use *ast.UseStatement, ms *js.ModuleSpec) (binding *ModuleBinding, _ bool, err error) {
	constructor := func(ctx context.Context) (*JSInstance, error) {
		return e.jsBindingConstructor(ctx, use, ms)
	}

	destructor := func(res *JSInstance) {
		// clear the interrupt
		res.rt.ClearInterrupt()
	}

	perchLoader := func(ctx context.Context, _ string) (*ModuleBinding, error) {
		jsInstancePool, err := puddle.NewPool(&puddle.Config[*JSInstance]{
			Constructor: constructor,
			Destructor:  destructor,
			MaxSize:     10,
		})
		if err != nil {
			return nil, err
		}
		// warm up the pool - this will create a couple of VMs - and also verify that we can actually acquire them
		if err := jsInstancePool.CreateResource(ctx); err != nil {
			return nil, err
		}
		return &ModuleBinding{
			CanonicalKey: ms.KeyOrPath(),
			Alias:        use.As,
			instancePool: jsInstancePool,
		}, nil
	}

	return e.moduleBindingPerch.Get(ctx, ms.KeyOrPath(), -1, perchLoader)
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
		whenVal = trinary.From(cond)
		rn.Attach(wn)
	}

	if !whenVal.IsTrue() {
		// the default response is NA
		theDefault := DecisionOf(trinary.Unknown)

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
