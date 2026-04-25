// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
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
