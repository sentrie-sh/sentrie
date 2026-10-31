// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

// resolveIsExpressions settles every `x is Name` whose right-hand side the
// parser could not classify (ast.IsExpression.Ambiguous): a bare or
// slash-qualified name that may be a shape or a value. Each one is resolved in
// the scope it is evaluated in:
//
//   - the name is a shape visible from the site (LookupShape) and not a value
//     in scope: a type check, so Right is cleared;
//   - the name is a single identifier that is both a shape and a value in
//     scope: an error, since either reading would silently change meaning
//     when the other declaration is added or removed;
//   - the name is not a shape: the legacy equality `x == Name`, so Type is
//     cleared, provided every identifier in it is a value in scope; otherwise
//     an error, which is what a misspelt shape name produces;
//   - a qualified name whose namespace exists but does not hold or export the
//     shape: an error.
//
// Every shape named inside the resulting type (`list[Usr]`, `Usr?`,
// `Usr@c()`) must also be visible from the site, so a misspelt shape is always
// an index error rather than a runtime one.
//
// It must run before any walk that treats Right as an expression (cycle
// detection, derive purity, builtin checks), since a shape name is not a
// reference to a value. Errors are reported in source order.
func (idx *Index) resolveIsExpressions(ctx context.Context) error {
	var errs []isSiteError
	err := idx.forEachScopedExpr(ctx, func(k *kindCheckCtx, e ast.Expression) {
		is, ok := e.(*ast.IsExpression)
		if !ok {
			return
		}
		if is.Ambiguous() {
			if siteErr := k.resolveIs(is); siteErr != nil {
				errs = append(errs, *siteErr)
				return
			}
		}
		if is.Type != nil {
			errs = append(errs, k.checkIsTypeShapes(is.Type)...)
		}
	})
	if err != nil {
		return err
	}
	slices.SortStableFunc(errs, func(a, b isSiteError) int {
		return cmp.Or(
			cmp.Compare(a.at.File, b.at.File),
			cmp.Compare(a.at.From.Line, b.at.From.Line),
			cmp.Compare(a.at.From.Column, b.at.From.Column),
		)
	})
	joined := make([]error, len(errs))
	for i, e := range errs {
		joined[i] = e.err
	}
	return errors.Join(joined...)
}

// isSiteError is a resolution error with the span it is reported at, kept so
// the errors can be sorted into source order.
type isSiteError struct {
	at  tokens.Range
	err error
}

func (k *kindCheckCtx) resolveIs(is *ast.IsExpression) *isSiteError {
	shapeRef, ok := is.Type.(*ast.ShapeTypeRef)
	if !ok || shapeRef.Ref == nil {
		// The parser only leaves a shape reference ambiguous.
		is.Type = nil
		return nil
	}
	ref := shapeRef.Ref
	at := is.Right.Span()

	shape, err := k.idx.LookupShape(k.policy, k.shapeNamespace(), ref)
	if err != nil {
		return isResolveError(at, ref.String(), err.Error())
	}

	if shape != nil {
		if len(ref.Parts) == 1 && k.namesValue(ref.Parts[0]) {
			return isResolveError(at, ref.String(), fmt.Sprintf("%q names both a shape and a value in scope; rename one, or use == to compare values", ref.String()))
		}
		is.Right = nil
		return nil
	}

	for _, part := range ref.Parts {
		if !k.namesValue(part) {
			return isResolveError(at, ref.String(), fmt.Sprintf("%q is neither a shape nor a value in scope", part))
		}
	}
	is.Type = nil
	return nil
}

// checkIsTypeShapes reports every shape reference in t that is not visible
// from the site.
func (k *kindCheckCtx) checkIsTypeShapes(t ast.TypeRef) []isSiteError {
	switch tr := t.(type) {
	case *ast.ShapeTypeRef:
		if tr.Ref == nil {
			return nil
		}
		shape, err := k.idx.LookupShape(k.policy, k.shapeNamespace(), tr.Ref)
		if err != nil {
			return []isSiteError{*isResolveError(tr.Span(), tr.Ref.String(), err.Error())}
		}
		if shape == nil {
			return []isSiteError{*isResolveError(tr.Span(), tr.Ref.String(), fmt.Sprintf("shape %q is not visible here", tr.Ref.String()))}
		}
	case *ast.NullableTypeRef:
		return k.checkIsTypeShapes(tr.Inner)
	case *ast.ListTypeRef:
		return k.checkIsTypeShapes(tr.ElemType)
	case *ast.SetTypeRef:
		return k.checkIsTypeShapes(tr.ElemType)
	case *ast.DictTypeRef:
		return k.checkIsTypeShapes(tr.ValueType)
	case *ast.RecordTypeRef:
		var errs []isSiteError
		for _, f := range tr.Fields {
			errs = append(errs, k.checkIsTypeShapes(f)...)
		}
		return errs
	}
	return nil
}

func isResolveError(at tokens.Range, name, msg string) *isSiteError {
	return &isSiteError{at: at, err: fmt.Errorf("%s: `is %s`: %s: %w", at.String(), name, msg, xerr.ErrIndex)}
}

// shapeNamespace is the namespace whose shapes are visible at the site: the
// policy's, or for a namespace-level derive the derive's.
func (k *kindCheckCtx) shapeNamespace() *Namespace {
	if k.policy != nil {
		return k.policy.Namespace
	}
	if k.derive != nil {
		return k.derive.Namespace
	}
	return nil
}

// namesValue reports whether name resolves to a value at the site, mirroring
// runtime evalIdentifier: a local, fact, let or visible derive (the scope),
// and outside a derive body a rule of the policy.
func (k *kindCheckCtx) namesValue(name string) bool {
	if k.isShadowed(name) {
		return true
	}
	if k.derive == nil && k.policy != nil {
		if _, ok := k.policy.Rules[name]; ok {
			return true
		}
	}
	return false
}

// LookupShape resolves a shape reference the way the runtime's
// validateAgainstShapeTypeRef does: the policy's shapes, then the namespace's,
// then, for a name of three or more segments, the shape its last segment names
// in the namespace the rest names, which must export it. It returns nil and no
// error when no namespace or shape of that name is visible, and an error when
// the prefix names a namespace that lacks the shape or does not export it.
func (idx *Index) LookupShape(policy *Policy, ns *Namespace, ref *ast.FQN) (*Shape, error) {
	if ref == nil || ref.IsEmpty() {
		return nil, nil
	}
	name := ref.String()
	if policy != nil {
		if s, ok := policy.Shapes[name]; ok {
			return s, nil
		}
	}
	if ns != nil {
		if s, ok := ns.Shapes[name]; ok {
			return s, nil
		}
	}
	if len(ref.Parts) <= 2 {
		return nil, nil
	}
	target := idx.Namespaces[ref.Parent().String()]
	if target == nil {
		return nil, nil
	}
	last := ref.LastSegment()
	s, ok := target.Shapes[last]
	if !ok {
		return nil, xerr.ErrShapeNotFound(ShapeFQN(target.FQN.String(), last))
	}
	if err := target.VerifyShapeExported(last); err != nil {
		return nil, err
	}
	return s, nil
}
