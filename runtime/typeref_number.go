// SPDX-License-Identifier: Apache-2.0
//
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

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstNumberTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.NumberTypeRef, pos tokens.Range) error {
	if _, ok := v.(float64); !ok {
		return errors.Errorf("value %v is not a number", v)
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]any, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return err
			}
			if i64, ok := csArg.(int64); ok {
				csArg = float64(i64)
			}
			args[i] = csArg
		}
		checker, ok := constraints.NumberContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, v.(float64), args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}
	return nil
}
