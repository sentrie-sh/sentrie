// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

// checkShapeRefs resolves every shape named in a type ref (#282): facts,
// policy lets, shape aliases and fields, derive signatures, block lets, lambda
// signatures and casts, and the element, value and field types nested in
// them. Each name resolves as the runtime's validateAgainstShapeTypeRef
// resolves it (LookupShape), in the scope forEachTypeRef gives the site, so a
// shape that does not exist, or one another namespace does not export, is an
// index error instead of a runtime one; a typed let that is never evaluated
// used to never report it at all.
//
// The types of `x is T` are skipped: resolveIsExpressions has already
// resolved and reported their shape names. Errors are reported in source
// order.
func (idx *Index) checkShapeRefs(ctx context.Context) error {
	var errs []siteError
	err := idx.forEachTypeRef(ctx, func(site typeRefSite, t ast.TypeRef) {
		ref, ok := t.(*ast.ShapeTypeRef)
		if !ok || ref.Ref == nil || site.isType {
			return
		}
		if siteErr := idx.shapeRefError(site, ref); siteErr != nil {
			errs = append(errs, *siteErr)
		}
	})
	if err != nil {
		return err
	}
	return joinSiteErrors(errs)
}

// shapeRefError reports ref when no shape of its name is visible from site,
// in the text the runtime and shape composition use for a missing shape.
func (idx *Index) shapeRefError(site typeRefSite, ref *ast.ShapeTypeRef) *siteError {
	shape, err := idx.LookupShape(site.policy, site.ns, ref.Ref)
	if err == nil && shape != nil {
		return nil
	}
	at := ref.Span()
	verb := "not found"
	if errors.Is(err, xerr.NotExportedError{}) {
		verb = "not exported"
	}
	return &siteError{at: at, err: fmt.Errorf("shape '%s' %s at %s: %w", ref.Ref.String(), verb, at.String(), xerr.ErrIndex)}
}
