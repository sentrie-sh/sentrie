// SPDX-License-Identifier: Apache-2.0

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

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

func validateAgainstShapeTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.ShapeTypeRef, pos tokens.Range) error {
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
	vm, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("value %v is not a shape at %s - expected shape", v, pos)
	}

	// check the fields
	for _, field := range shape.Model.Fields {
		// if not nullable, the field MUST exist and MUST NOT be null
		// if optional, the field MAY exist and MAY be null

		// if required, the field MUST exist
		if _, ok := vm[field.Name]; !ok && field.Required {
			return errors.Errorf("field %s is required at %s - expected field", field.Name, pos)
		}

		if field.NotNullable && vm[field.Name] == nil {
			return errors.Errorf("field %s cannot be null at %s - expected field", field.Name, pos)
		}

		value := vm[field.Name]
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, value, field.TypeRef, pos); err != nil {
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
		checker, ok := constraints.ShapeContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, v.(map[string]any), args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}
