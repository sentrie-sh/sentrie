// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/sentrie-sh/sentrie/box"
)

var StringContraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"length": define(ConstraintDefinition{
		Name:    "length",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			expectedLen, _ := args[0].NumberValue()
			if len(s) != int(expectedLen) {
				return fmt.Errorf("string length %d is not equal to %g", len(s), expectedLen)
			}
			return nil
		},
	}),
	"minlength": define(ConstraintDefinition{
		Name:    "minlength",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			expectedLen, _ := args[0].NumberValue()
			if len(s) < int(expectedLen) {
				return fmt.Errorf("string length %d is not greater than or equal to %g", len(s), expectedLen)
			}
			return nil
		},
	}),
	"maxlength": define(ConstraintDefinition{
		Name:    "maxlength",
		NumArgs: 1,
		Arg:     numberArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			expectedLen, _ := args[0].NumberValue()
			if len(s) > int(expectedLen) {
				return fmt.Errorf("string length %d is not less than or equal to %g", len(s), expectedLen)
			}
			return nil
		},
	}),
	"regexp": define(ConstraintDefinition{
		Name:    "regexp",
		NumArgs: 1,
		Arg:     regexpArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			pattern, _ := args[0].StringValue()
			// CheckArgs has compiled the pattern, so MustCompile cannot panic.
			if !regexp.MustCompile(pattern).MatchString(s) {
				return fmt.Errorf("string %q does not match pattern %q", s, pattern)
			}
			return nil
		},
	}),
	"starts_with": define(ConstraintDefinition{
		Name:    "starts_with",
		NumArgs: 1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			prefix, _ := args[0].StringValue()
			if !strings.HasPrefix(s, prefix) {
				return fmt.Errorf("string %q does not start with %q", s, prefix)
			}
			return nil
		},
	}),
	"ends_with": define(ConstraintDefinition{
		Name:    "ends_with",
		NumArgs: 1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			suffix, _ := args[0].StringValue()
			if !strings.HasSuffix(s, suffix) {
				return fmt.Errorf("string %q does not end with %q", s, suffix)
			}
			return nil
		},
	}),
	"has_substring": define(ConstraintDefinition{
		Name:    "has_substring",
		NumArgs: 1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			substring, _ := args[0].StringValue()
			if !strings.Contains(s, substring) {
				return fmt.Errorf("string %q does not contain %q", s, substring)
			}
			return nil
		},
	}),
	"not_has_substring": define(ConstraintDefinition{
		Name:    "not_has_substring",
		NumArgs: 1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			substring, _ := args[0].StringValue()
			if strings.Contains(s, substring) {
				return fmt.Errorf("string %q contains %q", s, substring)
			}
			return nil
		},
	}),
	"email": define(ConstraintDefinition{
		Name:    "email",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(s) {
				return fmt.Errorf("string %q is not a valid email", s)
			}
			return nil
		},
	}),
	"url": define(ConstraintDefinition{
		Name:    "url",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`)
			if !urlRegex.MatchString(s) {
				return fmt.Errorf("string %q is not a valid URL", s)
			}
			return nil
		},
	}),
	"uuid": define(ConstraintDefinition{
		Name:    "uuid",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			err := uuid.Validate(s)
			if err != nil {
				return fmt.Errorf("string %q is not a valid UUID: %v", s, err)
			}
			return nil
		},
	}),
	"alphanumeric": define(ConstraintDefinition{
		Name:    "alphanumeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			for _, r := range s {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return fmt.Errorf("string %q contains non-alphanumeric characters", s)
				}
			}
			return nil
		},
	}),
	"alpha": define(ConstraintDefinition{
		Name:    "alpha",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return fmt.Errorf("string %q contains non-letter characters", s)
				}
			}
			return nil
		},
	}),
	"numeric": define(ConstraintDefinition{
		Name:    "numeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			// Check if it's a valid integer or float
			// First try to parse as float64 to support both integers and floats
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return fmt.Errorf("string %q is not a valid numeric value", s)
			}
			return nil
		},
	}),
	"lowercase": define(ConstraintDefinition{
		Name:    "lowercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.ToLower(s) {
				return fmt.Errorf("string %q is not lowercase", s)
			}
			return nil
		},
	}),
	"uppercase": define(ConstraintDefinition{
		Name:    "uppercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.ToUpper(s) {
				return fmt.Errorf("string %q is not uppercase", s)
			}
			return nil
		},
	}),
	"trimmed": define(ConstraintDefinition{
		Name:    "trimmed",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.TrimSpace(s) {
				return fmt.Errorf("string %q has leading or trailing whitespace", s)
			}
			return nil
		},
	}),
	"not_empty": define(ConstraintDefinition{
		Name:    "not_empty",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s == "" {
				return fmt.Errorf("string is empty")
			}
			return nil
		},
	}),
	"one_of": define(ConstraintDefinition{
		Name:    "one_of",
		NumArgs: -1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			for _, arg := range args {
				argString, _ := arg.StringValue()
				if s == argString {
					return nil
				}
			}
			return fmt.Errorf("string %q is not one of the allowed values", s)
		},
	}),
	"not_one_of": define(ConstraintDefinition{
		Name:    "not_one_of",
		NumArgs: -1,
		Arg:     stringArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			for _, arg := range args {
				argString, _ := arg.StringValue()
				if s == argString {
					return fmt.Errorf("string %q is one of the allowed values", s)
				}
			}

			return nil
		},
	}),
}
