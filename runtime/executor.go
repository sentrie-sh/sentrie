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
	"strings"
	"sync"
	"time"

	"github.com/binaek/perch"
	"github.com/dop251/goja"
	"github.com/jackc/puddle/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	otelconfig "github.com/sentrie-sh/sentrie/otel"
	"github.com/sentrie-sh/sentrie/runtime/js"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// ExecutionMetrics holds metrics for runtime execution
type ExecutionMetrics struct {
	JSCallCount    metric.Int64Counter
	JSCallDuration metric.Float64Histogram
	JSCallErrors   metric.Int64Counter
	JSPoolGuages   *perch.Perch[metric.Int64UpDownCounter] // pool name -> gauge
}

type NewExecutorOption func(*executorImpl)

// The number of Megabytes to allocate for the call memoize cache
func WithCallMemoizeCacheSize(size int) NewExecutorOption {
	return func(e *executorImpl) {
		e.callMemoizePerch = perch.New[any](size << 20 /* size in megabytes */)
	}
}

// WithOTelConfig sets the OpenTelemetry configuration for the executor
func WithOTelConfig(config *otelconfig.OTelConfig) NewExecutorOption {
	return func(e *executorImpl) {
		e.otelConfig = config
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
	// Tracer returns the tracer for the executor
	Tracer() oteltrace.Tracer
	// Meter returns the meter for the executor
	Meter() metric.Meter
	// OTelConfig returns the OpenTelemetry configuration for the executor
	OTelConfig() *otelconfig.OTelConfig
	// Metrics returns the metrics for the executor
	Metrics() *ExecutionMetrics
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
	tracer             oteltrace.Tracer
	meter              metric.Meter
	otelConfig         *otelconfig.OTelConfig
	metrics            *ExecutionMetrics // Execution metrics (initialized once)
}

func (e *executorImpl) Tracer() oteltrace.Tracer {
	return e.tracer
}

func (e *executorImpl) Meter() metric.Meter {
	return e.meter
}

func (e *executorImpl) OTelConfig() *otelconfig.OTelConfig {
	return e.otelConfig
}

func (e *executorImpl) Metrics() *ExecutionMetrics {
	return e.metrics
}

// NewExecutor builds an Executor with built-in @sentra/* modules registered.
func NewExecutor(idx *index.Index, opts ...NewExecutorOption) (Executor, error) {
	exec := &executorImpl{
		index:              idx,
		jsRegistry:         js.NewRegistry(idx.Pack.Location),
		moduleBindingPerch: perch.New[*ModuleBinding](100 << 20 /* 100 MB */), // --> (policy.useAlias) -> module binding
		callMemoizePerch:   perch.New[any](10 << 20 /* 10 MB */),
		tracer:             otel.Tracer("sentrie/executor"),
		meter:              otel.Meter("sentrie/executor"),
		metrics:            nil,
		otelConfig: &otelconfig.OTelConfig{
			Enabled:        false,
			TraceExecution: false,
		},
	}

	exec.jsRegistry.RegisterBuiltin("uuid", js.BuiltinUuidGo)
	exec.jsRegistry.RegisterBuiltin("crypto", js.BuiltinCryptoGo)
	exec.jsRegistry.RegisterBuiltin("base64", js.BuiltinBase64Go)

	for _, opt := range opts {
		opt(exec)
	}

	// Initialize execution metrics if tracing is enabled and meter is available
	if exec.otelConfig.Enabled && exec.otelConfig.TraceExecution && exec.meter != nil {
		exec.metrics = &ExecutionMetrics{}

		var err error
		exec.metrics.JSCallCount, err = exec.meter.Int64Counter(
			"sentrie.js.call.count",
			metric.WithDescription("Number of JavaScript function calls"),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create JS call count metric")
		}

		exec.metrics.JSCallDuration, err = exec.meter.Float64Histogram(
			"sentrie.js.call.duration",
			metric.WithDescription("JavaScript call execution duration in milliseconds"),
			metric.WithUnit("ms"),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create JS call duration metric")
		}

		exec.metrics.JSCallErrors, err = exec.meter.Int64Counter(
			"sentrie.js.call.errors",
			metric.WithDescription("Number of JavaScript call failures"),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create JS call errors metric")
		}
	}

	// Reserve the cache slots
	if err := exec.moduleBindingPerch.Reserve(); err != nil {
		return nil, err
	}

	if err := exec.callMemoizePerch.Reserve(); err != nil {
		return nil, err
	}

	return exec, nil
}

func (e *executorImpl) Index() *index.Index {
	return e.index
}

// ExecPolicy executes all exported rules and returns the results
func (e *executorImpl) ExecPolicy(ctx context.Context, namespace, policy string, facts map[string]any) ([]*ExecutorOutput, error) {
	// Use otelConfig for decision-level tracing
	var span oteltrace.Span
	if e.otelConfig.Enabled {
		ctx, span = e.tracer.Start(ctx, "executor.exec_policy")
		defer span.End()

		span.SetAttributes(
			attribute.String("sentrie.namespace", namespace),
			attribute.String("sentrie.policy", policy),
			attribute.Int("sentrie.facts.count", len(facts)),
		)
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		// Record policy execution duration metric
		if policyDuration, err := e.meter.Float64Histogram("sentrie.policy.exec.duration"); err == nil {
			policyDuration.Record(ctx, float64(duration.Nanoseconds())/1e6,
				metric.WithAttributes(
					attribute.String("sentrie.namespace", namespace),
					attribute.String("sentrie.policy", policy),
				),
			)
		}
	}()

	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		if e.otelConfig.Enabled && span != nil {
			span.RecordError(err)
		}
		return nil, err
	}

	theLock := &sync.Mutex{}
	var compositeErr error
	outputs := make([]*ExecutorOutput, 0, len(p.RuleExports))
	wg := &sync.WaitGroup{} // should this be a WaitGroup or an ErrorGroup?
	for _, ruleExport := range p.RuleExports {
		wg.Go(func() {
			output, err := e.ExecRule(ctx, namespace, policy, ruleExport.RuleName, facts)

			// now that we have the output, we can add it to the outputs slice,
			// but we need to lock the mutex to avoid race conditions
			theLock.Lock()
			defer theLock.Unlock()
			if err != nil {
				if e.otelConfig.Enabled && span != nil {
					span.RecordError(err)
				}
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
func (e *executorImpl) ExecRule(ctx context.Context, namespace, policy, rule string, injectFacts map[string]any) (*ExecutorOutput, error) {
	// Use otelConfig for decision-level tracing
	var span oteltrace.Span
	if e.otelConfig.Enabled {
		ctx, span = e.tracer.Start(ctx, "ExecRule")
		defer span.End()

		span.SetAttributes(
			attribute.String("sentrie.namespace", namespace),
			attribute.String("sentrie.policy", policy),
			attribute.String("sentrie.rule", rule),
			attribute.Int("sentrie.facts.count", len(injectFacts)),
		)
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		// Record rule execution duration metric
		if ruleDuration, err := e.meter.Float64Histogram("sentrie.rule.exec.duration"); err == nil {
			ruleDuration.Record(ctx, float64(duration.Nanoseconds())/1e6,
				metric.WithAttributes(
					attribute.String("sentrie.namespace", namespace),
					attribute.String("sentrie.policy", policy),
					attribute.String("sentrie.rule", rule),
				),
			)
		}
	}()

	// Validate exported
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		if e.otelConfig.Enabled && span != nil {
			span.RecordError(err)
		}
		return nil, err
	}
	if err := p.VerifyRuleExported(rule); err != nil {
		if e.otelConfig.Enabled && span != nil {
			span.RecordError(err)
		}
		return nil, err
	}

	ec := NewExecutionContext(p, e)
	defer ec.Dispose()

	for factName, factStatement := range p.Facts {
		// look for a value for this fact in the passed in facts map
		if _, ok := injectFacts[factName]; ok {
			if err := ec.InjectFact(ctx, factName, injectFacts[factName], false, factStatement.Type); err != nil {
				if e.otelConfig.Enabled && span != nil {
					span.RecordError(err)
				}
				return nil, err
			}
			continue // move on to the next fact
		}

		// if the fact is required, and no value was passed in, we error
		if factStatement.Required {
			return nil, xerr.ErrRequiredFact(factName)
		}

		// if the fact has a default value, evaluate it and inject it into the context
		if factStatement.Default != nil {
			// evaluate the default value, this will be injected into the context
			val, _, err := eval(ctx, ec, e, p, factStatement.Default)
			if err != nil {
				return nil, errors.Wrap(xerr.ErrUnresolvableFact(factName), err.Error())
			}

			// inject the default value
			if err := ec.InjectFact(ctx, factStatement.Name, val, true, factStatement.Type); err != nil {
				return nil, err
			}
		}

		// if the fact is required, and no value was passed in, and no default value was provided, we error
		if factStatement.Required && !ec.IsFactInjected(factName) {
			return nil, xerr.ErrRequiredFact(factName)
		}
	}

	// bind lets
	for k, v := range p.Lets {
		ec.InjectLet(k, v)
	}

	// Bind `use` modules
	if err := e.bindUses(ctx, ec, p); err != nil {
		if e.otelConfig.Enabled && span != nil {
			span.RecordError(err)
		}
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
	p, err := e.index.ResolvePolicy(namespace, policy)
	if err != nil {
		return nil, nil, nil, err
	}

	r, ok := p.Rules[rule]
	if !ok {
		return nil, nil, nil, xerr.ErrRuleNotFound(index.RuleFQN(namespace, policy, rule))
	}

	// Check for infinite recursion before evaluating the rule
	if err := ec.PushRefStack(rule); err != nil {
		return nil, nil, nil, err
	}
	defer ec.PopRefStack()

	// Wrap rule evaluation in a decision node
	ctx, ruleNode, done := trace.New(ctx, r.Node, "rule-outcome", map[string]any{
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
		stmt := p.Facts[name]
		// validate the value against the type
		if err := validateValueAgainstTypeRef(ctx, ec, e, p, value.value, value.typeRef, stmt.Span()); err != nil {
			return nil, nil, nil, err
		}
	}

	d, node, err := evaluateRuleOutcome(ctx, ec, e, p, r)
	ruleNode.Attach(node)
	ruleNode.SetResult(d)
	ruleNode.SetErr(err)
	if err != nil {
		return d, nil, ruleNode, err
	}

	// Compute attachment values if exported
	attachments := map[string]any{}
	if ex, ok := p.RuleExports[rule]; ok {
		for _, attachment := range ex.Attachments {
			ctx, attachmentNode, done := trace.New(ctx, attachment.Value, "attachment", map[string]any{
				"name": attachment.Name,
			})
			defer done()

			v, node, err := eval(ctx, ec, e, p, attachment.Value)
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
		ms, err := e.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, fileDir)
		if err != nil {
			return err
		}

		// Resolve and ensure program exists
		binding, _, err := e.getModuleBinding(ctx, use, ms)
		if err != nil {
			return err
		}

		ec.BindModule(use.As, binding)
	}
	return nil
}

// getModuleBinding resolves and caches a module binding for a given use statement and module spec.
func (e *executorImpl) getModuleBinding(ctx context.Context, use *ast.UseStatement, ms *js.ModuleSpec) (binding *ModuleBinding, _ bool, err error) {
	return e.moduleBindingPerch.Get(ctx, ms.KeyOrPath(), -1, func(ctx context.Context, _ string) (*ModuleBinding, error) {
		jsInstancePool, err := puddle.NewPool(&puddle.Config[*JSInstance]{
			Constructor: func(ctx context.Context) (*JSInstance, error) {
				b, err := e.jsBindingConstructor(ctx, use, ms)
				if err != nil {
					return nil, err
				}

				if e.otelConfig.Enabled && e.metrics != nil {
					counter, _, err := e.metrics.JSPoolGuages.Get(ctx, ms.KeyOrPath(), -1, func(ctx context.Context, _ string) (metric.Int64UpDownCounter, error) {
						return e.meter.Int64UpDownCounter(
							fmt.Sprintf("sentrie.js.pool.count.%s", strings.ReplaceAll(ms.KeyOrPath(), "/", ".")),
							metric.WithDescription("Number of JavaScript instances in the pool"),
						)
					})
					if err == nil {
						counter.Add(ctx, 1)
					}
				}
				return b, nil
			},
			Destructor: func(res *JSInstance) {
				if e.otelConfig.Enabled && e.metrics != nil {
					counter, _, err := e.metrics.JSPoolGuages.Get(ctx, ms.KeyOrPath(), -1, func(ctx context.Context, _ string) (metric.Int64UpDownCounter, error) {
						return e.meter.Int64UpDownCounter(
							fmt.Sprintf("sentrie.js.pool.count.%s", strings.ReplaceAll(ms.KeyOrPath(), "/", ".")),
							metric.WithDescription("Number of JavaScript instances in the pool"),
						)
					})
					if err == nil {
						counter.Add(ctx, -1)
					}
				}
				// clear the interrupt
				res.VM.ClearInterrupt()
			},
			MaxSize: 10,
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
			VMPool:       jsInstancePool,
		}, nil
	})
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
