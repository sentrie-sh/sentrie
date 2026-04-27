// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

func (r *RuntimeTestSuite) TestValidateAgainstShapeTypeRef_FieldPresenceAndNullabilityMatrix() {
	shapeRef := ast.NewFQN([]string{"app", "UserShape"}, stubRange())
	typeRef := ast.NewShapeTypeRef(shapeRef.Ptr(), stubRange())

	newPolicy := func(field index.ShapeModelField) *index.Policy {
		return &index.Policy{
			Shapes: map[string]*index.Shape{
				"app/UserShape": {
					Model: &index.ShapeModel{
						Fields: map[string]*index.ShapeModelField{
							"name": {
								Name:     field.Name,
								Optional: field.Optional,
								TypeRef:  field.TypeRef,
							},
						},
					},
				},
			},
			Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
		}
	}

	cases := []struct {
		name      string
		field     index.ShapeModelField
		input     box.Value
		expectErr bool
	}{
		{
			name: "required_non_null_absent_is_invalid",
			field: index.ShapeModelField{
				Name: "name", Optional: false, TypeRef: ast.NewStringTypeRef(stubRange()),
			},
			input: box.FromAny(map[string]any{}), expectErr: true,
		},
		{
			name: "required_non_null_null_is_invalid",
			field: index.ShapeModelField{
				Name: "name", Optional: false, TypeRef: ast.NewStringTypeRef(stubRange()),
			},
			input: box.FromAny(map[string]any{"name": nil}), expectErr: true,
		},
		{
			name: "optional_non_null_absent_is_valid",
			field: index.ShapeModelField{
				Name: "name", Optional: true, TypeRef: ast.NewStringTypeRef(stubRange()),
			},
			input: box.FromAny(map[string]any{}), expectErr: false,
		},
		{
			name: "optional_non_null_null_is_invalid",
			field: index.ShapeModelField{
				Name: "name", Optional: true, TypeRef: ast.NewStringTypeRef(stubRange()),
			},
			input: box.FromAny(map[string]any{"name": nil}), expectErr: true,
		},
		{
			name: "required_nullable_null_is_valid",
			field: index.ShapeModelField{
				Name: "name", Optional: false, TypeRef: ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
			},
			input: box.FromAny(map[string]any{"name": nil}), expectErr: false,
		},
		{
			name: "required_nullable_absent_is_invalid",
			field: index.ShapeModelField{
				Name: "name", Optional: false, TypeRef: ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
			},
			input: box.FromAny(map[string]any{}), expectErr: true,
		},
		{
			name: "optional_nullable_absent_is_valid",
			field: index.ShapeModelField{
				Name: "name", Optional: true, TypeRef: ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
			},
			input: box.FromAny(map[string]any{}), expectErr: false,
		},
		{
			name: "optional_nullable_null_is_valid",
			field: index.ShapeModelField{
				Name: "name", Optional: true, TypeRef: ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
			},
			input: box.FromAny(map[string]any{"name": nil}), expectErr: false,
		},
		{
			name: "present_undefined_is_invalid",
			field: index.ShapeModelField{
				Name: "name", Optional: true, TypeRef: ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
			},
			input: box.FromAny(map[string]box.Value{"name": box.Undefined()}), expectErr: true,
		},
	}

	for _, tc := range cases {
		r.Run(tc.name, func() {
			err := validateAgainstShapeTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, newPolicy(tc.field), tc.input, typeRef, stubRange())
			if tc.expectErr {
				r.Error(err)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (s *RuntimeTestSuite) TestValidateAgainstShapeTypeRefFieldErrorBranches() {
	typeRef := ast.NewShapeTypeRef(ast.NewFQN([]string{"UserShape"}, stubRange()).Ptr(), stubRange())
	policy := &index.Policy{
		Shapes: map[string]*index.Shape{
			"UserShape": {
				Model: &index.ShapeModel{
					Fields: map[string]*index.ShapeModelField{
						"name": {Name: "name", Optional: false, TypeRef: ast.NewStringTypeRef(stubRange())},
					},
				},
			},
		},
		Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
	}

	err := validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field name is required")

	policy.Shapes["UserShape"] = &index.Shape{
		Model: &index.ShapeModel{
			Fields: map[string]*index.ShapeModelField{
				"name": {Name: "name", TypeRef: ast.NewStringTypeRef(stubRange())},
			},
		},
	}
	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{"name": nil}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field 'name' is not valid")

	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]box.Value{"name": box.Undefined()}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot be undefined")

	policy.Shapes["UserShape"] = &index.Shape{
		Model: &index.ShapeModel{
			Fields: map[string]*index.ShapeModelField{
				"age": {Name: "age", TypeRef: ast.NewNumberTypeRef(stubRange())},
			},
		},
	}
	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.FromAny(map[string]any{"age": "bad"}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "field 'age' is not valid")
}

func (s *RuntimeTestSuite) TestValidateAgainstShapeTypeRefGlobalResolutionBranches() {
	typeRef := ast.NewShapeTypeRef(ast.NewFQN([]string{"ext", "models", "User"}, stubRange()).Ptr(), stubRange())
	policy := &index.Policy{
		Shapes:    map[string]*index.Shape{},
		Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
	}
	idx := index.CreateIndex()
	exec := &executorImpl{index: idx}

	err := validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, exec, policy, box.FromAny(map[string]any{}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")

	nsFQN := ast.NewFQN([]string{"ext", "models"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{"User": {Model: &index.ShapeModel{Fields: map[string]*index.ShapeModelField{}}}},
		ShapeExports: map[string]*index.ExportedShape{},
		Children:     []*index.Namespace{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, exec, policy, box.FromAny(map[string]any{}), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "is not exported")

	ns.ShapeExports["User"] = &index.ExportedShape{Name: "User"}
	err = validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, exec, policy, box.FromAny(map[string]any{}), typeRef, stubRange())
	s.Require().NoError(err)
}

func (s *RuntimeTestSuite) TestValidateAgainstShapeTypeRefConstraintBranches() {
	typeRef := ast.NewShapeTypeRef(ast.NewFQN([]string{"UserShape"}, stubRange()).Ptr(), stubRange())
	policy := &index.Policy{
		Shapes: map[string]*index.Shape{
			"UserShape": {
				Model: &index.ShapeModel{
					Fields: map[string]*index.ShapeModelField{},
				},
			},
		},
		Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
	}

	err := validateAgainstShapeTypeRef(context.Background(), &ExecutionContext{}, &executorImpl{}, policy, box.Number(1), typeRef, stubRange())
	s.Require().Error(err)
	s.Contains(err.Error(), "is not a shape")
}
