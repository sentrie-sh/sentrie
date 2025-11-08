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

package constraints

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

// TrinaryConstraintCheckers contains supported boolean constraint validators.
// For trinaries, common constraints include eq/neq (compare to true/false/unknown),
// and truthiness helpers like is_true/is_false.
var TrinaryConstraintCheckers map[string]ConstraintDefinition[trinary.Value] = map[string]ConstraintDefinition[trinary.Value]{
	"not_unknown": {
		Name:    "not_unknown",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val trinary.Value, args []any) error {
			if val == trinary.Unknown {
				return fmt.Errorf("value is unknown")
			}
			return nil
		},
	},
	"eq": {
		Name:    "eq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val trinary.Value, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("eq constraint requires 1 argument")
			}
			expected, ok := args[0].(trinary.Value)
			if !ok {
				return fmt.Errorf("eq constraint expects a boolean argument")
			}
			if val != expected {
				return fmt.Errorf("value %v is not equal to %v", val, expected)
			}
			return nil
		},
	},
	"neq": {
		Name:    "neq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val trinary.Value, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("neq constraint requires 1 argument")
			}
			expected, ok := args[0].(trinary.Value)
			if !ok {
				return fmt.Errorf("neq constraint expects a boolean argument")
			}
			if val == expected {
				return fmt.Errorf("value %v is equal to %v - expected not equal", val, expected)
			}
			return nil
		},
	},
	"is_true": {
		Name:    "is_true",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val trinary.Value, args []any) error {
			if val != trinary.True {
				return fmt.Errorf("value %v is not true", val)
			}
			return nil
		},
	},
	"is_false": {
		Name:    "is_false",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val trinary.Value, args []any) error {
			if val != trinary.False {
				return fmt.Errorf("value %v is not false", val)
			}
			return nil
		},
	},
}
