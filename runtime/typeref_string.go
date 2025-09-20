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
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func validateAgainstStringTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef *ast.StringTypeRef, expr ast.Expression) error {
	if _, ok := v.(string); !ok {
		return errors.Errorf("value %v is not a string", v)
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
		if _, ok := stringContraintCheckers[constraint.Name]; !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := stringContraintCheckers[constraint.Name](ctx, p, v.(string), args); err != nil {
			return ErrConstraintFailed(expr, constraint, err)
		}
	}

	return nil
}

var stringContraintCheckers map[string]constraintChecker[string] = map[string]constraintChecker[string]{
	"length": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("length constraint requires 1 argument")
		}
		expectedLen := args[0].(int64)
		if len(val) != int(expectedLen) {
			return fmt.Errorf("string length %d is not equal to %d", len(val), expectedLen)
		}
		return nil
	},
	"minlength": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("minlength constraint requires 1 argument")
		}
		expectedLen := args[0].(int64)
		if len(val) < int(expectedLen) {
			return fmt.Errorf("string length %d is not greater than or equal to %d", len(val), expectedLen)
		}
		return nil
	},
	"maxlength": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("maxlength constraint requires 1 argument")
		}
		expectedLen := args[0].(int64)
		if len(val) > int(expectedLen) {
			return fmt.Errorf("string length %d is not less than or equal to %d", len(val), expectedLen)
		}
		return nil
	},
	"regexp": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("regexp constraint requires 1 argument")
		}
		pattern := args[0].(string)
		matched, err := regexp.MatchString(pattern, val)
		if err != nil {
			return fmt.Errorf("invalid regexp pattern: %v", err)
		}
		if !matched {
			return fmt.Errorf("string %q does not match pattern %q", val, pattern)
		}
		return nil
	},
	"starts_with": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("starts_with constraint requires 1 argument")
		}
		prefix := args[0].(string)
		if !strings.HasPrefix(val, prefix) {
			return fmt.Errorf("string %q does not start with %q", val, prefix)
		}
		return nil
	},
	"ends_with": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("ends_with constraint requires 1 argument")
		}
		suffix := args[0].(string)
		if !strings.HasSuffix(val, suffix) {
			return fmt.Errorf("string %q does not end with %q", val, suffix)
		}
		return nil
	},
	"has_substring": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("has_substring constraint requires 1 argument")
		}
		substring := args[0].(string)
		if !strings.Contains(val, substring) {
			return fmt.Errorf("string %q does not contain %q", val, substring)
		}
		return nil
	},
	"not_has_substring": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) != 1 {
			return fmt.Errorf("not_has_substring constraint requires 1 argument")
		}
		substring := args[0].(string)
		if strings.Contains(val, substring) {
			return fmt.Errorf("string %q contains %q", val, substring)
		}
		return nil
	},
	"email": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(val) {
			return fmt.Errorf("string %q is not a valid email", val)
		}
		return nil
	},
	"url": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`)
		if !urlRegex.MatchString(val) {
			return fmt.Errorf("string %q is not a valid URL", val)
		}
		return nil
	},
	"uuid": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		err := uuid.Validate(val)
		if err != nil {
			return fmt.Errorf("string %q is not a valid UUID: %v", val, err)
		}
		return nil
	},
	"alphanumeric": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		for _, r := range val {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return fmt.Errorf("string %q contains non-alphanumeric characters", val)
			}
		}
		return nil
	},
	"alpha": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		for _, r := range val {
			if !unicode.IsLetter(r) {
				return fmt.Errorf("string %q contains non-letter characters", val)
			}
		}
		return nil
	},
	"numeric": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		// Check if it's a valid integer or float
		// First try to parse as float64 to support both integers and floats
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return fmt.Errorf("string %q is not a valid numeric value", val)
		}
		return nil
	},
	"lowercase": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if val != strings.ToLower(val) {
			return fmt.Errorf("string %q is not lowercase", val)
		}
		return nil
	},
	"uppercase": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if val != strings.ToUpper(val) {
			return fmt.Errorf("string %q is not uppercase", val)
		}
		return nil
	},
	"trimmed": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if val != strings.TrimSpace(val) {
			return fmt.Errorf("string %q has leading or trailing whitespace", val)
		}
		return nil
	},
	"not_empty": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if val == "" {
			return fmt.Errorf("string is empty")
		}
		return nil
	},
	"one_of": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) < 1 {
			return fmt.Errorf("one_of constraint requires at least 1 argument")
		}
		for _, arg := range args {
			if val == arg.(string) {
				return nil
			}
		}
		return fmt.Errorf("string %q is not one of the allowed values", val)
	},
	"not_one_of": func(ctx context.Context, p *index.Policy, val string, args []any) error {
		if len(args) < 1 {
			return fmt.Errorf("not_one_of constraint requires at least 1 argument")
		}
		for _, arg := range args {
			if val == arg.(string) {
				return fmt.Errorf("string %q is one of the allowed values", val)
			}
		}

		return nil
	},
}
