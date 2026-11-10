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

	shape, err := resolveShapeTypeRef(exec, p, typeRef)
	if err != nil {
		return err
	}

	// The shape's alias and field types resolve where the shape is declared
	// (its policy, or its namespace), however it was reached: the caller's
	// policy would miss, or silently pick, a different shape of the same name
	// when the shape comes from another namespace (#281). The constraints on
	// the reference itself, below, stay in the caller's scope.
	scope := shape.ShapeScope()
	if scope == nil {
		scope = p
	}

	// a simple shape is an alias to another typeref
	if shape.AliasOf != nil {
		return validateValueAgainstTypeRef(ctx, ec, exec, scope, v, shape.AliasOf, pos)
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

		// A field given undefined (missing data) passes through, as an
		// untyped value does (#148). An absent required field still fails.
		if fieldValue.IsUndefined() {
			continue
		}

		// A field composed in through `with` resolves in its declaring shape.
		if err := validateValueAgainstTypeRef(ctx, ec, exec, field.TypeScope(scope), fieldValue, field.TypeRef, pos); err != nil {
			return fmt.Errorf("field '%s' is not valid: %w", field.Name, err)
		}
	}

	return checkConstraints(ctx, v, bound, pos)
}

// resolveShapeTypeRef finds the shape typeRef names: in the policy, which
// overrides the namespace, then in the policy's namespace, then for a name of
// three or more segments as an exported shape of the namespace it names.
func resolveShapeTypeRef(exec Executor, p *index.Policy, typeRef *ast.ShapeTypeRef) (*index.Shape, error) {
	shapeFqn := typeRef.Ref.String()

	if shape, ok := p.Shapes[shapeFqn]; ok {
		return shape, nil
	}
	if shape, ok := p.Namespace.Shapes[shapeFqn]; ok {
		return shape, nil
	}

	if len(typeRef.Ref.Parts) > 2 {
		ns := typeRef.Ref.Parent()
		name := typeRef.Ref.LastSegment()

		namespace, err := exec.Index().ResolveNamespace(ns.String())
		if err != nil {
			return nil, err
		}
		if namespace == nil {
			return nil, xerr.ErrNamespaceNotFound(ns.String())
		}
		if err := namespace.VerifyShapeExported(name); err != nil {
			return nil, err
		}

		shape, err := exec.Index().ResolveShape(ns.String(), name)
		if err != nil {
			return nil, err
		}
		if shape != nil {
			return shape, nil
		}
	}

	return nil, xerr.ErrShapeNotFound(fmt.Sprintf("shape '%s' not found at %s", shapeFqn, typeRef.Span()))
}
