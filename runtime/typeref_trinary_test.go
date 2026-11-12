// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (r *RuntimeTestSuite) TestValidateAgainstBoolTypeRef() {
	typeRef := ast.NewTrinaryTypeRef(tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny("not a bool"), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value 'not a bool' is not a trinary at %s - expected true, false or unknown", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(int64(123)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a trinary at %s - expected true, false or unknown", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is a float64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(123)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a trinary at %s - expected true, false or unknown", mockExpr.Span()), err.Error())
	})

	r.Run("should return an error if the value is a string number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny("123"), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal(fmt.Sprintf("value '123' is not a trinary at %s - expected true, false or unknown", mockExpr.Span()), err.Error())
	})

	r.Run("should not return an error if the value is boolean true", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(true), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is boolean false", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(false), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary unknown", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(trinary.Unknown), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary true", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(trinary.True), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is trinary false", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(trinary.False), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("boolean rejects unknown and accepts true and false", func() {
		boolRef := ast.NewTrinaryTypeRef(typeRef.Span())
		boolRef.Boolean = true
		at := typeRef.Span()
		err := validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.Trinary(trinary.Unknown), boolRef, at)
		r.Require().Error(err)
		r.Equal(fmt.Sprintf("value 'unknown' is not a boolean at %s - expected true or false; use trinary to allow unknown", at), err.Error())
		r.NoError(validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.Bool(true), boolRef, at))
		r.NoError(validateAgainstTrinaryTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.Bool(false), boolRef, at))
	})
}
