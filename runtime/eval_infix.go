// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
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

// concatStrings implements `+` when either operand is a string. ok is false
// when neither is, so the caller falls through to numeric addition. Scalars
// join the way JavaScript joins them, in either order: a number renders as
// JavaScript's Number.prototype.toString would (jsNumberString), a bool as
// true or false, null as null. A list, dict, document or callable operand is
// an error rather than a Go rendering of the container. An undefined or
// unknown operand, in either order, makes the result the trinary unknown: the
// text cannot be known, and unlike other operators string `+` does not
// propagate undefined.
func concatStrings(l, r box.Value) (out box.Value, ok bool, err error) {
	if l.Kind() != box.ValueString && r.Kind() != box.ValueString {
		return box.Value{}, false, nil
	}
	if unknownConcatOperand(l) || unknownConcatOperand(r) {
		return box.Trinary(trinary.Unknown), true, nil
	}
	ls, err := concatOperand(l)
	if err != nil {
		return box.Value{}, true, err
	}
	rs, err := concatOperand(r)
	if err != nil {
		return box.Value{}, true, err
	}
	return box.String(ls + rs), true, nil
}

// unknownConcatOperand reports whether v is undefined or the trinary unknown
// (the only value that boxes as ValueTrinary).
func unknownConcatOperand(v box.Value) bool {
	return v.IsUndefined() || v.Kind() == box.ValueTrinary
}

// concatOperand renders one side of a string concatenation.
func concatOperand(v box.Value) (string, error) {
	switch v.Kind() {
	case box.ValueString, box.ValueBool, box.ValueNull:
		return v.String(), nil
	case box.ValueNumber:
		return jsNumberString(v), nil
	default:
		return "", fmt.Errorf("+ cannot join a string with a %s", v.Kind())
	}
}

// jsNumberString renders a number the way JavaScript's Number.prototype.toString
// does, so string concatenation matches JavaScript: 1.0 is "1", 1e20 is
// "100000000000000000000", 1e21 is "1e+21", 0.000001 is "0.000001", 1e-7 is
// "1e-7", -0 is "0" and the infinities are "Infinity" and "-Infinity". An
// exact integer renders all its digits, as a JavaScript BigInt would.
func jsNumberString(v box.Value) string {
	if i, ok := v.IntValue(); ok {
		return strconv.FormatInt(i, 10)
	}
	f, _ := v.NumberValue()
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0:
		return "0"
	}
	sign := ""
	if f < 0 {
		sign = "-"
		f = -f
	}
	// The shortest round-trip digits, as d.ddde±x; JavaScript chooses the
	// same digits and differs only in where it puts the decimal point.
	e := strconv.FormatFloat(f, 'e', -1, 64)
	mant, exp, _ := strings.Cut(e, "e")
	digits := strings.Replace(mant, ".", "", 1)
	x, _ := strconv.Atoi(exp)
	k := len(digits)
	n := x + 1 // the decimal point sits after n digits
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	expSign := "+"
	if n-1 < 0 {
		expSign = "-"
	}
	exponent := expSign + strconv.Itoa(abs(n-1))
	if k == 1 {
		return sign + digits + "e" + exponent
	}
	return sign + digits[:1] + "." + digits[1:] + "e" + exponent
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func evalInfix(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, in, "infix", map[string]any{
		"operator": in.Operator,
	})
	defer done()

	switch in.Operator {
	case "and", "or", "xor":
		return evalLogical(ctx, ec, exec, p, in, node)
	}

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

	// String concatenation runs before undefined propagation: a string joined
	// with undefined is unknown, not undefined (concatStrings).
	if in.Operator == "+" {
		if out, ok, err := concatStrings(l, r); ok {
			if err != nil {
				return box.Undefined(), node.SetErr(err), err
			}
			return out, node.SetResult(out), nil
		}
	}

	if l.IsUndefined() || r.IsUndefined() {
		return box.Undefined(), node.SetResult(box.Undefined()), nil
	}

	switch in.Operator {
	case "+":
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
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c < 0)
			return out, node.SetResult(out), nil
		}
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c < 0)
		return out, node.SetResult(out), nil
	case "<=":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c <= 0)
			return out, node.SetResult(out), nil
		}
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c <= 0)
		return out, node.SetResult(out), nil
	case ">":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c > 0)
			return out, node.SetResult(out), nil
		}
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c > 0)
		return out, node.SetResult(out), nil
	case ">=":
		if c, ok := compareStrings(l, r); ok {
			out := box.Bool(c >= 0)
			return out, node.SetResult(out), nil
		}
		c, ordered, err := box.CompareNumbers(l, r)
		if err != nil {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(ordered && c >= 0)
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

// evalLogical evaluates `and`, `or` and `xor` with Kleene three-valued logic.
// An undefined or null operand reads as unknown through box.TrinaryFrom, so
// it does not absorb the result the way it does for the other operators.
// `and` and `or` short-circuit: when the left side decides the result (false
// for `and`, true for `or`) the right side is never evaluated, so its errors,
// module calls and trace node never happen. `xor` always needs both sides.
func evalLogical(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression, node *trace.Node) (box.Value, *trace.Node, error) {
	l, ln, err := eval(ctx, ec, exec, p, in.Left)
	node.Attach(ln)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}
	left := box.TrinaryFrom(l)

	if (in.Operator == "and" && left == trinary.False) || (in.Operator == "or" && left == trinary.True) {
		out := box.Trinary(left)
		return out, node.SetResult(out), nil
	}

	r, rn, err := eval(ctx, ec, exec, p, in.Right)
	node.Attach(rn)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}
	right := box.TrinaryFrom(r)

	var res trinary.Value
	switch in.Operator {
	case "and":
		res = left.And(right)
	case "or":
		res = left.Or(right)
	default: // xor
		res = left.Or(right).And(left.And(right).Not())
	}
	out := box.Trinary(res)
	return out, node.SetResult(out), nil
}
