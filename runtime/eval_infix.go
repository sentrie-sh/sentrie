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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
	"github.com/binaek/sentra/trinary"
	"github.com/pkg/errors"
)

func isString(v any) bool {
	switch v.(type) {
	case string:
		return true
	}
	return false
}

func evalInfix(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, in *ast.InfixExpression) (any, *trace.Node, error) {
	node, done := trace.New("infix", in.Operator, in, map[string]any{})
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
		out := num(l) / num(r)
		return out, node.SetResult(out), nil
	case "%":
		out := float64(int64(num(l)) % int64(num(r)))
		return out, node.SetResult(out), nil

	case "==":
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

	case "in", "contains":
		out := contains(r, l)
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
