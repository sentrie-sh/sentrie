package runtime

import (
	"context"
	"fmt"
	"slices"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
)

func validateAgainstIntTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, val any, typeRef *ast.IntTypeRef) error {
	if _, ok := val.(int64); !ok {
		return fmt.Errorf("value %v is not an int64", val)
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]any, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return err
			}
			args[i] = csArg
		}
		if _, ok := intContraintCheckers[constraint.Name]; !ok {
			return fmt.Errorf("unknown constraint: %s applied to int64 at %s", constraint.Name, typeRef.Position())
		}

		if err := intContraintCheckers[constraint.Name](ctx, p, val.(int64), args); err != nil {
			return err
		}
	}

	return nil
}

var intContraintCheckers map[string]constraintChecker[int64] = map[string]constraintChecker[int64]{
	"gte": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("gte constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val < arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"lte": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lte constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val > arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"eq": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("eq constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val != arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"neq": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("neq constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val == arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"gt": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("gt constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val <= arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"lt": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lt constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val >= arg {
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"in": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
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
			return fmt.Errorf("value %v is not an int64", val)
		}
		return nil
	},
	"not_in": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
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
			return fmt.Errorf("value %v is in the set", val)
		}
		return nil
	},
	"range": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
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
	"multiple_of": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
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
	"even": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val%2 != 0 {
			return fmt.Errorf("value %v is not even", val)
		}
		return nil
	},
	"odd": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val%2 == 0 {
			return fmt.Errorf("value %v is not odd", val)
		}
		return nil
	},
	"positive": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val <= 0 {
			return fmt.Errorf("value %v is not positive", val)
		}
		return nil
	},
	"negative": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val >= 0 {
			return fmt.Errorf("value %v is not negative", val)
		}
		return nil
	},
	"non_negative": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val < 0 {
			return fmt.Errorf("value %v is negative", val)
		}
		return nil
	},
	"non_positive": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val > 0 {
			return fmt.Errorf("value %v is positive", val)
		}
		return nil
	},
}
