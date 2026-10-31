// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

func validateAgainstShapeTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ShapeTypeRef, pos tokens.Range) error {
	shape, err := resolveShapeTypeRef(exec, p, typeRef)
	if err != nil {
		return err
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

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]box.Value, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return errCheckFailed(err)
			}
			args[i] = csArg
		}
		checker, ok := constraints.ShapeContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, v, args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
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
