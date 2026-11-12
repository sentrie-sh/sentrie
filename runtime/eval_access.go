// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"math"

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
		i, found, err := listIndex(idx, len(c))
		if err != nil || !found {
			return box.Undefined(), err
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
			i, found, err := listIndex(idx, len(c))
			if err != nil || !found {
				return box.Undefined(), err
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

// listIndex resolves idx against a list of length n. found is false, with no
// error, for an undefined index (undefined propagates) and for an integer
// outside the list, including one too large for an int. Any other index is an
// error: a string, bool, null or container, and a number with a fractional
// part. An integral float such as 2.0 is the integer it holds.
func listIndex(idx box.Value, n int) (i int, found bool, err error) {
	if idx.IsUndefined() {
		return 0, false, nil
	}
	f, ok := idx.NumberValue()
	if !ok {
		return 0, false, fmt.Errorf("list index must be an integer, got %s", idx.Kind())
	}
	if iv, ok := idx.IntegralValue(); ok {
		if iv < 0 || iv >= int64(n) {
			return 0, false, nil
		}
		return int(iv), true, nil
	}
	if f != math.Trunc(f) {
		return 0, false, fmt.Errorf("list index must be an integer, got %s", idx.String())
	}
	// Integral but outside int64, or an infinity: out of range.
	return 0, false, nil
}
