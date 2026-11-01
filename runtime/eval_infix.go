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
)

func evalInfix(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, in, "infix", map[string]any{
		"operator": in.Operator,
	})
	defer done()

	l, ln, err := eval(ctx, ec, exec, p, in.Left)
	node.Attach(ln)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}
	r, rn, err := eval(ctx, ec, exec, p, in.Right)
	node.Attach(rn)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}

	if l.IsUndefined() || r.IsUndefined() {
		return box.Undefined(), node.SetResult(box.Undefined()), nil
	}

	switch in.Operator {
	case "+":
		if ls, ok := l.StringValue(); ok {
			out := box.String(ls + r.String())
			return out, node.SetResult(out), nil
		}
		if rs, ok := r.StringValue(); ok {
			out := box.String(l.String() + rs)
			return out, node.SetResult(out), nil
		}
		out, err := box.Add(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		return out, node.SetResult(out), nil
	case "-":
		out, err := box.Sub(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		return out, node.SetResult(out), nil
	case "*":
		out, err := box.Mul(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		return out, node.SetResult(out), nil
	case "/":
		out, err := box.Div(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		return out, node.SetResult(out), nil
	case "%":
		out, err := box.Mod(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		return out, node.SetResult(out), nil

	case "==", "is":
		out := box.Bool(box.EqualValues(l, r))
		return out, node.SetResult(out), nil
	case "!=":
		out := box.Bool(!box.EqualValues(l, r))
		return out, node.SetResult(out), nil
	case "<":
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c < 0)
		return out, node.SetResult(out), nil
	case "<=":
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c <= 0)
		return out, node.SetResult(out), nil
	case ">":
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c > 0)
		return out, node.SetResult(out), nil
	case ">=":
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c >= 0)
		return out, node.SetResult(out), nil

	case "and":
		out := box.Trinary(box.TrinaryFrom(l).And(box.TrinaryFrom(r)))
		return out, node.SetResult(out), nil

	case "or":
		out := box.Trinary(box.TrinaryFrom(l).Or(box.TrinaryFrom(r)))
		return out, node.SetResult(out), nil

	case "xor":
		left := box.TrinaryFrom(l)
		right := box.TrinaryFrom(r)
		out := box.Trinary(left.Or(right).And(left.And(right).Not()))
		return out, node.SetResult(out), nil

	case "in":
		out := box.Bool(box.ContainsValue(r, l))
		return out, node.SetResult(out), nil

	case "contains":
		out := box.Bool(box.ContainsValue(l, r))
		return out, node.SetResult(out), nil

	case "matches":
		out, err := box.MatchesValue(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		b := box.Bool(out)
		return b, node.SetResult(b), nil

	default:
		err := fmt.Errorf("unsupported infix op: %s", in.Operator)
		return box.Undefined(), node.SetErr(err), err
	}
}
