// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalFieldAccess(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.FieldAccessExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "field_access", map[string]any{
		"field": t.Field,
	})
	defer done()

	recv, rn, err := eval(ctx, ec, exec, p, t.Left)
	if err != nil {
		return box.Value{}, node.SetErr(err), err
	}
	node.Attach(rn)
	out, err := accessField(ctx, recv, t.Field)
	node.SetResult(out).SetErr(err)
	return out, node, err
}

func evalIndexAccess(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.IndexAccessExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "index_access", map[string]any{
		"index": t.Index,
	})
	defer done()

	col, cn, err := eval(ctx, ec, exec, p, t.Left)
	if err != nil {
		return box.Value{}, node.SetErr(err), err
	}
	node.Attach(cn)

	idx, in, err := eval(ctx, ec, exec, p, t.Index)
	node.Attach(in)
	if err != nil {
		return box.Value{}, node.SetErr(err), err
	}
	out, err := accessIndex(ctx, col, idx)
	node.SetResult(out).SetErr(err)
	return out, node, err
}

func accessField(_ context.Context, obj box.Value, field string) (box.Value, error) {
	if obj.IsUndefined() {
		return box.Undefined(), nil
	}
	if m, ok := obj.DictValue(); ok {
		if v, exists := m[field]; exists {
			return v, nil
		}
		return box.Undefined(), nil
	}
	if ref, ok := obj.ObjectRef(); ok {
		switch o := ref.(type) {
		case map[string]any:
			if v, ok := o[field]; ok {
				return box.FromBoundaryAny(v), nil
			}
			return box.Undefined(), nil
		default:
			return box.Value{}, fmt.Errorf("cannot access field '%s' on %T", field, obj)
		}
	}
	return box.Value{}, fmt.Errorf("cannot access field '%s' on %T", field, obj)
}

func accessIndex(_ context.Context, col box.Value, idx box.Value) (box.Value, error) {
	if col.IsUndefined() {
		return box.Undefined(), nil
	}
	if c, ok := col.ListValue(); ok {
		n, _ := idx.NumberValue()
		i := int(n)
		if i < 0 || i >= len(c) {
			return box.Undefined(), nil
		}
		return c[i], nil
	}
	if c, ok := col.DictValue(); ok {
		if s, sok := idx.StringValue(); sok {
			if v, exists := c[s]; exists {
				return v, nil
			}
		}
		return box.Undefined(), nil
	}
	if ref, ok := col.ObjectRef(); ok {
		switch c := ref.(type) {
		case []any:
			n, _ := idx.NumberValue()
			i := int(n)
			if i < 0 || i >= len(c) {
				return box.Undefined(), nil
			}
			return box.FromBoundaryAny(c[i]), nil
		case map[string]any:
			if s, ok := idx.StringValue(); ok {
				if v, exists := c[s]; exists {
					return box.FromBoundaryAny(v), nil
				}
				return box.Undefined(), nil
			}
			return box.Undefined(), nil
		default:
			return box.Value{}, fmt.Errorf("index access not supported on %T", col)
		}
	}
	return box.Value{}, fmt.Errorf("index access not supported on %T", col)
}
