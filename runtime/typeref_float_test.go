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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/tokens"
)

func (r *RuntimeTestSuite) TestValidateAgainstFloatTypeRef() {
	typeRef := &ast.FloatTypeRef{
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
			value:         "not a float",
			expectError:   true,
			expectedError: "value not a float is not a float64",
		},
		{
			name:          "should return an error if the value is an int64",
			value:         int64(123),
			expectError:   true,
			expectedError: "value 123 is not a float64",
		},
		{
			name:          "should return an error if the value is a bool",
			value:         true,
			expectError:   true,
			expectedError: "value true is not a float64",
		},
		{
			name:          "should return an error if the value is a string number",
			value:         "123.45",
			expectError:   true,
			expectedError: "value 123.45 is not a float64",
		},
		{
			name:        "should not return an error if the value is a positive float64",
			value:       float64(123.45),
			expectError: false,
		},
		{
			name:        "should not return an error if the value is a negative float64",
			value:       float64(-123.45),
			expectError: false,
		},
		{
			name:        "should not return an error if the value is zero",
			value:       float64(0),
			expectError: false,
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			err := validateAgainstFloatTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef)

			if tt.expectError {
				r.Error(err)
				r.Equal(tt.expectedError, err.Error())
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstFloatTypeRefWithConstraints() {
	// Test positive constraint (no arguments)
	typeRef := &ast.FloatTypeRef{
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
			value:       float64(15.0),
			expectError: false,
		},
		{
			name:        "should pass when value is a small positive number",
			value:       float64(0.001),
			expectError: false,
		},
		{
			name:          "should fail when value is zero",
			value:         float64(0),
			expectError:   true,
			expectedError: "constraint is not valid: value 0 is not positive",
		},
		{
			name:          "should fail when value is negative",
			value:         float64(-5.0),
			expectError:   true,
			expectedError: "constraint is not valid: value -5 is not positive",
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			err := validateAgainstFloatTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstFloatTypeRefEdgeCases() {
	// Test finite constraint
	typeRef := &ast.FloatTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a finite constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "finite",
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
			name:        "should pass when value is a normal finite number",
			value:       float64(123.45),
			expectError: false,
		},
		{
			name:        "should pass when value is zero",
			value:       float64(0),
			expectError: false,
		},
		{
			name:        "should pass when value is a very small number",
			value:       float64(1e-300),
			expectError: false,
		},
		{
			name:        "should pass when value is a very large number",
			value:       float64(1e300),
			expectError: false,
		},
		{
			name:          "should fail when value is positive infinity",
			value:         math.Inf(1),
			expectError:   true,
			expectedError: "constraint is not valid: value +Inf is not finite",
		},
		{
			name:          "should fail when value is negative infinity",
			value:         math.Inf(-1),
			expectError:   true,
			expectedError: "constraint is not valid: value -Inf is not finite",
		},
		{
			name:          "should fail when value is NaN",
			value:         math.NaN(),
			expectError:   true,
			expectedError: "constraint is not valid: value NaN is not finite",
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			err := validateAgainstFloatTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}
