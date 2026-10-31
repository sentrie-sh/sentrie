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
	"github.com/sentrie-sh/sentrie/xerr"
)

func validateAgainstShapeTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ShapeTypeRef, pos tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	var shape *index.Shape

	shapeFqn := typeRef.Ref.String()

	// look for the shape in the policy - this will override any shape that may have been defined in the namespace
	shape, ok := p.Shapes[shapeFqn]

	// couldn't find the shape in the policy - check if it's in the namespace of the policy
	if !ok {
		s, o := p.Namespace.Shapes[shapeFqn]
		if o {
			shape = s
		}
		ok = o
	}

	// we couldn't find the shape in the policy - go global.
	// lookup the index with the shape
	if !ok && len(typeRef.Ref.Parts) > 2 {
		ns := typeRef.Ref.Parent()
		name := typeRef.Ref.LastSegment()

		// get the namespace
		namespace, err := exec.Index().ResolveNamespace(ns.String())
		if err != nil {
			return err
		}
		if namespace == nil {
			return xerr.ErrNamespaceNotFound(ns.String())
		}
		if err := namespace.VerifyShapeExported(name); err != nil {
			return err
		}

		shape, err = exec.Index().ResolveShape(ns.String(), name)
		if err != nil {
			return err
		}
	}

	// if we still don't have a shape, return an error
	if shape == nil {
		return xerr.ErrShapeNotFound(fmt.Sprintf("shape '%s' not found at %s", shapeFqn, typeRef.Span()))
	}

	// a simple shape is an alias to another typeref
	if shape.AliasOf != nil {
		return validateValueAgainstTypeRef(ctx, ec, exec, p, v, shape.AliasOf, pos)
	}

	// at this point, we know it's a complex shape
	// so we need to validate the value against the complex shape
	vm, ok := v.DictValue()
	if !ok {
		return errMismatch("value %v is not a shape at %s - expected shape", v, pos)
	}

	// check the fields
	for _, field := range shape.Model.Fields {
		fieldValue, ok := vm[field.Name]
		if !ok {
			if field.Optional {
				continue
			}
			return errMismatch("field %s is required at %s - expected field", field.Name, pos)
		}

		if fieldValue.IsUndefined() {
			return errMismatch("field %s cannot be undefined at %s - expected field value", field.Name, pos)
		}

		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, fieldValue, field.TypeRef, pos); err != nil {
			return fmt.Errorf("field '%s' is not valid: %w", field.Name, err)
		}
	}

	return checkConstraints(ctx, v, bound, pos)
}
