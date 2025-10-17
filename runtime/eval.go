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

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
)

// eval walks an ast.Expression and returns (value, decision node, error).
func eval(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, e ast.Expression) (any, *trace.Node, error) {
	switch t := e.(type) {

	case *ast.PrecedingCommentExpression:
		// evaluate the wrapped expression, then return the value
		return eval(ctx, ec, exec, p, t.Wrap)

	case *ast.TrailingCommentExpression:
		// evaluate the wrapped expression, then return the value
		return eval(ctx, ec, exec, p, t.Wrap)

	case *ast.NullLiteral:
		n, done := trace.New("literal", "nil", t, map[string]any{})
		defer done()

		n.SetResult(nil)
		return nil, n, nil

	case *ast.TrinaryLiteral:
		n, done := trace.New("literal", "tristate", t, map[string]any{})
		defer done()

		n.SetResult(t.Value)
		return t.Value, n, nil

	case *ast.IntegerLiteral:
		n, done := trace.New("literal", "int", t, map[string]any{})
		defer done()

		n.SetResult(t.Value)
		return t.Value, n, nil

	case *ast.FloatLiteral:
		n, done := trace.New("literal", "float", t, map[string]any{})
		defer done()

		n.SetResult(t.Value)
		return t.Value, n, nil

	case *ast.StringLiteral:
		n, done := trace.New("literal", "string", t, map[string]any{})
		defer done()

		n.SetResult(t.Value)
		return t.Value, n, nil

	case *ast.ListLiteral:
		n, done := trace.New("literal", "list", t, map[string]any{})
		defer done()

		arr := make([]any, 0, len(t.Values))
		for _, it := range t.Values {
			v, child, err := eval(ctx, ec, exec, p, it)
			n.Attach(child)
			if err != nil {
				return nil, n.SetErr(err), err
			}
			arr = append(arr, v)
		}
		return arr, n.SetResult(arr), nil

	case *ast.MapLiteral:
		n, done := trace.New("literal", "map", t, map[string]any{})
		defer done()

		m := map[string]any{}
		for _, kv := range t.Entries {
			key, child, err := eval(ctx, ec, exec, p, kv.Key)
			n.Attach(child)
			if err != nil {
				return nil, n.SetErr(err), err
			}
			keyValue, ok := key.(string)
			if !ok {
				err := errors.Wrapf(xerr.ErrInvalidType(fmt.Sprintf("%T", key), "string"), "map key is not a string at %s", kv.Key.Position())
				return nil, n.SetErr(err), err
			}

			v, child, err := eval(ctx, ec, exec, p, kv.Value)
			n.Attach(child)
			if err != nil {
				return nil, n.SetErr(err), err
			}
			m[keyValue] = v
		}
		return m, n.SetResult(m), nil

	case *ast.Identifier:
		return evalIdent(ctx, ec, exec, p, t.Value)

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

	case *ast.AnyExpression:
		return evalAny(ctx, ec, exec, p, t)

	case *ast.AllExpression:
		return evalAll(ctx, ec, exec, p, t)

	case *ast.FilterExpression:
		return evalFilter(ctx, ec, exec, p, t)

	case *ast.ReduceExpression:
		return evalReduce(ctx, ec, exec, p, t)

	case *ast.MapExpression:
		return evalMap(ctx, ec, exec, p, t)

	case *ast.TransformExpression:
		return evalTransform(ctx, ec, exec, p, t)

	case *ast.DistinctExpression:
		return evalDistinct(ctx, ec, exec, p, t)

	default:
		err := fmt.Errorf("unsupported expression node: %T", t)
		return nil, trace.UnsupportedExpression(t).SetErr(err), err
	}
}
