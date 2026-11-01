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
		if v.Kind() != box.ValueNumber {
			err := fmt.Errorf("unary + requires number")
			return box.Value{}, node.SetErr(err), err
		}
		return v, node.SetResult(v), nil
	case "-":
		if v.Kind() != box.ValueNumber {
			err := fmt.Errorf("unary - requires number")
			return box.Value{}, node.SetErr(err), err
		}
		out, err := box.Negate(v)
		if err != nil {
			return box.Value{}, node.SetErr(err), err
		}
		return out, node.SetResult(out), nil
	default:
		err := fmt.Errorf("unsupported unary op: %s", u.Operator)
		return box.Value{}, node.SetErr(err), err
	}
}
