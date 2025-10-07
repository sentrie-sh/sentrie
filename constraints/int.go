package constraints

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/index"
)

var IntContraintCheckers map[string]ConstraintDefinition[int64] = map[string]ConstraintDefinition[int64]{
	"gte": {
		Name:    "gte",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("gte constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val < arg {
				return fmt.Errorf("value %v is not greater than or equal to %v", val, arg)
			}
			return nil
		},
	},
	"lte": {
		Name:    "lte",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("lte constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val > arg {
				return fmt.Errorf("value %v is not less than or equal to %v", val, arg)
			}
			return nil
		},
	},
	"eq": {
		Name:    "eq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("eq constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val != arg {
				return fmt.Errorf("value %v is not equal to %v", val, arg)
			}
			return nil
		},
	},
	"neq": {
		Name:    "neq",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("neq constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val == arg {
				return fmt.Errorf("value %v is not not equal to %v", val, arg)
			}
			return nil
		},
	},
	"gt": {
		Name:    "gt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("gt constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val <= arg {
				return fmt.Errorf("value %v is not greater than %v", val, arg)
			}
			return nil
		},
	},
	"lt": {
		Name:    "lt",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("lt constraint requires 1 argument")
			}
			arg := args[0].(int64)
			if val >= arg {
				return fmt.Errorf("value %v is not less than %v", val, arg)
			}
			return nil
		},
	},
	"in": {
		Name:    "in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("in constraint requires 1 argument")
			}
			var set []int64
			if _, ok := args[0].([]int64); ok {
				// if the first argument is a list of int64, use it
				set = args[0].([]int64)
			} else {
				// if the first argument is not a list of int64, use it as a single int64
				set = []int64{args[0].(int64)}
			}
			if !slices.Contains(set, val) {
				return fmt.Errorf("value %v is not in the set - expected in the set", val)
			}
			return nil
		},
	},
	"not_in": {
		Name:    "not_in",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("not_in constraint requires 1 argument")
			}
			var set []int64
			if _, ok := args[0].([]int64); ok {
				// if the first argument is a list of int64, use it
				set = args[0].([]int64)
			} else {
				// if the first argument is not a list of int64, use it as a single int64
				set = []int64{args[0].(int64)}
			}

			if slices.Contains(set, val) {
				return fmt.Errorf("value %v is in the set - expected not in the set", val)
			}
			return nil
		},
	},
	"range": {
		Name:    "range",
		NumArgs: 2,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 2 {
				return fmt.Errorf("range constraint requires 2 arguments")
			}
			min := args[0].(int64)
			max := args[1].(int64)
			if val < min || val > max {
				return fmt.Errorf("value %v is not in range [%v, %v]", val, min, max)
			}
			return nil
		},
	},
	"multiple_of": {
		Name:    "multiple_of",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("multiple_of constraint requires 1 argument")
			}
			divisor := args[0].(int64)
			if divisor == 0 {
				return fmt.Errorf("divisor cannot be zero")
			}
			if val%divisor != 0 {
				return fmt.Errorf("value %v is not a multiple of %v", val, divisor)
			}
			return nil
		},
	},
	"even": {
		Name:    "even",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val%2 != 0 {
				return fmt.Errorf("value %v is not even - expected even", val)
			}
			return nil
		},
	},
	"odd": {
		Name:    "odd",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val%2 == 0 {
				return fmt.Errorf("value %v is not odd - expected odd", val)
			}
			return nil
		},
	},
	"positive": {
		Name:    "positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val <= 0 {
				return fmt.Errorf("value %v is not positive - expected positive", val)
			}
			return nil
		},
	},
	"negative": {
		Name:    "negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val >= 0 {
				return fmt.Errorf("value %v is not negative - expected negative", val)
			}
			return nil
		},
	},
	"non_negative": {
		Name:    "non_negative",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val < 0 {
				return fmt.Errorf("value %v is negative - expected non-negative", val)
			}
			return nil
		},
	},
	"non_positive": {
		Name:    "non_positive",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val int64, args []any) error {
			if val > 0 {
				return fmt.Errorf("value %v is positive - expected non-positive", val)
			}
			return nil
		},
	},
}
