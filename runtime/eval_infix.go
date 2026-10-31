// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

// compareStrings orders two string operands by byte order (strings.Compare),
// which is not locale-aware: "B" < "a" and "Z" < "a". ok is false unless both
// operands are strings, so the relational operators fall through to the
// numeric path and a mixed string/number pair keeps its operand error.
func compareStrings(l, r box.Value) (c int, ok bool) {
	ls, ok := l.StringValue()
	if !ok {
		return 0, false
	}
	rs, ok := r.StringValue()
	if !ok {
		return 0, false
	}
	return strings.Compare(ls, rs), true
}

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
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Number(ln + rn)
		return out, node.SetResult(out), nil
	case "-":
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Number(ln - rn)
		return out, node.SetResult(out), nil
	case "*":
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Number(ln * rn)
		return out, node.SetResult(out), nil
	case "/":
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		if rn == 0 {
			err := fmt.Errorf("divide by zero")
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Number(ln / rn)
		return out, node.SetResult(out), nil
	case "%":
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		if rn == 0 {
			err := fmt.Errorf("divide by zero")
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Number(math.Mod(ln, rn))
		return out, node.SetResult(out), nil

	case "==", "is":
		out := box.Bool(box.EqualValues(l, r))
		return out, node.SetResult(out), nil
	case "!=":
		out := box.Bool(!box.EqualValues(l, r))
		return out, node.SetResult(out), nil
	case "<":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c < 0)
			return out, node.SetResult(out), nil
		}
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ln < rn)
		return out, node.SetResult(out), nil
	case "<=":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c <= 0)
			return out, node.SetResult(out), nil
		}
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ln <= rn)
		return out, node.SetResult(out), nil
	case ">":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c > 0)
			return out, node.SetResult(out), nil
		}
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ln > rn)
		return out, node.SetResult(out), nil
	case ">=":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c >= 0)
			return out, node.SetResult(out), nil
		}
		ln, rn, err := box.MustNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ln >= rn)
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
