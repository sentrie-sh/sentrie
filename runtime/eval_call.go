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
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func evalCall(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.CallExpression) (response any, traceNode *trace.Node, err error) {
	n, done := trace.New("call", "", t, map[string]any{
		"target": t.Callee.String(),
		"args":   t.Arguments,
	})
	defer done()

	// Create OpenTelemetry span for JavaScript calls if tracing is enabled
	var span oteltrace.Span
	if ec.executor.TraceExecution() {
		ctx, span = ec.executor.Tracer().Start(ctx, "call")
		defer span.End()

		span.SetAttributes(
			attribute.String("sentrie.ast.node.kind", "call"),
			attribute.String("sentrie.ast.node.range", t.Span().String()),
			attribute.String("sentrie.call.target", t.Callee.String()),
			attribute.Int("sentrie.call.args.count", len(t.Arguments)),
			attribute.String("sentrie.ast.node.range", t.Span().String()),
		)
	}

	args := make([]any, 0, len(t.Arguments))
	for _, a := range t.Arguments {
		v, child, err := eval(ctx, ec, exec, p, a)
		n.Attach(child)
		if err != nil {
			return nil, n.SetErr(err), err
		}
		args = append(args, v)
	}

	target, err := getTarget(ctx, ec, p, t)
	if err != nil {
		return nil, n.SetErr(err), err
	}

	// use a thin wrapper around the target to handle the caching
	wrappedTarget := func(ctx context.Context, args ...any) (any, error) {
		if !t.Memoized {
			// no memoization, so we can just call the target
			// quickly call the target without caching
			return target(ctx, args...)
		}

		ttl := 5 * time.Minute // default to 5 minutes
		if t.MemoizeTTL != nil {
			ttl = *t.MemoizeTTL
		}

		hashKey := calculateHashKey(t, args)
		loader := func(ctx context.Context, key string) (any, error) {
			return target(ctx, args...)
		}
		out, _, err := exec.callMemoizePerch.Get(ctx, hashKey, ttl, loader)
		return out, err
	}

	// call the target
	out, err := wrappedTarget(ctx, args...)
	if err != nil {
		if errors.Is(err, xerr.InjectedError{}) {
			// if this error is injected from code, we revert to the error message
			return nil, n.SetErr(err), err
		}
		err = errors.Wrapf(err, "failed to call function '%s'", t.Callee.String())
		return nil, n.SetErr(err), err
	}
	return out, n.SetResult(out), nil
}

// Helper to split "alias.fn" if ever needed
func splitAliasFn(s string) (string, string) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return s, ""
	}
	return parts[0], parts[1]
}

func calculateHashKey(node *ast.CallExpression, args []any) string {
	arghash, err := hashstructure.Hash(args, hashstructure.FormatV2, nil)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%p:%d", node, arghash)
}

func getTarget(_ context.Context, ec *ExecutionContext, p *index.Policy, c *ast.CallExpression) (func(context.Context, ...any) (any, error), error) {
	callee := c.Callee.String()

	// check if we have a builtin function
	if builtin, ok := Builtins[callee]; ok {
		return func(ctx context.Context, args ...any) (any, error) {
			return builtin(ctx, args)
		}, nil
	}

	// otherwise, assume that's a module function
	module, fn := splitAliasFn(callee)

	// if the module or fn are empty, it's a problem
	if module == "" || fn == "" {
		e := xerr.ErrImportResolution(module, p.Namespace.FQN.String())
		return nil, e
	}

	modulebinding, ok := ec.Module(module)
	if !ok {
		e := xerr.ErrModuleInvocation(module, fn)
		return nil, e
	}

	return func(ctx context.Context, args ...any) (any, error) {
		start := time.Now()
		result, err := modulebinding.Call(ctx, fn, args...)

		// Record metrics using the executor's stored instruments
		if metrics := ec.executor.Metrics(); metrics != nil {
			attrs := []attribute.KeyValue{
				attribute.String("sentrie.js.call.module.canonical_key", modulebinding.CanonicalKey),
				attribute.String("sentrie.js.call.module.alias", modulebinding.Alias),
				attribute.String("sentrie.js.call.module.function", fn),
			}
			if err != nil {
				attrs = append(attrs, attribute.String("sentrie.js.call.error", err.Error()))
			}

			attrSet := attribute.NewSet(attrs...)
			dur := time.Since(start)

			// Record duration in milliseconds
			metrics.JSCallDuration.Record(ctx, float64(dur.Nanoseconds())/1e6, metric.WithAttributeSet(attrSet))
			if err != nil {
				metrics.JSCallErrors.Add(ctx, 1, metric.WithAttributeSet(attrSet))
			}
			metrics.JSCallCount.Add(ctx, 1, metric.WithAttributeSet(attrSet))
		}

		return result, err
	}, nil
}
