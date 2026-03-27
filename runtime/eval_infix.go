// SPDX-License-Identifier: Apache-2.0
//
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
	"math"
	"regexp"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

func evalInfix(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression) (Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, in, "infix", map[string]any{
		"operator": in.Operator,
	})
	defer done()

	l, ln, err := eval(ctx, ec, exec, p, in.Left)
	node.Attach(ln)
	if err != nil {
		return Value{}, node.SetErr(err), err
	}
	r, rn, err := eval(ctx, ec, exec, p, in.Right)
	node.Attach(rn)
	if err != nil {
		return Value{}, node.SetErr(err), err
	}

	if l.IsUndefined() || r.IsUndefined() {
		return Undefined(), node.SetResult(Undefined()), nil
	}

	switch in.Operator {
	case "+":
		if ls, ok := l.StringValue(); ok {
			out := String(ls + r.String())
			return out, node.SetResult(out.Any()), nil
		}
		if rs, ok := r.StringValue(); ok {
			out := String(l.String() + rs)
			return out, node.SetResult(out.Any()), nil
		}
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Number(ln + rn)
		return out, node.SetResult(out.Any()), nil
	case "-":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Number(ln - rn)
		return out, node.SetResult(out.Any()), nil
	case "*":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Number(ln * rn)
		return out, node.SetResult(out.Any()), nil
	case "/":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		if rn == 0 {
			err := fmt.Errorf("divide by zero")
			return Value{}, node.SetErr(err), err
		}
		out := Number(ln / rn)
		return out, node.SetResult(out.Any()), nil
	case "%":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		if rn == 0 {
			err := fmt.Errorf("divide by zero")
			return Value{}, node.SetErr(err), err
		}
		out := Number(math.Mod(ln, rn))
		return out, node.SetResult(out.Any()), nil

	case "==", "is":
		out := Bool(equalValues(l, r))
		return out, node.SetResult(out.Any()), nil
	case "!=":
		out := Bool(!equalValues(l, r))
		return out, node.SetResult(out.Any()), nil
	case "<":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Bool(ln < rn)
		return out, node.SetResult(out.Any()), nil
	case "<=":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Bool(ln <= rn)
		return out, node.SetResult(out.Any()), nil
	case ">":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Bool(ln > rn)
		return out, node.SetResult(out.Any()), nil
	case ">=":
		ln, rn, err := mustNumbers(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		out := Bool(ln >= rn)
		return out, node.SetResult(out.Any()), nil

	case "and":
		out := Trinary(trinary.From(l.Any()).And(trinary.From(r.Any())))
		return out, node.SetResult(out.Any()), nil

	case "or":
		out := Trinary(trinary.From(l.Any()).Or(trinary.From(r.Any())))
		return out, node.SetResult(out.Any()), nil

	case "xor":
		left := trinary.From(l.Any())
		right := trinary.From(r.Any())
		out := Trinary(left.Or(right).And(left.And(right).Not()))
		return out, node.SetResult(out.Any()), nil

	case "in":
		out := Bool(containsValue(r, l))
		return out, node.SetResult(out.Any()), nil

	case "contains":
		out := Bool(containsValue(l, r))
		return out, node.SetResult(out.Any()), nil

	case "matches":
		out, err := matchesValue(l, r)
		if err != nil {
			return Value{}, node.SetErr(err), err
		}
		b := Bool(out)
		return b, node.SetResult(b.Any()), nil

	default:
		err := fmt.Errorf("unsupported infix op: %s", in.Operator)
		return Value{}, node.SetErr(err), err
	}
}

func mustNumbers(lhs, rhs Value) (float64, float64, error) {
	l, ok := lhs.NumberValue()
	if !ok {
		return 0, 0, fmt.Errorf("left operand is not a number")
	}
	r, ok := rhs.NumberValue()
	if !ok {
		return 0, 0, fmt.Errorf("right operand is not a number")
	}
	return l, r, nil
}

func equalValues(a, b Value) bool {
	if a.Kind() != b.Kind() {
		an, aok := a.NumberValue()
		bn, bok := b.NumberValue()
		return aok && bok && an == bn
	}

	switch a.Kind() {
	case ValueUndefined, ValueNull:
		return true
	case ValueBool:
		av, _ := a.BoolValue()
		bv, _ := b.BoolValue()
		return av == bv
	case ValueNumber:
		av, _ := a.NumberValue()
		bv, _ := b.NumberValue()
		return av == bv
	case ValueString:
		av, _ := a.StringValue()
		bv, _ := b.StringValue()
		return av == bv
	case ValueTrinary:
		av, _ := a.TrinaryValue()
		bv, _ := b.TrinaryValue()
		return av == bv
	case ValueList:
		al, _ := a.ListValue()
		bl, _ := b.ListValue()
		if len(al) != len(bl) {
			return false
		}
		for i := range al {
			if !equalValues(al[i], bl[i]) {
				return false
			}
		}
		return true
	case ValueMap:
		am, _ := a.MapValue()
		bm, _ := b.MapValue()
		if len(am) != len(bm) {
			return false
		}
		for k, av := range am {
			bv, ok := bm[k]
			if !ok || !equalValues(av, bv) {
				return false
			}
		}
		return true
	case ValueObject:
		return a.ref == b.ref
	default:
		return false
	}
}

func matchesValue(haystack, pattern Value) (bool, error) {
	h, ok := haystack.StringValue()
	if !ok {
		return false, fmt.Errorf("haystack must be a string")
	}
	p, ok := pattern.StringValue()
	if !ok {
		return false, fmt.Errorf("pattern must be a string")
	}
	return regexp.MatchString(p, h)
}

func containsValue(haystack, needle Value) bool {
	switch haystack.Kind() {
	case ValueString:
		h, _ := haystack.StringValue()
		n, ok := needle.StringValue()
		return ok && n != "" && strings.Contains(h, n)
	case ValueList:
		xs, _ := haystack.ListValue()
		for _, v := range xs {
			if equalValues(v, needle) {
				return true
			}
		}
		return false
	case ValueMap:
		m, _ := haystack.MapValue()
		if s, ok := needle.StringValue(); ok {
			_, ok2 := m[s]
			return ok2
		}
		if sub, ok := needle.MapValue(); ok {
			for k, v := range sub {
				mv, ok2 := m[k]
				if !ok2 {
					return false
				}
				if !equalValues(v, mv) {
					return false
				}
			}
			return true
		}
		for _, v := range m {
			if equalValues(v, needle) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
