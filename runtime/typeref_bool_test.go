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

func (r *RuntimeTestSuite) TestValidateAgainstBoolTypeRef() {
	typeRef := &ast.BoolTypeRef{
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
			value:         "not a bool",
			expectError:   true,
			expectedError: "value 'not a bool' is not a bool",
		},
		{
			name:          "should return an error if the value is an int64",
			value:         int64(123),
			expectError:   true,
			expectedError: "value '123' is not a bool",
		},
		{
			name:          "should return an error if the value is a float64",
			value:         float64(123),
			expectError:   true,
			expectedError: "value '123' is not a bool",
		},
		{
			name:          "should return an error if the value is a string number",
			value:         "123",
			expectError:   true,
			expectedError: "value '123' is not a bool",
		},
		{
			name:        "should not return an error if the value is true",
			value:       true,
			expectError: false,
		},
		{
			name:        "should not return an error if the value is false",
			value:       false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		r.Run(tt.name, func() {
			err := validateAgainstBoolTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, tt.value, typeRef)

			if tt.expectError {
				r.Error(err)
				r.Equal(tt.expectedError, err.Error())
			} else {
				r.NoError(err)
			}
		})
	}
}
