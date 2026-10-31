// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

// validateAgainstAnyTypeRef accepts every defined value, null included.
// Undefined is the absence of a value, which no type accepts.
func validateAgainstAnyTypeRef(v box.Value, pos tokens.Range) error {
	if v.IsUndefined() {
		return errMismatch("value is undefined at %s - expected any", pos)
	}
	return nil
}

// validateAgainstUnionTypeRef passes when v validates against any member.
// Otherwise, a member whose check could not be completed (not a mismatch, such
// as an unknown shape) is reported first, in member order, so a broken check
// is never read as "does not conform". When every member is a mismatch, the
// failure reported is the closest member's: the first member, in source
// order, whose outer kind accepts v (see outerKindAccepts), so a value that is
// a list of the wrong items reports the list member's item error. When no
// member's outer kind accepts v, the union itself is reported.
func validateAgainstUnionTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.UnionTypeRef, pos tokens.Range) error {
	errs := make([]error, len(typeRef.Members))
	for i, member := range typeRef.Members {
		err := validateValueAgainstTypeRef(ctx, ec, exec, p, v, member, pos)
		if err == nil {
			return nil
		}
		errs[i] = err
	}

	for _, err := range errs {
		if !isTypeMismatch(err) {
			return err
		}
	}

	for i, member := range typeRef.Members {
		if outerKindAccepts(exec, p, v, member) {
			return &typeRefFailure{
				err:      fmt.Errorf("value %v does not match %s at %s: closest member %s: %w", v, typeRef, pos, member, errs[i]),
				mismatch: true,
			}
		}
	}

	return errMismatch("value %v is not %s at %s - expected one of its members", v, typeRef, pos)
}

// outerKindAccepts reports whether v passes t's outermost kind check, before
// any element, field or constraint is looked at. A shape alias is read
// through to its target; a shape that cannot be resolved accepts nothing.
func outerKindAccepts(exec Executor, p *index.Policy, v box.Value, t ast.TypeRef) bool {
	switch tr := t.(type) {
	case *ast.AnyTypeRef:
		return !v.IsUndefined()
	case *ast.NullableTypeRef:
		return v.IsNull() || outerKindAccepts(exec, p, v, tr.Inner)
	case *ast.UnionTypeRef:
		for _, m := range tr.Members {
			if outerKindAccepts(exec, p, v, m) {
				return true
			}
		}
		return false
	case *ast.StringTypeRef:
		_, ok := v.StringValue()
		return ok
	case *ast.NumberTypeRef:
		_, ok := v.NumberValue()
		return ok
	case *ast.TrinaryTypeRef:
		if _, ok := v.TrinaryValue(); ok {
			return true
		}
		_, ok := v.DecisionState()
		return ok
	case *ast.ListTypeRef, *ast.RecordTypeRef:
		_, ok := v.ListValue()
		return ok
	case *ast.DictTypeRef, *ast.DocumentTypeRef:
		_, ok := v.DictValue()
		return ok
	case *ast.ShapeTypeRef:
		shape, err := resolveShapeTypeRef(exec, p, tr)
		if err != nil || shape == nil {
			return false
		}
		if shape.AliasOf != nil {
			return outerKindAccepts(exec, p, v, shape.AliasOf)
		}
		_, ok := v.DictValue()
		return ok
	}
	return false
}
