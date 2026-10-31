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
// Callees resolve exactly as the runtime's getTarget resolves them: a visible
// derive first, then a builtin. getTarget never resolves a call through a
// local binding, so a let or lambda parameter named like a builtin does not
// shadow it: `(now) => { yield now() }` still calls the builtin `now` at
// runtime and is reported here. (checkBuiltinCalls' isBuiltinCall still
// treats locals as shadowing, which is a separate, pre-existing difference.)
func (idx *Index) forEachBuiltinCall(ctx context.Context, visit builtinCallVisitor) error {
	return idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		if call, ok := e.(*ast.CallExpression); ok {
			if decl, ok := k.runtimeBuiltinCallee(call); ok {
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
				for _, c := range rule.Contributions {
					k.walkScopedExprs(c.When, visit)
					k.walkScopedExprs(c.Body, visit)
				}
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
// binds each let after walking its initializer, and a lambda pushes its
// parameters. visit runs on a node before its children are walked, so a
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

// runtimeBuiltinCallee reports the builtin a call dispatches to at runtime,
// mirroring getTarget in runtime/eval_call.go: an identifier callee that does
// not name a visible derive (DefineShort inside a derive, otherwise the
// policy's then the namespace's derives) resolves to builtins.Table. Local
// bindings are deliberately not consulted, because getTarget does not.
func (k *kindCheckCtx) runtimeBuiltinCallee(c *ast.CallExpression) (*builtins.Decl, bool) {
	id, ok := c.Callee.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	if k.lookupDerive(id.Value) != nil {
		return nil, false
	}
	decl, ok := builtins.Table[id.Value]
	return decl, ok
}
