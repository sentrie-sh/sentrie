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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalFieldAccess(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.FieldAccessExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "field_access", map[string]any{
		"field": t.Field,
	})
	defer done()

	recv, rn, err := eval(ctx, ec, exec, p, t.Left)
	if err != nil {
		return nil, node.SetErr(err), err
	}
	node.Attach(rn)
	out, err := accessField(ctx, recv, t.Field)
	node.SetResult(out).SetErr(err)
	return out, node, err
}

func evalIndexAccess(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.IndexAccessExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "index_access", map[string]any{
		"index": t.Index,
	})
	defer done()

	col, cn, err := eval(ctx, ec, exec, p, t.Left)
	if err != nil {
		return nil, node.SetErr(err), err
	}
	node.Attach(cn)

	idx, in, err := eval(ctx, ec, exec, p, t.Index)
	node.Attach(in)
	if err != nil {
		return nil, node.SetErr(err), err
	}
	out, err := accessIndex(ctx, col, idx)
	node.SetResult(out).SetErr(err)
	return out, node, err
}

func accessField(_ context.Context, obj any, field string) (any, error) {
	if IsUndefined(obj) {
		return Undefined, nil
	}
	switch o := obj.(type) {
	case map[string]any:
		if v, ok := o[field]; ok {
			return v, nil
		}
		return Undefined, nil
	case *ExecutorOutput:
		switch field {
		case "state":
			return o.Decision.State, nil
		case "value":
			return o.Decision.Value, nil
		default:
			if v, ok := o.Attachments[field]; ok {
				return v, nil
			}
			return Undefined, nil
		}
	default:
		return nil, fmt.Errorf("cannot access field '%s' on %T", field, obj)
	}
}

func accessIndex(_ context.Context, col any, idx any) (any, error) {
	if IsUndefined(col) {
		return Undefined, nil
	}
	switch c := col.(type) {
	case []any:
		i := int(num(idx))
		if i < 0 || i >= len(c) {
			return Undefined, nil
		}
		return c[i], nil
	case map[string]any:
		if s, ok := idx.(string); ok {
			return c[s], nil
		}
		return Undefined, nil
	default:
		return nil, fmt.Errorf("index access not supported on %T", col)
	}
}
