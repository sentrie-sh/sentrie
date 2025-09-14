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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
)

func evalUnary(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, u *ast.UnaryExpression) (any, *trace.Node, error) {
	node, done := trace.New("unary", u.Operator, u, map[string]any{})
	defer done()

	v, child, err := eval(ctx, ec, exec, p, u.Right)
	node.Attach(child)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	switch u.Operator {
	case "not":
		fallthrough
	case "!":
		out := not(v)
		return out, node.SetResult(out), nil
	case "-":
		switch x := v.(type) {
		case int64:
			out := -x
			return out, node.SetResult(out), nil
		case int:
			out := -int(x)
			return out, node.SetResult(out), nil
		case float64:
			out := -x
			return out, node.SetResult(out), nil
		default:
			err := fmt.Errorf("bad operand for unary -: %T", v)
			return nil, node.SetErr(err), err
		}
	default:
		err := fmt.Errorf("unsupported unary op: %s", u.Operator)
		return nil, node.SetErr(err), err
	}
}

func not(v any) bool {
	return !IsTruthy(v)
}
