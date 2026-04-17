// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
)

// eval walks an ast.Expression and returns (value, decision node, error).
func eval(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, e ast.Expression) (box.Value, *trace.Node, error) {
	switch t := e.(type) {

	case *ast.PrecedingCommentExpression:
		// evaluate the wrapped expression, then return the value
		return eval(ctx, ec, exec, p, t.Wrap)

	case *ast.TrailingCommentExpression:
		// evaluate the wrapped expression, then return the value
		return eval(ctx, ec, exec, p, t.Wrap)

	case *ast.NullLiteral:
		_, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "null"})
		defer done()
		v := box.Null()
		n.SetResult(v)
		return v, n, nil

	case *ast.TrinaryLiteral:
		_, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "trinary"})
		defer done()
		v := box.Trinary(t.Value)
		n.SetResult(v)
		return v, n, nil

	case *ast.IntegerLiteral:
		_, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "number"})
		defer done()
		v := box.Number(t.Value)
		n.SetResult(v)
		return v, n, nil

	case *ast.FloatLiteral:
		_, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "number"})
		defer done()
		v := box.Number(t.Value)
		n.SetResult(v)
		return v, n, nil

	case *ast.StringLiteral:
		_, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "string"})
		defer done()

		v := box.String(t.Value)
		n.SetResult(v)
		return v, n, nil

	case *ast.ListLiteral:
		ctx, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "list"})
		defer done()

		arr := make([]box.Value, 0, len(t.Values))
		for _, it := range t.Values {
			v, child, err := eval(ctx, ec, exec, p, it)
			n.Attach(child)
			if err != nil {
				return box.Undefined(), n.SetErr(err), err
			}
			arr = append(arr, v)
		}
		out := box.List(arr)
		return out, n.SetResult(out), nil

	case *ast.MapLiteral:
		ctx, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "dict"})
		defer done()

		m := map[string]box.Value{}
		for _, kv := range t.Entries {
			key, child, err := eval(ctx, ec, exec, p, kv.Key)
			n.Attach(child)
			if err != nil {
				return box.Undefined(), n.SetErr(err), err
			}
			keyValue, ok := key.StringValue()
			if !ok {
				err := fmt.Errorf("map key is not a string at %s: %w", kv.Key.Span(), xerr.ErrInvalidType(fmt.Sprintf("%T", key), "string"))
				return box.Undefined(), n.SetErr(err), err
			}

			v, child, err := eval(ctx, ec, exec, p, kv.Value)
			n.Attach(child)
			if err != nil {
				return box.Undefined(), n.SetErr(err), err
			}
			m[keyValue] = v
		}
		out := box.Dict(m)
		return out, n.SetResult(out), nil

	case *ast.Identifier:
		return evalIdent(ctx, ec, exec, p, t)

	case *ast.PipelineHoleExpression:
		err := fmt.Errorf("pipeline placeholder '#' must be used inside a pipeline call target")
		return box.Undefined(), trace.UnsupportedExpression(t).SetErr(err), err

	case *ast.CastExpression:
		return evalCast(ctx, ec, exec, p, t)

	case *ast.UnaryExpression:
		return evalUnary(ctx, ec, exec, p, t)

	case *ast.InfixExpression:
		return evalInfix(ctx, ec, exec, p, t)

	case *ast.BlockExpression:
		return evalBlock(ctx, ec, exec, p, t)

	case *ast.FieldAccessExpression:
		return evalFieldAccess(ctx, ec, exec, p, t)

	case *ast.IndexAccessExpression:
		return evalIndexAccess(ctx, ec, exec, p, t)

	case *ast.CallExpression:
		return evalCall(ctx, ec, exec, p, t)

	case *ast.ImportClause:
		return ImportDecision(ctx, exec, ec, p, t)

	case *ast.TernaryExpression:
		return evalTernary(ctx, ec, exec, p, t)

	case *ast.LambdaExpression:
		return evalLambda(ctx, ec, exec, p, t)

	case *ast.TransformExpression:
		return evalTransform(ctx, ec, exec, p, t)

	default:
		err := fmt.Errorf("unsupported expression node: %T", t)
		return box.Undefined(), trace.UnsupportedExpression(t).SetErr(err), err
	}
}
