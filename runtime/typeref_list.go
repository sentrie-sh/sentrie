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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/pkg/errors"
)

func validateAgainstListTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.ListTypeRef) error {
	if _, ok := v.([]any); !ok {
		return errors.Errorf("value %v is not an array", v)
	}

	for _, item := range v.([]any) {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, item, typeRef.ElemType); err != nil {
			return errors.Wrapf(err, "item is not valid")
		}
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
		if _, ok := listContraintCheckers[constraint.Name]; !ok {
			return errors.Errorf("unknown constraint: %s applied to int64 at %s", constraint.Name, typeRef.Position())
		}

		if err := listContraintCheckers[constraint.Name](ctx, p, v.([]any), args); err != nil {
			return errors.Wrapf(err, "constraint is not valid")
		}
	}

	return nil
}

var listContraintCheckers map[string]constraintChecker[[]any] = map[string]constraintChecker[[]any]{
	"not_empty": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(val) == 0 {
			return fmt.Errorf("list is empty")
		}
		return nil
	},
}
