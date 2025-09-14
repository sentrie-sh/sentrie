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

func validateAgainstRecordTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.RecordTypeRef) error {
	var value []any
	if arr, ok := v.([]any); ok {
		value = arr
	} else {
		return errors.Errorf("value %v is not a record", v) // TODO: improve this error message
	}

	if len(value) != len(typeRef.Fields) {
		return errors.Errorf("fields length mismatch: %v", v) // TODO: improve this error message
	}

	for i, field := range typeRef.Fields {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, value[i], field); err != nil {
			return errors.Wrapf(err, "%v is not a valid record field", v)
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
		if _, ok := recordContraintCheckers[constraint.Name]; !ok {
			return errors.Errorf("unknown constraint: %s applied to record at %s", constraint.Name, typeRef.Position())
		}

		if err := recordContraintCheckers[constraint.Name](ctx, p, v.([]any), args); err != nil {
			return errors.Wrapf(err, "constraint is not valid")
		}
	}

	return nil
}

var recordContraintCheckers map[string]constraintChecker[[]any] = map[string]constraintChecker[[]any]{}
