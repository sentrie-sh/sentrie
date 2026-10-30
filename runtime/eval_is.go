// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"reflect"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

// evalIsDefined evaluates 'x is defined'. The result is true unless the operand
// evaluates to undefined; null is defined. An undefined operand yields a definite
// false rather than propagating, while an evaluation error still propagates.
// 'x is not defined' is parsed as a 'not' unary around this node.
func evalIsDefined(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.IsDefinedExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "is", map[string]any{
		"operator": "defined",
	})
	defer done()

	v, child, err := eval(ctx, ec, exec, p, t.Left)
	node.Attach(child)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}

	out := box.Bool(!v.IsUndefined())
	return out, node.SetResult(out), nil
}

// evalIsEmpty evaluates 'x is empty'. The result is true for "", [], {}, null and
// undefined, and false for every other value, scalars included. An undefined
// operand yields a definite true rather than propagating, while an evaluation
// error still propagates. 'x is not empty' is parsed as a 'not' unary around this node.
func evalIsEmpty(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.IsEmptyExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, t, "is", map[string]any{
		"operator": "empty",
	})
	defer done()

	v, child, err := eval(ctx, ec, exec, p, t.Left)
	node.Attach(child)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}

	out := box.Bool(isEmptyValue(v))
	return out, node.SetResult(out), nil
}

// isEmptyValue reports whether v is undefined, null, or a zero-length string, list
// or dict. A document wraps a host value, typically a typed Go map, slice or string
// a builtin returned (an untyped map[string]any or []any is already a dict or list),
// so it is checked by reflected length; any other host value is never empty.
func isEmptyValue(v box.Value) bool {
	switch v.Kind() {
	case box.ValueUndefined, box.ValueNull:
		return true
	case box.ValueString:
		s, _ := v.StringValue()
		return len(s) == 0
	case box.ValueList:
		xs, _ := v.ListValue()
		return len(xs) == 0
	case box.ValueDict:
		m, _ := v.DictValue()
		return len(m) == 0
	case box.ValueDocument:
		ref, _ := v.DocumentRef()
		rv := reflect.ValueOf(ref)
		switch rv.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array, reflect.String:
			return rv.Len() == 0
		}
		return false
	default:
		return false
	}
}
