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

// visitBuiltinCalls walks e, tracking the bindings that shadow builtins: a
// block binds each let after walking its initializer, and a lambda pushes its
// parameters.
func (k *kindCheckCtx) visitBuiltinCalls(e ast.Expression, visit builtinCallVisitor) {
	if e == nil {
		return
	}

	switch n := e.(type) {
	case *ast.CallExpression:
		if decl, ok := k.runtimeBuiltinCallee(n); ok {
			visit(k, decl, n)
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
	case *ast.ImportClause:
		// A rule body of the form `import decision r from p with f as <expr>`.
		// Each with expression is evaluated in the importing rule's context
		// (runtime ImportDecision), so it is checked in this scope.
		// forEachDeriveExprChild has no case for import clauses, since a derive
		// cannot contain one.
		for _, w := range n.Withs {
			if w != nil {
				k.visitBuiltinCalls(w, visit)
			}
		}
	case *ast.WithClause:
		if n != nil {
			k.visitBuiltinCalls(n.Expr, visit)
		}
	default:
		_ = forEachDeriveExprChild(e, func(child ast.Expression) error {
			k.visitBuiltinCalls(child, visit)
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
