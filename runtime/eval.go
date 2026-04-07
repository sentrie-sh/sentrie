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
		ctx, n, done := trace.New(ctx, t, "literal", map[string]any{"type": "map"})
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
		out := box.Map(m)
		return out, n.SetResult(out), nil

	case *ast.Identifier:
		return evalIdent(ctx, ec, exec, p, t)

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

	case *ast.FirstExpression:
		return evalFirst(ctx, ec, exec, p, t)

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
		return box.Undefined(), trace.UnsupportedExpression(t).SetErr(err), err
	}
}
