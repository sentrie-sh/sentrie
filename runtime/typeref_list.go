// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstListTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ListTypeRef, pos tokens.Range) error {
	items, ok := v.ListValue()
	if !ok {
		return fmt.Errorf("value %v is not an array at %s - expected array", v, pos)
	}

	for _, item := range items {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, item, typeRef.ElemType, pos); err != nil {
			return fmt.Errorf("item is not valid at %s: %w", pos, err) // TODO: improve this error message
		}
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]box.Value, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return err
			}
			args[i] = csArg
		}
		checker, ok := constraints.ListContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, v, args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}
