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
)

func evalUnary(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, u *ast.UnaryExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, u, "unary", map[string]any{
		"operator": u.Operator,
	})
	defer done()

	v, child, err := eval(ctx, ec, exec, p, u.Right)
	node.Attach(child)
	if err != nil {
		return box.Value{}, node.SetErr(err), err
	}

	if v.IsUndefined() {
		return box.Undefined(), node, nil
	}

	switch u.Operator {
	case "!", "not":
		out := box.Trinary(box.TrinaryFrom(v).Not())
		return out, node.SetResult(out), nil
	case "+":
		num, ok := v.NumberValue()
		if !ok {
			err := fmt.Errorf("unary + requires number")
			return box.Value{}, node.SetErr(err), err
		}
		out := box.Number(num)
		return out, node.SetResult(out), nil
	case "-":
		num, ok := v.NumberValue()
		if !ok {
			err := fmt.Errorf("unary - requires number")
			return box.Value{}, node.SetErr(err), err
		}
		out := box.Number(-num)
		return out, node.SetResult(out), nil
	default:
		err := fmt.Errorf("unsupported unary op: %s", u.Operator)
		return box.Value{}, node.SetErr(err), err
	}
}
