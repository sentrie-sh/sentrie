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
	"reflect"
	"slices"

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
	"sorted": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if !isSorted(val, false) {
			return fmt.Errorf("list is not sorted in ascending order")
		}
		return nil
	},
	"sorted_desc": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if !isSorted(val, true) {
			return fmt.Errorf("list is not sorted in descending order")
		}
		return nil
	},
	"has_item": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("has_item constraint requires 1 argument")
		}
		item := args[0]
		if !slices.Contains(val, item) {
			return fmt.Errorf("list does not contain item %v", item)
		}
		return nil
	},
	"not_has_item": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("not_has_item constraint requires 1 argument")
		}
		item := args[0]
		if slices.Contains(val, item) {
			return fmt.Errorf("list contains item %v", item)
		}
		return nil
	},
	"subset_of": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("subset_of constraint requires 1 argument")
		}
		superset, ok := args[0].([]any)
		if !ok {
			return fmt.Errorf("subset_of argument must be a list")
		}
		for _, item := range val {
			if !slices.Contains(superset, item) {
				return fmt.Errorf("list item %v is not in superset", item)
			}
		}
		return nil
	},
	"superset_of": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("superset_of constraint requires 1 argument")
		}
		subset, ok := args[0].([]any)
		if !ok {
			return fmt.Errorf("superset_of argument must be a list")
		}
		for _, item := range subset {
			if !slices.Contains(val, item) {
				return fmt.Errorf("list does not contain subset item %v", item)
			}
		}
		return nil
	},
	"disjoint_from": func(ctx context.Context, p *index.Policy, val []any, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("disjoint_from constraint requires 1 argument")
		}
		other, ok := args[0].([]any)
		if !ok {
			return fmt.Errorf("disjoint_from argument must be a list")
		}
		for _, item := range val {
			if slices.Contains(other, item) {
				return fmt.Errorf("list item %v is also in the other list", item)
			}
		}
		return nil
	},
}

// isSorted checks if a slice is sorted in ascending (desc=false) or descending (desc=true) order
func isSorted(slice []any, desc bool) bool {
	if len(slice) <= 1 {
		return true
	}

	for i := 1; i < len(slice); i++ {
		prev := slice[i-1]
		curr := slice[i]

		// Compare based on type
		comparison := compareValues(prev, curr)
		if desc {
			if comparison < 0 {
				return false
			}
		} else {
			if comparison > 0 {
				return false
			}
		}
	}
	return true
}

// compareValues compares two values and returns -1, 0, or 1
func compareValues(a, b any) int {
	// Handle different types
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		// Convert to strings for comparison
		return compareStrings(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}

	switch va := a.(type) {
	case int64:
		vb := b.(int64)
		if va < vb {
			return -1
		} else if va > vb {
			return 1
		}
		return 0
	case float64:
		vb := b.(float64)
		if va < vb {
			return -1
		} else if va > vb {
			return 1
		}
		return 0
	case string:
		vb := b.(string)
		return compareStrings(va, vb)
	default:
		// Fallback to string comparison
		return compareStrings(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
	}
}

// compareStrings compares two strings lexicographically
func compareStrings(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}
