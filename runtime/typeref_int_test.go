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
	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstIntTypeRef() {
	typeRef := &ast.IntTypeRef{
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
			value:         "not an int",
			expectError:   true,
			expectedError: "value not an int is not an int at :2:2 - expected int",
		},
		{
			name:          "should return an error if the value is a float64",
			value:         float64(123.45),
			expectError:   true,
			expectedError: "value 123.45 is not an int at :2:2 - expected int",
		},
		{
			name:          "should return an error if the value is a bool",
			value:         true,
			expectError:   true,
			expectedError: "value true is not an int at :2:2 - expected int",
		},
		{
			name:          "should return an error if the value is a string number",
			value:         "123",
			expectError:   true,
			expectedError: "value 123 is not an int at :2:2 - expected int",
		},
		{
			name:        "should not return an error if the value is a positive int64",
			value:       int64(123),
			expectError: false,
		},
		{
			name:        "should not return an error if the value is a negative int64",
			value:       int64(-123),
			expectError: false,
		},
		{
			name:        "should not return an error if the value is zero",
			value:       int64(0),
			expectError: false,
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstIntTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Equal(tt.expectedError, err.Error())
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstIntTypeRefWithConstraints() {
	// Test positive constraint (no arguments)
	typeRef := &ast.IntTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a positive constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "positive",
		Args: []ast.Expression{},
	}
	_ = typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when value is positive",
			value:       int64(15),
			expectError: false,
		},
		{
			name:        "should pass when value is a small positive number",
			value:       int64(1),
			expectError: false,
		},
		{
			name:          "should fail when value is zero",
			value:         int64(0),
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when value is negative",
			value:         int64(-5),
			expectError:   true,
			expectedError: "constraint failed",
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstIntTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstIntTypeRefEdgeCases() {
	// Test even constraint (no arguments)
	typeRef := &ast.IntTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an even constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "even",
		Args: []ast.Expression{},
	}
	_ = typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when value is even",
			value:       int64(4),
			expectError: false,
		},
		{
			name:        "should pass when value is zero (even)",
			value:       int64(0),
			expectError: false,
		},
		{
			name:        "should pass when value is a large even number",
			value:       int64(1000000),
			expectError: false,
		},
		{
			name:        "should pass when value is a negative even number",
			value:       int64(-4),
			expectError: false,
		},
		{
			name:          "should fail when value is odd",
			value:         int64(3),
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when value is a large odd number",
			value:         int64(1000001),
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when value is a negative odd number",
			value:         int64(-3),
			expectError:   true,
			expectedError: "constraint failed",
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			// Create a mock expression for the test
			mockExpr := &ast.Identifier{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			}
			err := validateAgainstIntTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}
