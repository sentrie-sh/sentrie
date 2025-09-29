package runtime

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstIntTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, val any, typeRef *ast.IntTypeRef, pos tokens.Position) error {
	if _, ok := val.(int64); !ok {
		return fmt.Errorf("value %v is not an int at %s - expected int", val, pos)
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
			return ErrUnknownConstraint(constraint)
		}

		if err := intContraintCheckers[constraint.Name](ctx, p, val.(int64), args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
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
			return fmt.Errorf("value %v is not greater than or equal to %v", val, arg)
		}
		return nil
	},
	"lte": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lte constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val > arg {
			return fmt.Errorf("value %v is not less than or equal to %v", val, arg)
		}
		return nil
	},
	"eq": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("eq constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val != arg {
			return fmt.Errorf("value %v is not equal to %v", val, arg)
		}
		return nil
	},
	"neq": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("neq constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val == arg {
			return fmt.Errorf("value %v is not not equal to %v", val, arg)
		}
		return nil
	},
	"gt": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("gt constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val <= arg {
			return fmt.Errorf("value %v is not greater than %v", val, arg)
		}
		return nil
	},
	"lt": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lt constraint requires 1 argument")
		}
		arg := args[0].(int64)
		if val >= arg {
			return fmt.Errorf("value %v is not less than %v", val, arg)
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
			return fmt.Errorf("value %v is not in the set - expected in the set", val)
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
			return fmt.Errorf("value %v is in the set - expected not in the set", val)
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
			return fmt.Errorf("value %v is not even - expected even", val)
		}
		return nil
	},
	"odd": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val%2 == 0 {
			return fmt.Errorf("value %v is not odd - expected odd", val)
		}
		return nil
	},
	"positive": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val <= 0 {
			return fmt.Errorf("value %v is not positive - expected positive", val)
		}
		return nil
	},
	"negative": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val >= 0 {
			return fmt.Errorf("value %v is not negative - expected negative", val)
		}
		return nil
	},
	"non_negative": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val < 0 {
			return fmt.Errorf("value %v is negative - expected non-negative", val)
		}
		return nil
	},
	"non_positive": func(ctx context.Context, p *index.Policy, val int64, args []any) error {
		if val > 0 {
			return fmt.Errorf("value %v is positive - expected non-positive", val)
		}
		return nil
	},
}
