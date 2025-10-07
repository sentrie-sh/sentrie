package constraints

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

// BoolConstraintCheckers contains supported boolean constraint validators.
// For booleans, common constraints include eq/neq (compare to true/false),
// and truthiness helpers like is_true/is_false.
var BoolConstraintCheckers map[string]ConstraintDefinition[trinary.Value] = map[string]ConstraintDefinition[trinary.Value]{
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
