// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalTernary(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.TernaryExpression) (box.Value, *trace.Node, error) {
	ctx, n, done := trace.New(ctx, t, "ternary", map[string]any{"elvis": t.Elvis})
	defer done()

	if t.Elvis {
		lhs, child, err := eval(ctx, ec, exec, p, t.Condition)
		n.Attach(child)
		if err != nil {
			return box.Undefined(), n.SetErr(err), err
		}
		if lhs.IsUndefined() || lhs.IsNull() {
			v, child2, err2 := eval(ctx, ec, exec, p, t.ElseBranch)
			n.Attach(child2)
			if err2 != nil {
				return box.Undefined(), n.SetErr(err2), err2
			}
			return v, n.SetResult(v), nil
		}
		return lhs, n.SetResult(lhs), nil
	}

	c, cn, err := eval(ctx, ec, exec, p, t.Condition)
	n.Attach(cn)
	if err != nil {
		return box.Undefined(), n.SetErr(err), err
	}
	if box.TrinaryFrom(c).IsTrue() {
		v, tn, err := eval(ctx, ec, exec, p, t.ThenBranch)
		n.Attach(tn)
		n.SetResult(v)
		return v, n, err
	}
	v, en, err := eval(ctx, ec, exec, p, t.ElseBranch)
	n.Attach(en)
	n.SetResult(v)
	return v, n, err
}
