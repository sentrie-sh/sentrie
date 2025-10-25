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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func evalBlock(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, block *ast.BlockExpression) (any, *trace.Node, error) {
	n, done := trace.New("block", "", block, map[string]any{})
	defer done()

	// Create OpenTelemetry span for JavaScript calls if tracing is enabled
	var span oteltrace.Span
	if cfg := ec.executor.OTelConfig(); cfg.Enabled && cfg.TraceExecution {
		ctx, span = ec.executor.Tracer().Start(ctx, "block")
		defer span.End()

		span.SetAttributes(
			attribute.String("sentrie.ast.node.kind", block.Kind()),
			attribute.String("sentrie.ast.node.range", block.Span().String()),
		)
	}

	ec = ec.AttachedChildContext()
	defer ec.Dispose()

	for _, s := range block.Statements {
		switch st := s.(type) {
		case *ast.VarDeclaration:
			if ec.IsLetInjected(st.Name) {
				e := xerr.ErrConflict(st.Name)
				return nil, n.SetErr(e), e
			}
			ec.InjectLet(st.Name, st)
		case *ast.CommentStatement:
			_ = "noop"
		default:
			n.Attach(trace.IgnoredStmt(st))
		}
	}

	val, child, err := eval(ctx, ec, exec, p, block.Yield)
	if err != nil {
		return nil, n.SetErr(err), err
	}
	n.Attach(child).SetResult(val).SetErr(err)
	return val, n, err
}
