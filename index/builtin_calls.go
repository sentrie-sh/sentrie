// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/xerr"
)

// builtinCallVisitor receives every call that resolves to a native builtin.
type builtinCallVisitor func(decl *builtins.Decl, call *ast.CallExpression)

// forEachBuiltinCall visits every call to a native builtin in the index: fact
// defaults, policy lets, rule default, when and body slots, export attachments
// and derive bodies. It resolves callees with the same scoping as
// checkBuiltinCalls (local binding > derive > builtin, matching the runtime's
// getTarget), so a name shadowed by a let, a parameter or a derive is not
// reported as a builtin.
func (idx *Index) forEachBuiltinCall(ctx context.Context, visit builtinCallVisitor) error {
	cancelled := func() error {
		if ctx.Err() != nil {
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		return nil
	}
	if err := cancelled(); err != nil {
		return err
	}

	for _, ns := range idx.Namespaces {
		for _, policy := range ns.Policies {
			if err := cancelled(); err != nil {
				return err
			}
			k := newRuleKindCheckCtx(idx, policy)
			for _, fact := range policy.Facts {
				k.visitBuiltinCalls(fact.Default, visit)
			}
			for _, let := range policy.Lets {
				k.visitBuiltinCalls(let.Value, visit)
			}
			for _, rule := range policy.Rules {
				k.visitBuiltinCalls(rule.Default, visit)
				k.visitBuiltinCalls(rule.When, visit)
				k.visitBuiltinCalls(rule.Body, visit)
			}
			for _, export := range policy.RuleExports {
				if export == nil {
					continue
				}
				for _, attachment := range export.Attachments {
					if attachment != nil {
						k.visitBuiltinCalls(attachment.Value, visit)
					}
				}
			}
		}
	}

	for _, d := range idx.DerivesByFQN {
		if err := cancelled(); err != nil {
			return err
		}
		if d.Lambda != nil && d.Lambda.Body != nil {
			newDeriveKindCheckCtx(idx, d).visitBuiltinCalls(d.Lambda.Body, visit)
		}
	}
	return nil
}

// visitBuiltinCalls walks e with the scoping rules of walkExpr.
func (k *kindCheckCtx) visitBuiltinCalls(e ast.Expression, visit builtinCallVisitor) {
	if e == nil {
		return
	}

	switch n := e.(type) {
	case *ast.CallExpression:
		if decl, ok := k.isBuiltinCall(n); ok {
			visit(decl, n)
		}
		k.visitBuiltinCalls(n.Callee, visit)
		for _, a := range n.Arguments {
			k.visitBuiltinCalls(a, visit)
		}
	case *ast.BlockExpression:
		child := *k
		child.scope = cloneBindingScope(k.scope)
		for _, st := range n.Statements {
			vd, ok := st.(*ast.VarDeclaration)
			if !ok {
				continue
			}
			child.visitBuiltinCalls(vd.Value, visit)
			child.bindLet(vd)
		}
		child.visitBuiltinCalls(n.Yield, visit)
	case *ast.LambdaExpression:
		k.pushLambdaScope(n).visitBuiltinCalls(n.Body, visit)
	default:
		_ = forEachDeriveExprChild(e, func(child ast.Expression) error {
			k.visitBuiltinCalls(child, visit)
			return nil
		})
	}
}
