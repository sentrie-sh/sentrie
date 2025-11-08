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
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (r *RuntimeTestSuite) TestValidateAgainstBoolTypeRef() {
	typeRef := ast.NewTrinaryTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not a bool", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value 'not a bool' is not a bool at %s - expected bool", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, int64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a bool at %s - expected bool", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is a float64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a bool at %s - expected bool", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is a string number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a bool at %s - expected bool", mockExpr.Span()), err.Error())
	})

	r.Run("should not return an error if the value is boolean true", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, true, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is boolean false", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, false, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary unknown", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, trinary.Unknown, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary true", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, trinary.True, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary false", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, trinary.False, typeRef, mockExpr.Span())

		r.NoError(err)
	})
}
