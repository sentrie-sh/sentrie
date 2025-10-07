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
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstDocumentTypeRef() {
	typeRef := &ast.DocumentTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:          "should return an error if the value is a string",
			value:         "not a document",
			expectError:   true,
			expectedError: "value not a document is not a document",
		},
		{
			name:          "should return an error if the value is an int64",
			value:         int64(123),
			expectError:   true,
			expectedError: "value 123 is not a document",
		},
		{
			name:          "should return an error if the value is a float64",
			value:         float64(123.45),
			expectError:   true,
			expectedError: "value 123.45 is not a document",
		},
		{
			name:          "should return an error if the value is a bool",
			value:         true,
			expectError:   true,
			expectedError: "value true is not a document",
		},
		{
			name:          "should return an error if the value is a list",
			value:         []interface{}{"item1", "item2"},
			expectError:   true,
			expectedError: "value [item1 item2] is not a document",
		},
		{
			name:          "should return an error if the value is nil",
			value:         nil,
			expectError:   true,
			expectedError: "value <nil> is not a document",
		},
		{
			name:        "should not return an error if the value is an empty document",
			value:       map[string]interface{}{},
			expectError: false,
		},
		{
			name: "should not return an error if the value is a document with string values",
			value: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   "30",
			},
			expectError: false,
		},
		{
			name: "should not return an error if the value is a document with mixed types",
			value: map[string]interface{}{
				"name":    "John Doe",
				"age":     int64(30),
				"active":  true,
				"score":   float64(95.5),
				"tags":    []interface{}{"admin", "user"},
				"profile": map[string]interface{}{"bio": "Software engineer"},
			},
			expectError: false,
		},
		{
			name: "should not return an error if the value is a document with nested documents",
			value: map[string]interface{}{
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
			},
			expectError: false,
		},
		{
			name: "should not return an error if the value is a document with array values",
			value: map[string]interface{}{
				"items":   []interface{}{"item1", "item2", "item3"},
				"numbers": []interface{}{int64(1), int64(2), int64(3)},
				"mixed":   []interface{}{"string", int64(42), true},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock execution context and executor
			ec := &ExecutionContext{}
			exec := &executorImpl{}
			p := &index.Policy{}

			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstDocumentTypeRef(context.Background(), ec, exec, p, tt.value, typeRef, mockExpr.Position())

			if tt.expectError {
				r.Error(err)
				if tt.expectedError != "" {
					r.Contains(err.Error(), tt.expectedError)
				}
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstDocumentTypeRefWithConstraints() {
	// Test with constraints (even though none are currently implemented)
	typeRef := &ast.DocumentTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a constraint to test constraint handling
	constraint := &ast.TypeRefConstraint{
		Name: "minlength",
		Args: []ast.Expression{
			&ast.IntegerLiteral{Value: 1},
		},
	}
	_ = typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:          "should return an error for unknown constraint",
			value:         map[string]interface{}{"key": "value"},
			expectError:   true,
			expectedError: "unknown constraint: 'minlength' at :1:1: unknown constraint: typeref error",
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock execution context and executor
			ec := &ExecutionContext{}
			exec := &executorImpl{}
			p := &index.Policy{}

			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstDocumentTypeRef(context.Background(), ec, exec, p, tt.value, typeRef, mockExpr.Position())

			if tt.expectError {
				r.Error(err)
				if tt.expectedError != "" {
					r.Contains(err.Error(), tt.expectedError)
				}
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstDocumentTypeRefEdgeCases() {
	typeRef := &ast.DocumentTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:          "should return an error if the value is a map with non-string keys",
			value:         map[int]interface{}{1: "value"},
			expectError:   true,
			expectedError: "value map[1:value] is not a document",
		},
		{
			name: "should handle document with empty string keys",
			value: map[string]interface{}{
				"":        "empty key",
				"normal":  "normal key",
				"another": "another key",
			},
			expectError: false,
		},
		{
			name: "should handle document with special characters in keys",
			value: map[string]interface{}{
				"key-with-dash":       "value1",
				"key_with_underscore": "value2",
				"key.with.dots":       "value3",
				"key with spaces":     "value4",
			},
			expectError: false,
		},
		{
			name: "should handle document with unicode keys and values",
			value: map[string]interface{}{
				"姓名":    "张三",
				"email": "zhang@example.com",
				"café":  "français",
				"emoji": "🚀",
			},
			expectError: false,
		},
		{
			name: "should handle very large document",
			value: func() map[string]interface{} {
				doc := make(map[string]interface{})
				for i := 0; i < 1000; i++ {
					doc[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
				}
				return doc
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock execution context and executor
			ec := &ExecutionContext{}
			exec := &executorImpl{}
			p := &index.Policy{}

			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstDocumentTypeRef(context.Background(), ec, exec, p, tt.value, typeRef, mockExpr.Position())

			if tt.expectError {
				r.Error(err)
				if tt.expectedError != "" {
					r.Contains(err.Error(), tt.expectedError)
				}
			} else {
				r.NoError(err)
			}
		})
	}
}
