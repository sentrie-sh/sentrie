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
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/tokens"
	"github.com/pkg/errors"
)

func validateAgainstFloatTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.FloatTypeRef, pos tokens.Position) error {
	if _, ok := v.(float64); !ok {
		return errors.Errorf("value %v is not a float64", v)
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
		if _, ok := floatContraintCheckers[constraint.Name]; !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := floatContraintCheckers[constraint.Name](ctx, p, v.(float64), args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}
	return nil
}

var floatContraintCheckers map[string]constraintChecker[float64] = map[string]constraintChecker[float64]{
	"gte": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("gte constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val < arg {
			return fmt.Errorf("value %v is not >= %v", val, arg)
		}
		return nil
	},
	"lte": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lte constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val > arg {
			return fmt.Errorf("value %v is not <= %v", val, arg)
		}
		return nil
	},
	"eq": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("eq constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val != arg {
			return fmt.Errorf("value %v is not equal to %v", val, arg)
		}
		return nil
	},
	"neq": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("neq constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val == arg {
			return fmt.Errorf("value %v is equal to %v", val, arg)
		}
		return nil
	},
	"gt": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("gt constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val <= arg {
			return fmt.Errorf("value %v is not > %v", val, arg)
		}
		return nil
	},
	"lt": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("lt constraint requires 1 argument")
		}
		arg := args[0].(float64)
		if val >= arg {
			return fmt.Errorf("value %v is not < %v", val, arg)
		}
		return nil
	},
	"in": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
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
	"not_in": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
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
	"range": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
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
	"multiple_of": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
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
	"positive": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if val <= 0 {
			return fmt.Errorf("value %v is not positive", val)
		}
		return nil
	},
	"negative": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if val >= 0 {
			return fmt.Errorf("value %v is not negative", val)
		}
		return nil
	},
	"non_negative": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if val < 0 {
			return fmt.Errorf("value %v is negative", val)
		}
		return nil
	},
	"non_positive": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if val > 0 {
			return fmt.Errorf("value %v is positive", val)
		}
		return nil
	},
	"finite": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if math.IsInf(val, 0) || math.IsNaN(val) {
			return fmt.Errorf("value %v is not finite", val)
		}
		return nil
	},
	"infinite": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if !math.IsInf(val, 0) {
			return fmt.Errorf("value %v is not infinite", val)
		}
		return nil
	},
	"nan": func(ctx context.Context, p *index.Policy, val float64, args []any) error {
		if !math.IsNaN(val) {
			return fmt.Errorf("value %v is not NaN", val)
		}
		return nil
	},
}
