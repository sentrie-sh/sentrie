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

func (r *RuntimeTestSuite) TestValidateAgainstBoolTypeRef() {
	typeRef := &ast.BoolTypeRef{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	}

	r.Run("should return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "not a bool", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 'not a bool' is not a bool at test.sentra:1:1 - expected bool", err.Error())
	})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, int64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value '123' is not a bool at test.sentra:1:1 - expected bool", err.Error())
	})

	r.Run("should return an error if the value is a float64", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, float64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value '123' is not a bool at test.sentra:1:1 - expected bool", err.Error())
	})

	r.Run("should return an error if the value is a string number", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, "123", typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value '123' is not a bool at test.sentra:1:1 - expected bool", err.Error())
	})

	r.Run("should not return an error if the value is true", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, true, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is false", func() {
		// Create a mock expression for the test
		mockExpr := &ast.Identifier{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Value: "test",
		}
		err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, false, typeRef, mockExpr.Span())

		r.NoError(err)
	})
}
