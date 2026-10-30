// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

// TrinaryConstraintCheckers contains supported boolean constraint validators.
// For trinaries, common constraints include eq/neq (compare to true/false/unknown),
// and truthiness helpers like is_true/is_false.
var TrinaryConstraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"not_unknown": {
		Name:    "not_unknown",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			tv, ok := val.TrinaryValue()
			if !ok {
				return fmt.Errorf("expected trinary, got %s", val.Kind())
			}
			if tv == trinary.Unknown {
				return fmt.Errorf("value is unknown")
			}
			return nil
		},
	},
	"eq": {
		Name:    "eq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			tv, ok := val.TrinaryValue()
			if !ok {
				return fmt.Errorf("expected trinary, got %s", val.Kind())
			}
			if len(args) != 1 {
				return invalidArgf("eq constraint requires 1 argument")
			}
			expected, oka := args[0].TrinaryValue()
			if !oka {
				return invalidArgf("eq constraint expects a boolean argument")
			}
			if tv != expected {
				return fmt.Errorf("value %v is not equal to %v", tv, expected)
			}
			return nil
		},
	},
	"neq": {
		Name:    "neq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			tv, ok := val.TrinaryValue()
			if !ok {
				return fmt.Errorf("expected trinary, got %s", val.Kind())
			}
			if len(args) != 1 {
				return invalidArgf("neq constraint requires 1 argument")
			}
			expected, oka := args[0].TrinaryValue()
			if !oka {
				return invalidArgf("neq constraint expects a boolean argument")
			}
			if tv == expected {
				return fmt.Errorf("value %v is equal to %v - expected not equal", tv, expected)
			}
			return nil
		},
	},
	"is_true": {
		Name:    "is_true",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			tv, ok := val.TrinaryValue()
			if !ok {
				return fmt.Errorf("expected trinary, got %s", val.Kind())
			}
			if tv != trinary.True {
				return fmt.Errorf("value %v is not true", tv)
			}
			return nil
		},
	},
	"is_false": {
		Name:    "is_false",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			tv, ok := val.TrinaryValue()
			if !ok {
				return fmt.Errorf("expected trinary, got %s", val.Kind())
			}
			if tv != trinary.False {
				return fmt.Errorf("value %v is not false", tv)
			}
			return nil
		},
	},
}
