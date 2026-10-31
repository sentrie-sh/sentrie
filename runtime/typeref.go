// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

// typeCheckScope is the policy whose shapes a type check resolves against:
// the evaluating policy, or inside a derive body the derive's own policy, or
// for a namespace-level derive a scope holding only the derive's namespace
// (index.Derive.ShapeScope). Every type check in a derive body (casts, typed
// lets, `is`, and the types of a lambda written there) resolves here, as the
// index does; the calling policy would fail on, or silently pick, a different
// shape when the derive is called from another namespace or from a policy
// that declares a shape of the same name.
//
// ec.evalDerive is the lexically enclosing derive, so a lambda written in a
// policy and applied inside a derive still resolves in that policy.
func typeCheckScope(ec *ExecutionContext, p *index.Policy) *index.Policy {
	if d := ec.evalDerive; d != nil {
		return d.ShapeScope()
	}
	return p
}

func validateValueAgainstTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef ast.TypeRef, valueRange tokens.Range) error {
	if ast.IsNullableTypeRef(typeRef) {
		if v.IsNull() {
			return nil
		}
		typeRef = ast.UnwrapNullableTypeRef(typeRef)
	}

	switch t := typeRef.(type) {
	case *ast.StringTypeRef:
		return validateAgainstStringTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.TrinaryTypeRef:
		return validateAgainstTrinaryTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.NumberTypeRef:
		return validateAgainstNumberTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.ListTypeRef:
		return validateAgainstListTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.DictTypeRef:
		return validateAgainstDictTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.ShapeTypeRef:
		return validateAgainstShapeTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.DocumentTypeRef:
		return validateAgainstDocumentTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.RecordTypeRef:
		return validateAgainstRecordTypeRef(ctx, ec, exec, p, v, t, valueRange)
	}

	return nil
}
