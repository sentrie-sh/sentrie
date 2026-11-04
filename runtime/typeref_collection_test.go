// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

func (r *RuntimeTestSuite) TestValidateAgainstListTypeRef() {
	typeRef := ast.NewListTypeRef(ast.NewStringTypeRef(stubRange()), stubRange())

	r.Run("rejects non-array inputs", func() {
		err := validateAgainstListTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.String("not-an-array"), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "is not an array")
	})

	r.Run("rejects array item with invalid type", func() {
		err := validateAgainstListTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"ok", 2.0}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "item is not valid")
	})

	r.Run("accepts valid list item types", func() {
		err := validateAgainstListTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"a", "b"}), typeRef, stubRange())
		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstMapTypeRef() {
	typeRef := ast.NewDictTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange())

	r.Run("rejects non-dict values", func() {
		err := validateAgainstDictTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"x"}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "is not a dict")
	})

	r.Run("rejects dict values of the wrong element type", func() {
		err := validateAgainstDictTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(map[string]any{"x": "any-value", "y": 2.0}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "value for key \"x\" is not valid")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstRecordTypeRef() {
	typeRef := ast.NewRecordTypeRef([]ast.TypeRef{
		ast.NewStringTypeRef(stubRange()),
		ast.NewNumberTypeRef(stubRange()),
	}, stubRange())

	r.Run("rejects non-array value", func() {
		err := validateAgainstRecordTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(map[string]any{"x": "y"}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "is not a record")
	})

	r.Run("rejects field length mismatch", func() {
		err := validateAgainstRecordTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"name"}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "fields length mismatch")
	})

	r.Run("rejects invalid field type", func() {
		err := validateAgainstRecordTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"name", "bad"}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "not a valid record field")
	})

	r.Run("accepts valid record values", func() {
		err := validateAgainstRecordTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny([]any{"name", 7.0}), typeRef, stubRange())
		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstShapeTypeRef() {
	shapeRef := ast.NewFQN([]string{"app", "UserShape"}, stubRange())
	typeRef := ast.NewShapeTypeRef(shapeRef.Ptr(), stubRange())

	newPolicy := func() *index.Policy {
		return &index.Policy{
			Shapes: map[string]*index.Shape{},
			Namespace: &index.Namespace{
				Shapes: map[string]*index.Shape{},
			},
		}
	}

	r.Run("returns shape not found when shape is missing", func() {
		err := validateAgainstShapeTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, newPolicy(), box.FromAny(map[string]any{}), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "shape 'app/UserShape' not found")
	})

	r.Run("validates via alias typeref when shape is alias", func() {
		p := newPolicy()
		p.Shapes["app/UserShape"] = &index.Shape{
			AliasOf: ast.NewStringTypeRef(stubRange()),
		}
		r.NoError(validateAgainstShapeTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, p, box.String("alice"), typeRef, stubRange()))
	})

	r.Run("rejects non-dict for complex shapes", func() {
		p := newPolicy()
		p.Shapes["app/UserShape"] = &index.Shape{
			Model: &index.ShapeModel{
				Fields: map[string]*index.ShapeModelField{
					"name": {Name: "name", Optional: false, TypeRef: ast.NewStringTypeRef(stubRange())},
				},
			},
		}
		err := validateAgainstShapeTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, p, box.String("bad"), typeRef, stubRange())
		r.Error(err)
		r.Contains(err.Error(), "is not a shape")
	})
}
