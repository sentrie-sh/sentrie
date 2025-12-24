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
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstDocumentTypeRef() {
	typeRef := ast.NewDocumentTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a string", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, "not a document", typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value not a document is not a document")
	})

	r.Run("should return an error if the value is an int64", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, int64(123), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value 123 is not a document")
	})

	r.Run("should return an error if the value is a float64", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, float64(123.45), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value 123.45 is not a document")
	})

	r.Run("should return an error if the value is a bool", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, true, typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value true is not a document")
	})

	r.Run("should return an error if the value is a list", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, []interface{}{"item1", "item2"}, typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value [item1 item2] is not a document")
	})

	r.Run("should return an error if the value is nil", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, nil, typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value <nil> is not a document")
	})

	r.Run("should not return an error if the value is an empty document", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, map[string]interface{}{}, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a document with string values", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   "30",
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a document with mixed types", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"name":    "John Doe",
			"age":     int64(30),
			"active":  true,
			"score":   float64(95.5),
			"tags":    []interface{}{"admin", "user"},
			"profile": map[string]interface{}{"bio": "Software engineer"},
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a document with nested documents", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John Doe",
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "New York",
				},
			},
			"metadata": map[string]interface{}{
				"created_at": "2023-01-01",
				"updated_at": "2023-12-31",
			},
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a document with array values", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"items":   []interface{}{"item1", "item2", "item3"},
			"numbers": []interface{}{int64(1), int64(2), int64(3)},
			"mixed":   []interface{}{"string", int64(42), true},
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstDocumentTypeRefEdgeCases() {
	typeRef := ast.NewDocumentTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a map with non-string keys", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, map[int]interface{}{1: "value"}, typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "value map[1:value] is not a document")
	})

	r.Run("should handle document with empty string keys", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"":        "empty key",
			"normal":  "normal key",
			"another": "another key",
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should handle document with special characters in keys", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"key-with-dash":       "value1",
			"key_with_underscore": "value2",
			"key.with.dots":       "value3",
			"key with spaces":     "value4",
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should handle document with unicode keys and values", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := map[string]interface{}{
			"姓名":    "张三",
			"email": "zhang@example.com",
			"café":  "français",
			"emoji": "🚀",
		}
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should handle very large document", func() {
		// Create a mock execution context and executor
		ec := &ExecutionContext{}
		exec := &executorImpl{}
		p := &index.Policy{}

		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		value := func() map[string]interface{} {
			doc := make(map[string]interface{})
			for i := 0; i < 1000; i++ {
				doc[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
			}
			return doc
		}()
		err := validateAgainstDocumentTypeRef(r.T().Context(), ec, exec, p, value, typeRef, mockExpr.Span())

		r.NoError(err)
	})
}
