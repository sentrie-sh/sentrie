// SPDX-License-Identifier: Apache-2.0
//
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
	"math"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRef() {
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not a number", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value not a number is not a number", err.Error())
	})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, int64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 123 is not a number", err.Error())
	})

	r.Run("should return an error if the value is a bool", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, true, typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value true is not a number", err.Error())
	})

	r.Run("should return an error if the value is a string number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123.45", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 123.45 is not a number", err.Error())
	})

	r.Run("should not return an error if the value is a positive number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(123.45), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a negative number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(-123.45), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(0), typeRef, mockExpr.Span())

		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRefWithConstraints() {
	// Test positive constraint (no arguments)
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a positive constraint
	constraint := ast.NewTypeRefConstraint(
		"positive",
		[]ast.Expression{},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when value is positive", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(15.0), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a small positive number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(0.001), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(0), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is negative", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(-5.0), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRefEdgeCases() {
	// Test finite constraint
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a finite constraint
	constraint := ast.NewTypeRefConstraint(
		"finite",
		[]ast.Expression{},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when value is a normal finite number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(123.45), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(0), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a very small number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(1e-300), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a very large number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(1e300), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when value is positive infinity", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, math.Inf(1), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is negative infinity", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, math.Inf(-1), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is NaN", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, math.NaN(), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}
