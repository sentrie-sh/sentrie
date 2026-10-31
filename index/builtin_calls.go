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

// builtinCallVisitor receives every call that resolves to a native builtin,
// with the scope it was resolved in (checkBuiltinCalls needs it to resolve
// argument kinds; other visitors ignore it).
type builtinCallVisitor func(k *kindCheckCtx, decl *builtins.Decl, call *ast.CallExpression)

// forEachBuiltinCall visits every call to a native builtin in the index: fact
// defaults, policy lets, rule default, when and body slots, export attachments,
// `import decision ... with x as <expr>` clauses, and derive bodies.
//
// Callees resolve exactly as the runtime's getTarget resolves them
// (isBuiltinCall): a value binding first, then a visible derive, then a
// builtin. A let or lambda parameter named like a builtin shadows it, so
// `(now) => { yield now() }` calls the parameter, not the builtin `now`, and
// is not reported here (#161).
func (idx *Index) forEachBuiltinCall(ctx context.Context, visit builtinCallVisitor) error {
	return idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		if call, ok := e.(*ast.CallExpression); ok {
			if decl, ok := k.isBuiltinCall(call); ok {
				visit(k, decl, call)
			}
		}
	})
}

// scopedExprVisitor receives an expression with the lexical scope it is
// evaluated in.
type scopedExprVisitor func(k *kindCheckCtx, e ast.Expression)

// forEachScopedExpr visits, in pre-order, every expression the index holds:
// fact defaults, policy lets, rule default, when and body slots, export
// attachments, `import decision ... with x as <expr>` clauses, and derive
// bodies. Each expression comes with the scope it is evaluated in.
func (idx *Index) forEachScopedExpr(ctx context.Context, visit scopedExprVisitor) error {
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
		if err := cancelled(); err != nil {
			return err
		}
		for _, policy := range ns.Policies {
			if err := cancelled(); err != nil {
				return err
			}
			k := newRuleKindCheckCtx(idx, policy)
			for _, fact := range policy.Facts {
				k.walkScopedExprs(fact.Default, visit)
			}
			for _, let := range policy.Lets {
				k.walkScopedExprs(let.Value, visit)
			}
			for _, rule := range policy.Rules {
				k.walkScopedExprs(rule.Default, visit)
				k.walkScopedExprs(rule.When, visit)
				k.walkScopedExprs(rule.Body, visit)
			}
			for _, export := range policy.RuleExports {
				if export == nil {
					continue
				}
				for _, attachment := range export.Attachments {
					if attachment != nil {
						k.walkScopedExprs(attachment.Value, visit)
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
			newDeriveKindCheckCtx(idx, d).walkScopedExprs(d.Lambda.Body, visit)
		}
	}
	return nil
}

// walkScopedExprs walks e in pre-order, tracking the local bindings: a block
// binds all its lets before walking their initializers, as the runtime does,
// and a lambda pushes its parameters. visit runs on a node before its children are walked, so a
// visitor that rewrites a node (resolveIsExpressions) is walked as rewritten.
func (k *kindCheckCtx) walkScopedExprs(e ast.Expression, visit scopedExprVisitor) {
	if e == nil {
		return
	}

	visit(k, e)

	switch n := e.(type) {
	case *ast.BlockExpression:
		child := *k
		child.scope = cloneBindingScope(k.scope)
		// The runtime injects every let of a block before evaluating any of
		// them (evalBlock, lazily), so each initializer sees all of the
		// block's lets, itself and later ones included: a let lambda can call
		// itself or a later let (#161). Bind the names first, then refine each
		// binding's kind once its initializer has been walked.
		for _, st := range n.Statements {
			if vd, ok := st.(*ast.VarDeclaration); ok {
				child.scope[vd.Name] = bindingFromLet(child.idx, child.policy, vd)
			}
		}
		for _, st := range n.Statements {
			vd, ok := st.(*ast.VarDeclaration)
			if !ok {
				continue
			}
			child.walkScopedExprs(vd.Value, visit)
			child.bindLet(vd)
		}
		child.walkScopedExprs(n.Yield, visit)
	case *ast.LambdaExpression:
		k.pushLambdaScope(n).walkScopedExprs(n.Body, visit)
	case *ast.ImportClause:
		// A rule body of the form `import decision r from p with f as <expr>`.
		// Each with expression is evaluated in the importing rule's context
		// (runtime ImportDecision), so it is walked in this scope.
		// forEachDeriveExprChild has no case for import clauses, since a derive
		// cannot contain one.
		for _, w := range n.Withs {
			if w != nil {
				k.walkScopedExprs(w, visit)
			}
		}
	case *ast.WithClause:
		if n != nil {
			k.walkScopedExprs(n.Expr, visit)
		}
	default:
		_ = forEachDeriveExprChild(e, func(child ast.Expression) error {
			k.walkScopedExprs(child, visit)
			return nil
		})
	}
}
