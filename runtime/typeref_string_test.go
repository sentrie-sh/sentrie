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
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRef() {
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, int64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 123 is not a string", err.Error())
	})

	r.Run("should return an error if the value is a float64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(123.45), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 123.45 is not a string", err.Error())
	})

	r.Run("should return an error if the value is a bool", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, true, typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value true is not a string", err.Error())
	})

	r.Run("should not return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is an empty string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a long string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "this is a very long string with many characters", typeRef, mockExpr.Span())

		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefLengthConstraint() {
	// Test length constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a length constraint
	constraint := ast.NewTypeRefConstraint(
		"length",
		[]ast.Expression{
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string has exact length", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is too short", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hi", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is too long", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefMinLengthConstraint() {
	// Test minlength constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a minlength constraint
	constraint := ast.NewTypeRefConstraint(
		"minlength",
		[]ast.Expression{ast.NewIntegerLiteral(3, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string meets minimum length", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string equals minimum length", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "abc", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is too short", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hi", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefMaxLengthConstraint() {
	// Test maxlength constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a maxlength constraint
	constraint := ast.NewTypeRefConstraint(
		"maxlength",
		[]ast.Expression{ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is within maximum length", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string equals maximum length", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "abc", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is too long", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefRegexpConstraint() {
	// Test regexp constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a regexp constraint
	constraint := ast.NewTypeRefConstraint(
		"regexp",
		[]ast.Expression{ast.NewStringLiteral(`^[a-zA-Z0-9]+$`, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string matches pattern", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is only letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is only numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains special characters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello-world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string contains spaces", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefStartsWithConstraint() {
	// Test starts_with constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a starts_with constraint
	constraint := ast.NewTypeRefConstraint(
		"starts_with",
		[]ast.Expression{ast.NewStringLiteral("hello", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string starts with prefix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string equals prefix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string does not start with prefix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "world hello", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is shorter than prefix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hi", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefEndsWithConstraint() {
	// Test ends_with constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add an ends_with constraint
	constraint := ast.NewTypeRefConstraint(
		"ends_with",
		[]ast.Expression{ast.NewStringLiteral("world", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string ends with suffix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string equals suffix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string does not end with suffix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello there", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is shorter than suffix", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hi", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefHasSubstringConstraint() {
	// Test has_substring constraint
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a has_substring constraint
	constraint := ast.NewTypeRefConstraint(
		"has_substring",
		[]ast.Expression{ast.NewStringLiteral("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string contains substring", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "this is a test string", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string equals substring", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "test", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when substring is at the beginning", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "testing something", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when substring is at the end", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "something test", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string does not contain substring", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefEmailConstraint() {
	// Test email constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add an email constraint
	constraint := ast.NewTypeRefConstraint("email", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is a valid email", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "user@example.com", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid email with subdomain", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "user@mail.example.com", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid email with special characters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "user.name+tag@example.co.uk", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is not a valid email", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not-an-email", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is missing @ symbol", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "userexample.com", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUrlConstraint() {
	// Test url constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a url constraint
	constraint := ast.NewTypeRefConstraint("url", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is a valid HTTP URL", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "http://example.com", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid HTTPS URL", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "https://example.com", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid URL with path", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "https://example.com/path/to/page", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is not a valid URL", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not-a-url", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is missing protocol", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "example.com", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUuidConstraint() {
	// Test uuid constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a uuid constraint
	constraint := ast.NewTypeRefConstraint("uuid", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is a valid UUID", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "550e8400-e29b-41d4-a716-446655440000", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid UUID without dashes", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "550e8400e29b41d4a716446655440000", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is not a valid UUID", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not-a-uuid", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is too short", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefAlphanumericConstraint() {
	// Test alphanumeric constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add an alphanumeric constraint
	constraint := ast.NewTypeRefConstraint("alphanumeric", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string contains only letters and numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains special characters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello-world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string contains spaces", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefAlphaConstraint() {
	// Test alpha constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add an alpha constraint
	constraint := ast.NewTypeRefConstraint("alpha", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string contains only letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only uppercase letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "HELLO", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains mixed case letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "Hello", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello123", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string contains special characters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello-world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefNumericConstraint() {
	// Test numeric constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a numeric constraint
	constraint := ast.NewTypeRefConstraint("numeric", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is a valid integer", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid float", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123.45", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is a valid negative number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "-123.45", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello123", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is not numeric", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not-numeric", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefLowercaseConstraint() {
	// Test lowercase constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a lowercase constraint
	constraint := ast.NewTypeRefConstraint("lowercase", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is lowercase", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains uppercase letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "Hello World", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is all uppercase", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "HELLO", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUppercaseConstraint() {
	// Test uppercase constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add an uppercase constraint
	constraint := ast.NewTypeRefConstraint("uppercase", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is uppercase", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "HELLO WORLD", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only numbers", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string contains lowercase letters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "Hello World", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is all lowercase", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefTrimmedConstraint() {
	// Test trimmed constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a trimmed constraint
	constraint := ast.NewTypeRefConstraint("trimmed", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string has no leading or trailing whitespace", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string has leading whitespace", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, " hello world", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string has trailing whitespace", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world ", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string has both leading and trailing whitespace", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, " hello world ", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefNotEmptyConstraint() {
	// Test not_empty constraint (no arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a not_empty constraint
	constraint := ast.NewTypeRefConstraint("not_empty", []ast.Expression{}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is not empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "hello world", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains only spaces", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "   ", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string contains special characters", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "!@#$%", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefOneOfConstraint() {
	// Test one_of constraint (variable arguments)
	typeRef := ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a one_of constraint
	constraint := ast.NewTypeRefConstraint(
		"one_of",
		[]ast.Expression{
			ast.NewStringLiteral("red", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("green", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("blue", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when string is one of the allowed values", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "red", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is another allowed value", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "green", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when string is the third allowed value", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "blue", typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when string is not one of the allowed values", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "yellow", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when string is empty", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}
