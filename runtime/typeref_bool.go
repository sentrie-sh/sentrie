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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/pkg/errors"
)

func validateAgainstBoolTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.BoolTypeRef, expr ast.Expression) error {
	if _, ok := v.(bool); !ok {
		return errors.Errorf("value '%v' is not a bool at %s - expected bool", v, expr.Position())
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
		if _, ok := boolContraintCheckers[constraint.Name]; !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := boolContraintCheckers[constraint.Name](ctx, p, v.(bool), args); err != nil {
			return ErrConstraintFailed(expr, constraint, err)
		}
	}
	return nil
}

var boolContraintCheckers map[string]constraintChecker[bool] = map[string]constraintChecker[bool]{}
