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

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRef() {
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:          "should return an error if the value is an int64",
			value:         int64(123),
			expectError:   true,
			expectedError: "value 123 is not a string",
		},
		{
			name:          "should return an error if the value is a float64",
			value:         float64(123.45),
			expectError:   true,
			expectedError: "value 123.45 is not a string",
		},
		{
			name:          "should return an error if the value is a bool",
			value:         true,
			expectError:   true,
			expectedError: "value true is not a string",
		},
		{
			name:        "should not return an error if the value is a string",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should not return an error if the value is an empty string",
			value:       "",
			expectError: false,
		},
		{
			name:        "should not return an error if the value is a long string",
			value:       "this is a very long string with many characters",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Equal(tt.expectedError, err.Error())
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefLengthConstraint() {
	// Test length constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a length constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "length",
		Args: []ast.Expression{
			&ast.IntegerLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: 5,
			},
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
			name:        "should pass when string has exact length",
			value:       "hello",
			expectError: false,
		},
		{
			name:          "should fail when string is too short",
			value:         "hi",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is too long",
			value:         "hello world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefMinLengthConstraint() {
	// Test minlength constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a minlength constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "minlength",
		Args: []ast.Expression{
			&ast.IntegerLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: 3,
			},
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
			name:        "should pass when string meets minimum length",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "should pass when string equals minimum length",
			value:       "abc",
			expectError: false,
		},
		{
			name:          "should fail when string is too short",
			value:         "hi",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefMaxLengthConstraint() {
	// Test maxlength constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a maxlength constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "maxlength",
		Args: []ast.Expression{
			&ast.IntegerLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: 5,
			},
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
			name:        "should pass when string is within maximum length",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "should pass when string equals maximum length",
			value:       "abc",
			expectError: false,
		},
		{
			name:          "should fail when string is too long",
			value:         "hello world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefRegexpConstraint() {
	// Test regexp constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a regexp constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "regexp",
		Args: []ast.Expression{
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: `^[a-zA-Z0-9]+$`,
			},
		},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string matches pattern",
			value:       "hello123",
			expectError: false,
		},
		{
			name:        "should pass when string is only letters",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "should pass when string is only numbers",
			value:       "123",
			expectError: false,
		},
		{
			name:          "should fail when string contains special characters",
			value:         "hello-world",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string contains spaces",
			value:         "hello world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefStartsWithConstraint() {
	// Test starts_with constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a starts_with constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "starts_with",
		Args: []ast.Expression{
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "hello",
			},
		},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string starts with prefix",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should pass when string equals prefix",
			value:       "hello",
			expectError: false,
		},
		{
			name:          "should fail when string does not start with prefix",
			value:         "world hello",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is shorter than prefix",
			value:         "hi",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefEndsWithConstraint() {
	// Test ends_with constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an ends_with constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "ends_with",
		Args: []ast.Expression{
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "world",
			},
		},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string ends with suffix",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should pass when string equals suffix",
			value:       "world",
			expectError: false,
		},
		{
			name:          "should fail when string does not end with suffix",
			value:         "hello there",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is shorter than suffix",
			value:         "hi",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefHasSubstringConstraint() {
	// Test has_substring constraint
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a has_substring constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "has_substring",
		Args: []ast.Expression{
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "test",
			},
		},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string contains substring",
			value:       "this is a test string",
			expectError: false,
		},
		{
			name:        "should pass when string equals substring",
			value:       "test",
			expectError: false,
		},
		{
			name:        "should pass when substring is at the beginning",
			value:       "testing something",
			expectError: false,
		},
		{
			name:        "should pass when substring is at the end",
			value:       "something test",
			expectError: false,
		},
		{
			name:          "should fail when string does not contain substring",
			value:         "hello world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefEmailConstraint() {
	// Test email constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an email constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "email",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is a valid email",
			value:       "user@example.com",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid email with subdomain",
			value:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid email with special characters",
			value:       "user.name+tag@example.co.uk",
			expectError: false,
		},
		{
			name:          "should fail when string is not a valid email",
			value:         "not-an-email",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is missing @ symbol",
			value:         "userexample.com",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUrlConstraint() {
	// Test url constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a url constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "url",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is a valid HTTP URL",
			value:       "http://example.com",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid HTTPS URL",
			value:       "https://example.com",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid URL with path",
			value:       "https://example.com/path/to/page",
			expectError: false,
		},
		{
			name:          "should fail when string is not a valid URL",
			value:         "not-a-url",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is missing protocol",
			value:         "example.com",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUuidConstraint() {
	// Test uuid constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a uuid constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "uuid",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is a valid UUID",
			value:       "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid UUID without dashes",
			value:       "550e8400e29b41d4a716446655440000",
			expectError: false,
		},
		{
			name:          "should fail when string is not a valid UUID",
			value:         "not-a-uuid",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is too short",
			value:         "123",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefAlphanumericConstraint() {
	// Test alphanumeric constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an alphanumeric constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "alphanumeric",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string contains only letters and numbers",
			value:       "hello123",
			expectError: false,
		},
		{
			name:        "should pass when string contains only letters",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "should pass when string contains only numbers",
			value:       "123",
			expectError: false,
		},
		{
			name:          "should fail when string contains special characters",
			value:         "hello-world",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string contains spaces",
			value:         "hello world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefAlphaConstraint() {
	// Test alpha constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an alpha constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "alpha",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string contains only letters",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "should pass when string contains only uppercase letters",
			value:       "HELLO",
			expectError: false,
		},
		{
			name:        "should pass when string contains mixed case letters",
			value:       "Hello",
			expectError: false,
		},
		{
			name:          "should fail when string contains numbers",
			value:         "hello123",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string contains special characters",
			value:         "hello-world",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefNumericConstraint() {
	// Test numeric constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a numeric constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "numeric",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is a valid integer",
			value:       "123",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid float",
			value:       "123.45",
			expectError: false,
		},
		{
			name:        "should pass when string is a valid negative number",
			value:       "-123.45",
			expectError: false,
		},
		{
			name:          "should fail when string contains letters",
			value:         "hello123",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is not numeric",
			value:         "not-numeric",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefLowercaseConstraint() {
	// Test lowercase constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a lowercase constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "lowercase",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is lowercase",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should pass when string is empty",
			value:       "",
			expectError: false,
		},
		{
			name:        "should pass when string contains only numbers",
			value:       "123",
			expectError: false,
		},
		{
			name:          "should fail when string contains uppercase letters",
			value:         "Hello World",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is all uppercase",
			value:         "HELLO",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefUppercaseConstraint() {
	// Test uppercase constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add an uppercase constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "uppercase",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is uppercase",
			value:       "HELLO WORLD",
			expectError: false,
		},
		{
			name:        "should pass when string is empty",
			value:       "",
			expectError: false,
		},
		{
			name:        "should pass when string contains only numbers",
			value:       "123",
			expectError: false,
		},
		{
			name:          "should fail when string contains lowercase letters",
			value:         "Hello World",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is all lowercase",
			value:         "hello",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefTrimmedConstraint() {
	// Test trimmed constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a trimmed constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "trimmed",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string has no leading or trailing whitespace",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should pass when string is empty",
			value:       "",
			expectError: false,
		},
		{
			name:          "should fail when string has leading whitespace",
			value:         " hello world",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string has trailing whitespace",
			value:         "hello world ",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string has both leading and trailing whitespace",
			value:         " hello world ",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				e := err.Error()
				_ = e
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefNotEmptyConstraint() {
	// Test not_empty constraint (no arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a not_empty constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "not_empty",
		Args: []ast.Expression{},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is not empty",
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "should pass when string contains only spaces",
			value:       "   ",
			expectError: false,
		},
		{
			name:        "should pass when string contains special characters",
			value:       "!@#$%",
			expectError: false,
		},
		{
			name:          "should fail when string is empty",
			value:         "",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}

func (r *RuntimeTestSuite) TestValidateAgainstStringTypeRefOneOfConstraint() {
	// Test one_of constraint (variable arguments)
	typeRef := &ast.StringTypeRef{
		Pos: tokens.Position{Line: 1, Column: 1},
	}

	// Add a one_of constraint
	constraint := &ast.TypeRefConstraint{
		Pos:  tokens.Position{Line: 1, Column: 1},
		Name: "one_of",
		Args: []ast.Expression{
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "red",
			},
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "green",
			},
			&ast.StringLiteral{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Value: "blue",
			},
		},
	}
	typeRef.AddConstraint(constraint)

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		expectedError string
	}{
		{
			name:        "should pass when string is one of the allowed values",
			value:       "red",
			expectError: false,
		},
		{
			name:        "should pass when string is another allowed value",
			value:       "green",
			expectError: false,
		},
		{
			name:        "should pass when string is the third allowed value",
			value:       "blue",
			expectError: false,
		},
		{
			name:          "should fail when string is not one of the allowed values",
			value:         "yellow",
			expectError:   true,
			expectedError: "constraint failed",
		},
		{
			name:          "should fail when string is empty",
			value:         "",
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
			err := validateAgainstStringTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef, mockExpr)

			if tt.expectError {
				r.Error(err)
				r.Contains(err.Error(), tt.expectedError)
			} else {
				r.NoError(err)
			}
		})
	}
}
