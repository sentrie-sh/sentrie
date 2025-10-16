package constraints

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/sentrie-sh/sentrie/index"
)

var NumberContraintCheckers map[string]ConstraintDefinition[float64] = map[string]ConstraintDefinition[float64]{
	"min": {
		Name:    "min",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("min constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val < arg {
				return fmt.Errorf("value %v is not >= %v", val, arg)
			}
			return nil
		},
	},
	"max": {
		Name:    "max",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("max constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val > arg {
				return fmt.Errorf("value %v is not <= %v", val, arg)
			}
			return nil
		},
	},
	"eq": {
		Name:    "eq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("eq constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val != arg {
				return fmt.Errorf("value %v is not equal to %v", val, arg)
			}
			return nil
		},
	},
	"neq": {
		Name:    "neq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("neq constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val == arg {
				return fmt.Errorf("value %v is equal to %v", val, arg)
			}
			return nil
		},
	},
	"gt": {
		Name:    "gt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("gt constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val <= arg {
				return fmt.Errorf("value %v is not > %v", val, arg)
			}
			return nil
		},
	},
	"lt": {
		Name:    "lt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("lt constraint requires 1 argument")
			}
			arg := args[0].(float64)
			if val >= arg {
				return fmt.Errorf("value %v is not < %v", val, arg)
			}
			return nil
		},
	},
	"in": {
		Name:    "in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("in constraint requires 1 argument")
			}
			var set []float64
			if _, ok := args[0].([]float64); ok {
				// if the first argument is a list of float64, use it
				set = args[0].([]float64)
			} else {
				// if the first argument is not a list of float64, use it as a single float64
				set = []float64{args[0].(float64)}
			}
			if !slices.Contains(set, val) {
				return fmt.Errorf("value %v is not in the set", val)
			}
			return nil
		},
	},
	"not_in": {
		Name:    "not_in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("not_in constraint requires 1 argument")
			}
			var set []float64
			if _, ok := args[0].([]float64); ok {
				// if the first argument is a list of float64, use it
				set = args[0].([]float64)
			} else {
				// if the first argument is not a list of float64, use it as a single float64
				set = []float64{args[0].(float64)}
			}

			if slices.Contains(set, val) {
				return fmt.Errorf("value %v is in the set", val)
			}
			return nil
		},
	},
	"range": {
		Name:    "range",
		NumArgs: 2,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 2 {
				return fmt.Errorf("range constraint requires 2 arguments")
			}
			min := args[0].(float64)
			max := args[1].(float64)
			if val < min || val > max {
				return fmt.Errorf("value %v is not in range [%v, %v]", val, min, max)
			}
			return nil
		},
	},
	"even": {
		Name:    "even",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if math.Mod(val, 2) != 0 {
				return fmt.Errorf("value %v is not even", val)
			}
			return nil
		},
	},
	"odd": {
		Name:    "odd",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if math.Mod(val, 2) == 0 {
				return fmt.Errorf("value %v is not odd", val)
			}
			return nil
		},
	},
	"multiple_of": {
		Name:    "multiple_of",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("multiple_of constraint requires 1 argument")
			}
			divisor := args[0].(float64)
			if divisor == 0 {
				return fmt.Errorf("divisor cannot be zero")
			}
			// Use epsilon for floating point comparison
			epsilon := 1e-10
			remainder := math.Mod(val, divisor)
			if remainder > epsilon && remainder < divisor-epsilon {
				return fmt.Errorf("value %v is not a multiple of %v", val, divisor)
			}
			return nil
		},
	},
	"positive": {
		Name:    "positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if val <= 0 {
				return fmt.Errorf("value %v is not positive", val)
			}
			return nil
		},
	},
	"negative": {
		Name:    "negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if val >= 0 {
				return fmt.Errorf("value %v is not negative", val)
			}
			return nil
		},
	},
	"non_negative": {
		Name:    "non_negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if val < 0 {
				return fmt.Errorf("value %v is negative", val)
			}
			return nil
		},
	},
	"non_positive": {
		Name:    "non_positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if val > 0 {
				return fmt.Errorf("value %v is positive", val)
			}
			return nil
		},
	},
	"finite": {
		Name:    "finite",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if math.IsInf(val, 0) || math.IsNaN(val) {
				return fmt.Errorf("value %v is not finite", val)
			}
			return nil
		},
	},
	"infinite": {
		Name:    "infinite",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if !math.IsInf(val, 0) {
				return fmt.Errorf("value %v is not infinite", val)
			}
			return nil
		},
	},
	"nan": {
		Name:    "nan",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val float64, args []any) error {
			if !math.IsNaN(val) {
				return fmt.Errorf("value %v is not NaN", val)
			}
			return nil
		},
	},
}
