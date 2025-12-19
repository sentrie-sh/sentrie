// SPDX-License-Identifier: Apache-2.0

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
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

func isString(v any) bool {
	switch v.(type) {
	case string:
		return true
	}
	return false
}

func evalInfix(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, in, "infix", map[string]any{
		"operator": in.Operator,
	})
	defer done()

	l, ln, err := eval(ctx, ec, exec, p, in.Left)
	node.Attach(ln)
	if err != nil {
		return nil, node.SetErr(err), err
	}
	r, rn, err := eval(ctx, ec, exec, p, in.Right)
	node.Attach(rn)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	// If any of the operands is undefined, return undefined
	if IsUndefined(l) || IsUndefined(r) {
		return Undefined, node.SetResult(Undefined), nil
	}

	switch in.Operator {
	case "+":
		// If any of the operands is a string, we coerce the other operand to a string and concatenate
		if isString(l) || isString(r) {
			out := fmt.Sprintf("%v%v", l, r)
			return out, node.SetResult(out), nil
		}
		out := num(l) + num(r)
		return out, node.SetResult(out), nil
	case "-":
		out := num(l) - num(r)
		return out, node.SetResult(out), nil
	case "*":
		out := num(l) * num(r)
		return out, node.SetResult(out), nil
	case "/":
		lnum := num(l)
		rnum := num(r)
		if rnum == 0 {
			return 0, node.SetErr(fmt.Errorf("divide by zero")), nil
		}
		out := lnum / rnum
		return out, node.SetResult(out), nil
	case "%":
		lnum := num(l)
		rnum := num(r)
		if rnum == 0 {
			return 0, node.SetErr(fmt.Errorf("divide by zero")), nil
		}
		out := float64(int64(lnum) % int64(rnum))
		return out, node.SetResult(out), nil

	case "==", "is":
		out := equals(l, r)
		return out, node.SetResult(out), nil
	case "!=":
		out := !equals(l, r)
		return out, node.SetResult(out), nil
	case "<":
		out := num(l) < num(r)
		return out, node.SetResult(out), nil
	case "<=":
		out := num(l) <= num(r)
		return out, node.SetResult(out), nil
	case ">":
		out := num(l) > num(r)
		return out, node.SetResult(out), nil
	case ">=":
		out := num(l) >= num(r)
		return out, node.SetResult(out), nil

	case "and":
		out := and(l, r)
		return out, node.SetResult(out), nil

	case "or":
		out := or(l, r)
		return out, node.SetResult(out), nil

	case "xor":
		out := xor(l, r)
		return out, node.SetResult(out), nil

	case "in":
		out := contains(r, l)
		return out, node.SetResult(out), nil

	case "contains":
		out := contains(l, r)
		return out, node.SetResult(out), nil

	case "matches":
		out, err := matches(l, r)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		return out, node.SetResult(out), nil

	default:
		err := fmt.Errorf("unsupported infix op: %s", in.Operator)
		return nil, node.SetErr(err), err
	}
}

func xor(l, r any) trinary.Value {
	left := trinary.From(l)
	right := trinary.From(r)
	// XOR = (A OR B) AND NOT (A AND B)
	return left.Or(right).And(left.And(right).Not())
}

func and(l, r any) trinary.Value {
	left := trinary.From(l)
	right := trinary.From(r)
	return left.And(right)
}

func or(l, r any) trinary.Value {
	left := trinary.From(l)
	right := trinary.From(r)
	return left.Or(right)
}

func num(v any) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case float64:
		return t
	default:
		return 0
	}
}

func equals(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case int64, float64:
		return num(a) == num(b)
	}
	return a == b
}

func matches(haystack, pattern any) (bool, error) {
	if h, ok := haystack.(string); ok {
		if p, ok := pattern.(string); ok {
			return regexp.MatchString(p, h)
		}
	}
	return false, errors.New("haystack and pattern must be strings")
}

func contains(haystack, needle any) bool {
	switch h := haystack.(type) {
	case string:
		if s, ok := needle.(string); ok {
			return s != "" && strings.Contains(h, s)
		}
		return false
	case []any:
		for _, v := range h {
			if equals(v, needle) {
				return true
			}
		}
		return false
	case map[string]any:
		if s, ok := needle.(string); ok {
			_, ok2 := h[s]
			return ok2
		}
		if s, ok := needle.(map[string]any); ok {
			for k, v := range s {
				_, ok2 := h[k]
				if !ok2 {
					return false
				}
				if !equals(v, h[k]) {
					return false
				}
			}
			return true
		}
		return false
	default:
		return false
	}
}
