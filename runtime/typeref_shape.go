// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"
	"fmt"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/xerr"
	"github.com/pkg/errors"
)

func validateAgainstShapeTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.ShapeTypeRef) error {
	var shape *index.Shape

	shapeFqn := typeRef.Ref.String()

	// look for the shape in the policy
	if s, ok := p.Shapes[shapeFqn]; ok {
		shape = s
	}

	if shape == nil {
		if len(typeRef.Ref) < 2 {
			// this is a problem - we can't have a shape without a namespace and name
			return xerr.ErrShapeNotFound(shapeFqn)
		}
		// if the fqn is a fully qualified name, try to resolve it
		shapeName, ns := typeRef.Ref[len(typeRef.Ref)-1], typeRef.Ref[:len(typeRef.Ref)-1]
		s, err := exec.Index().ResolveShape(ns.String(), shapeName)
		if err != nil {
			return err
		}
		shape = s
	}

	// if we still don't have a shape, return an error
	if shape == nil {
		return xerr.ErrShapeNotFound(shapeFqn)
	}

	// a simple shape is an alias to another typeref
	if shape.Simple != nil {
		return validateValueAgainstTypeRef(ctx, ec, exec, p, v, shape.Simple)
	}

	// at this point, we know it's a complex shape
	// so we need to validate the value against the complex shape
	vm, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("value %v is not a shape", v)
	}

	// check the fields
	for _, field := range shape.Complex.Fields {
		// if not nullable, the field MUST exist and MUST NOT be null
		// if optional, the field MAY exist and MAY be null

		if !field.Optional {
			if _, ok := vm[field.Name]; !ok {
				return errors.Errorf("field %s is required", field.Name)
			}
		}

		if field.NotNullable && !field.Optional && vm[field.Name] == nil {
			return errors.Errorf("field %s cannot be null", field.Name)
		}

		value := vm[field.Name]
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, value, field.TypeRef); err != nil {
			return errors.Wrapf(err, "field '%s' is not valid", field.Name)
		}
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]any, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return err
			}
			args[i] = csArg
		}
		if _, ok := shapeContraintCheckers[constraint.Name]; !ok {
			return errors.Errorf("unknown constraint: %s applied to shape at %s", constraint.Name, typeRef.Position())
		}

		if err := shapeContraintCheckers[constraint.Name](ctx, p, v.(map[string]any), args); err != nil {
			return errors.Wrapf(err, "constraint is not valid")
		}
	}

	return nil
}

var shapeContraintCheckers map[string]constraintChecker[map[string]any] = map[string]constraintChecker[map[string]any]{}
