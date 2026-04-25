// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestValidateValueAgainstTypeRefNormalizesBoxedValue() {
	typeRef := ast.NewStringTypeRef(stubRange())
	err := validateValueAgainstTypeRef(
		s.T().Context(),
		&ExecutionContext{},
		&executorImpl{},
		&index.Policy{},
		box.String("hello"),
		typeRef,
		tokens.Range{File: "test.sentra"},
	)
	s.Require().NoError(err)
}

func (s *RuntimeTestSuite) TestValidateValueAgainstTypeRefNullableAndNilTypeRefBranches() {
	nullable := ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange())
	err := validateValueAgainstTypeRef(
		s.T().Context(),
		&ExecutionContext{},
		&executorImpl{},
		&index.Policy{},
		box.Null(),
		nullable,
		tokens.Range{File: "test.sentra"},
	)
	s.Require().NoError(err)

	err = validateValueAgainstTypeRef(
		s.T().Context(),
		&ExecutionContext{},
		&executorImpl{},
		&index.Policy{},
		box.String("anything"),
		nil,
		tokens.Range{File: "test.sentra"},
	)
	s.Require().NoError(err)
}

func (s *RuntimeTestSuite) TestValidateValueAgainstTypeRefDispatchesToAllKinds() {
	shapeRef := ast.NewFQN([]string{"app", "AliasShape"}, stubRange())
	policy := &index.Policy{
		Shapes: map[string]*index.Shape{
			"app/AliasShape": {AliasOf: ast.NewStringTypeRef(stubRange())},
		},
		Namespace: &index.Namespace{Shapes: map[string]*index.Shape{}},
	}
	testCases := []struct {
		name    string
		value   box.Value
		typeRef ast.TypeRef
	}{
		{name: "string", value: box.String("ok"), typeRef: ast.NewStringTypeRef(stubRange())},
		{name: "trinary", value: box.Trinary(trinary.True), typeRef: ast.NewTrinaryTypeRef(stubRange())},
		{name: "number", value: box.Number(1), typeRef: ast.NewNumberTypeRef(stubRange())},
		{name: "list", value: box.FromAny([]any{"a"}), typeRef: ast.NewListTypeRef(ast.NewStringTypeRef(stubRange()), stubRange())},
		{name: "dict", value: box.FromAny(map[string]any{"x": "y"}), typeRef: ast.NewDictTypeRef(ast.NewStringTypeRef(stubRange()), stubRange())},
		{name: "shape", value: box.String("alice"), typeRef: ast.NewShapeTypeRef(shapeRef.Ptr(), stubRange())},
		{name: "document", value: box.FromAny(map[string]any{"x": "y"}), typeRef: ast.NewDocumentTypeRef(stubRange())},
		{name: "record", value: box.FromAny([]any{"name"}), typeRef: ast.NewRecordTypeRef([]ast.TypeRef{ast.NewStringTypeRef(stubRange())}, stubRange())},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := validateValueAgainstTypeRef(s.T().Context(), &ExecutionContext{}, &executorImpl{}, policy, tc.value, tc.typeRef, stubRange())
			s.Require().NoError(err)
		})
	}
}
