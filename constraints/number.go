// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

var NumberContraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"min": {
		Name:    "min",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("min constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c < 0 {
				return fmt.Errorf("value %v is not >= %v", val, args[0])
			}
			return nil
		},
	},
	"max": {
		Name:    "max",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("max constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c > 0 {
				return fmt.Errorf("value %v is not <= %v", val, args[0])
			}
			return nil
		},
	},
	"eq": {
		Name:    "eq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("eq constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !ordered || c != 0 {
				return fmt.Errorf("value %v is not equal to %v", val, args[0])
			}
			return nil
		},
	},
	"neq": {
		Name:    "neq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("neq constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c == 0 {
				return fmt.Errorf("value %v is equal to %v", val, args[0])
			}
			return nil
		},
	},
	"gt": {
		Name:    "gt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("gt constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c <= 0 {
				return fmt.Errorf("value %v is not > %v", val, args[0])
			}
			return nil
		},
	},
	"lt": {
		Name:    "lt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("lt constraint requires 1 argument")
			}
			if args[0].Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			c, ordered, err := box.CompareNumbers(val, args[0])
			if err != nil {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if ordered && c >= 0 {
				return fmt.Errorf("value %v is not < %v", val, args[0])
			}
			return nil
		},
	},
	"in": {
		Name:    "in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("in constraint requires 1 argument")
			}
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
	},
	"not_in": {
		Name:    "not_in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("not_in constraint requires 1 argument")
			}
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
	},
	"range": {
		Name:    "range",
		NumArgs: 2,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 2 {
				return invalidArgf("range constraint requires 2 arguments")
			}
			if val.Kind() != box.ValueNumber {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			min, max := args[0], args[1]
			if min.Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", min.Kind())
			}
			if max.Kind() != box.ValueNumber {
				return invalidArgf("expected number, got %s", max.Kind())
			}
			if c, ordered, _ := box.CompareNumbers(min, max); ordered && c > 0 {
				return invalidArgf("range minimum %v is greater than maximum %v", min, max)
			}
			cMin, okMin, _ := box.CompareNumbers(val, min)
			cMax, okMax, _ := box.CompareNumbers(val, max)
			if (okMin && cMin < 0) || (okMax && cMax > 0) {
				return fmt.Errorf("value %v is not in range [%v, %v]", val, min, max)
			}
			return nil
		},
	},
	"even": {
		Name:    "even",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"odd": {
		Name:    "odd",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"multiple_of": {
		Name:    "multiple_of",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			if len(args) != 1 {
				return invalidArgf("multiple_of constraint requires 1 argument")
			}
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			divisor, ok := args[0].NumberValue()
			if !ok {
				return invalidArgf("expected number, got %s", args[0].Kind())
			}
			if divisor == 0 {
				return invalidArgf("divisor cannot be zero")
			}
			// Use epsilon for floating point comparison
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
			epsilon := 1e-10
			remainder := math.Mod(valNum, divisor)
			if remainder > epsilon && remainder < divisor-epsilon {
				return fmt.Errorf("value %v is not a multiple of %v", val, divisor)
			}
			return nil
		},
	},
	"positive": {
		Name:    "positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum <= 0 {
				return fmt.Errorf("value %v is not positive", val)
			}
			return nil
		},
	},
	"negative": {
		Name:    "negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum >= 0 {
				return fmt.Errorf("value %v is not negative", val)
			}
			return nil
		},
	},
	"non_negative": {
		Name:    "non_negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum < 0 {
				return fmt.Errorf("value %v is negative", val)
			}
			return nil
		},
	},
	"non_positive": {
		Name:    "non_positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if valNum > 0 {
				return fmt.Errorf("value %v is positive", val)
			}
			return nil
		},
	},
	"finite": {
		Name:    "finite",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if math.IsInf(valNum, 0) || math.IsNaN(valNum) {
				return fmt.Errorf("value %v is not finite", val)
			}
			return nil
		},
	},
	"infinite": {
		Name:    "infinite",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !math.IsInf(valNum, 0) {
				return fmt.Errorf("value %v is not infinite", val)
			}
			return nil
		},
	},
	"nan": {
		Name:    "nan",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			valNum, ok := val.NumberValue()
			if !ok {
				return fmt.Errorf("expected number, got %s", val.Kind())
			}
			if !math.IsNaN(valNum) {
				return fmt.Errorf("value %v is not NaN", val)
			}
			return nil
		},
	},
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
