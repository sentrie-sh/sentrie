// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
)

var NumberContraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"min": define(ConstraintDefinition{
		Name:    "min",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c < 0 {
				return fmt.Errorf("value %v is not >= %v", val, args[0])
			}
			return nil
		},
	}),
	"max": define(ConstraintDefinition{
		Name:    "max",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c > 0 {
				return fmt.Errorf("value %v is not <= %v", val, args[0])
			}
			return nil
		},
	}),
	"eq": define(ConstraintDefinition{
		Name:    "eq",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !ordered || c != 0 {
				return fmt.Errorf("value %v is not equal to %v", val, args[0])
			}
			return nil
		},
	}),
	"neq": define(ConstraintDefinition{
		Name:    "neq",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c == 0 {
				return fmt.Errorf("value %v is equal to %v", val, args[0])
			}
			return nil
		},
	}),
	"gt": define(ConstraintDefinition{
		Name:    "gt",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c <= 0 {
				return fmt.Errorf("value %v is not > %v", val, args[0])
			}
			return nil
		},
	}),
	"lt": define(ConstraintDefinition{
		Name:    "lt",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c >= 0 {
				return fmt.Errorf("value %v is not < %v", val, args[0])
			}
			return nil
		},
	}),
	"in": define(ConstraintDefinition{
		Name:    "in",
		NumArgs: 1,
		Arg:     numberSetArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			if val.Kind() != box.ValueNumber {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			set, err := numberConstraintSet(args[0])
			if err != nil {
				return err
			}

			if !slices.ContainsFunc(set, func(n box.Value) bool { return box.EqualValues(val, n) }) {
				return fmt.Errorf("value %v is not in the set", val)
			}
			return nil
		},
	}),
	"not_in": define(ConstraintDefinition{
		Name:    "not_in",
		NumArgs: 1,
		Arg:     numberSetArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			if val.Kind() != box.ValueNumber {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			set, err := numberConstraintSet(args[0])
			if err != nil {
				return err
			}

			if slices.ContainsFunc(set, func(n box.Value) bool { return box.EqualValues(val, n) }) {
				return fmt.Errorf("value %v is in the set", val)
			}
			return nil
		},
	}),
	"range": define(ConstraintDefinition{
		Name:    "range",
		NumArgs: 2,
		Arg:     numberArg,
		Args: func(args []box.Value) error {
			min, max := args[0], args[1]
			if c, ordered, _ := box.CompareNumbers(min, max); ordered && c > 0 {
				return invalidArgf("range minimum %v is greater than maximum %v", min, max)
			}
			return nil
		},
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			if val.Kind() != box.ValueNumber {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			min, max := args[0], args[1]
			cMin, okMin, _ := box.CompareNumbers(val, min)
			cMax, okMax, _ := box.CompareNumbers(val, max)
			if (okMin && cMin < 0) || (okMax && cMax > 0) {
				return fmt.Errorf("value %v is not in range [%v, %v]", val, min, max)
			}
			return nil
		},
	}),
	"even": define(ConstraintDefinition{
		Name:    "even",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			rem := math.Mod(valNum, 2)
			if i, ok := val.IntegralValue(); ok {
				// Exact: float64 would drop the low bit above 2^53.
				rem = float64(i % 2)
			}
			if rem != 0 {
				return fmt.Errorf("value %v is not even", val)
			}
			return nil
		},
	}),
	"odd": define(ConstraintDefinition{
		Name:    "odd",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			rem := math.Mod(valNum, 2)
			if i, ok := val.IntegralValue(); ok {
				// Exact: float64 would drop the low bit above 2^53.
				rem = float64(i % 2)
			}
			if rem == 0 {
				return fmt.Errorf("value %v is not odd", val)
			}
			return nil
		},
	}),
	"multiple_of": define(ConstraintDefinition{
		Name:    "multiple_of",
		NumArgs: 1,
		Arg:     divisorArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			divisor, _ := args[0].NumberValue()
			if vi, ok := val.IntegralValue(); ok {
				if di, ok := args[0].IntegralValue(); ok {
					// Both integral (an exact int or a float with no
					// fraction): no epsilon, and no rounding above 2^53.
					if vi%di != 0 {
						return fmt.Errorf("value %v is not a multiple of %v", val, args[0])
					}
					return nil
				}
			}
			// Use epsilon for floating point comparison
			epsilon := 1e-10
			remainder := math.Mod(valNum, divisor)
			if remainder > epsilon && remainder < divisor-epsilon {
				return fmt.Errorf("value %v is not a multiple of %v", val, divisor)
			}
			return nil
		},
	}),
	"positive": define(ConstraintDefinition{
		Name:    "positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum <= 0 {
				return fmt.Errorf("value %v is not positive", val)
			}
			return nil
		},
	}),
	"negative": define(ConstraintDefinition{
		Name:    "negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum >= 0 {
				return fmt.Errorf("value %v is not negative", val)
			}
			return nil
		},
	}),
	"non_negative": define(ConstraintDefinition{
		Name:    "non_negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum < 0 {
				return fmt.Errorf("value %v is negative", val)
			}
			return nil
		},
	}),
	"non_positive": define(ConstraintDefinition{
		Name:    "non_positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum > 0 {
				return fmt.Errorf("value %v is positive", val)
			}
			return nil
		},
	}),
	"finite": define(ConstraintDefinition{
		Name:    "finite",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if math.IsInf(valNum, 0) || math.IsNaN(valNum) {
				return fmt.Errorf("value %v is not finite", val)
			}
			return nil
		},
	}),
	"infinite": define(ConstraintDefinition{
		Name:    "infinite",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !math.IsInf(valNum, 0) {
				return fmt.Errorf("value %v is not infinite", val)
			}
			return nil
		},
	}),
	"nan": define(ConstraintDefinition{
		Name:    "nan",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !math.IsNaN(valNum) {
				return fmt.Errorf("value %v is not NaN", val)
			}
			return nil
		},
	}),
}

func numberConstraintSet(arg box.Value) ([]box.Value, error) {
	if argList, ok := arg.ListValue(); ok {
		for _, item := range argList {
			if item.Kind() != box.ValueNumber {
				return nil, invalidArgf("expected number, got %s", item.Kind())
			}
		}
		return argList, nil
	}

	if arg.Kind() != box.ValueNumber {
		return nil, invalidArgf("expected number, got %s", arg.Kind())
	}
	return []box.Value{arg}, nil
}
