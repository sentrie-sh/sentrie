// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

// typeRefSite is the scope a type ref's shape names resolve in, as LookupShape
// takes it: the policy's shapes, then the namespace's. policy is nil at
// namespace scope (a namespace shape or a namespace-level derive).
type typeRefSite struct {
	policy *Policy
	ns     *Namespace
	// isType marks the type of `x is T`, whose shape names
	// resolveIsExpressions has already resolved and reported.
	isType bool
}

// typeRefVisitor receives one type ref with the site it appears at.
type typeRefVisitor func(site typeRefSite, t ast.TypeRef)

// forEachTypeRef visits every type ref the index holds: facts, policy lets,
// declared rule types, shape aliases and fields (once per statement), derive parameter and return
// types, and inside expressions block lets, lambda parameter and return types,
// casts and `x is T`. Each top-level type ref is expanded by
// forEachNestedTypeRef, so a visitor sees list and set element, dict value and
// record field types too, and never a nullable wrapper.
//
// A shape's alias and field types resolve in the shape's own scope, a derive's
// signature and body in the derive's (its policy, or its namespace for a
// namespace-level derive), and everything else in the policy it is declared
// in.
func (idx *Index) forEachTypeRef(ctx context.Context, visit typeRefVisitor) error {
	each := func(site typeRefSite, t ast.TypeRef) {
		forEachNestedTypeRef(t, func(n ast.TypeRef) { visit(site, n) })
	}
	lambda := func(site typeRefSite, lam *ast.LambdaExpression) {
		for _, t := range lam.ParamTypes {
			each(site, t)
		}
		each(site, lam.ReturnType)
	}

	shapes := map[*ast.ShapeStatement]struct{}{}
	shape := func(s *Shape) {
		if s == nil || s.Statement == nil {
			return
		}
		if _, seen := shapes[s.Statement]; seen {
			return
		}
		shapes[s.Statement] = struct{}{}
		site := typeRefSite{policy: s.Policy, ns: s.Namespace}
		each(site, s.Statement.Simple)
		if s.Statement.Complex != nil {
			for _, f := range s.Statement.Complex.Fields {
				if f != nil {
					each(site, f.Type)
				}
			}
		}
	}

	for _, ns := range idx.Namespaces {
		for _, s := range ns.Shapes {
			shape(s)
		}
		for _, policy := range ns.Policies {
			for _, s := range policy.Shapes {
				shape(s)
			}
			site := typeRefSite{policy: policy, ns: policy.Namespace}
			for _, fact := range policy.Facts {
				each(site, fact.Type)
			}
			for _, let := range policy.Lets {
				each(site, let.Type)
			}
			for _, rule := range policy.Rules {
				if rule.Node != nil {
					each(site, rule.Node.Type)
				}
			}
		}
	}
	for _, d := range idx.DerivesByFQN {
		if d.Lambda != nil {
			lambda(typeRefSite{policy: d.Policy, ns: d.Namespace}, d.Lambda)
		}
	}

	return idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		site := typeRefSite{policy: k.policy, ns: k.shapeNamespace()}
		switch n := e.(type) {
		case *ast.CastExpression:
			each(site, n.TargetType)
		case *ast.IsExpression:
			site.isType = true
			each(site, n.Type)
		case *ast.LambdaExpression:
			lambda(site, n)
		case *ast.BlockExpression:
			for _, st := range n.Statements {
				if vd, ok := st.(*ast.VarDeclaration); ok {
					each(site, vd.Type)
				}
			}
		}
	})
}

// forEachNestedTypeRef visits t and, in pre-order, every type ref nested in
// it: list and set element, dict value, record field and union member types.
// A nullable is not visited itself; its inner type is.
func forEachNestedTypeRef(t ast.TypeRef, visit func(ast.TypeRef)) {
	switch n := t.(type) {
	case nil:
		return
	case *ast.NullableTypeRef:
		forEachNestedTypeRef(n.Inner, visit)
		return
	}
	visit(t)
	switch n := t.(type) {
	case *ast.ListTypeRef:
		forEachNestedTypeRef(n.ElemType, visit)
	case *ast.SetTypeRef:
		forEachNestedTypeRef(n.ElemType, visit)
	case *ast.DictTypeRef:
		forEachNestedTypeRef(n.ValueType, visit)
	case *ast.RecordTypeRef:
		for _, f := range n.Fields {
			forEachNestedTypeRef(f, visit)
		}
	case *ast.UnionTypeRef:
		for _, m := range n.Members {
			forEachNestedTypeRef(m, visit)
		}
	}
}
