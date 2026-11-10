// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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
			if n := utf8.RuneCountInString(s); n != int(expectedLen) {
				return fmt.Errorf("string length %d is not equal to %g", n, expectedLen)
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
			if n := utf8.RuneCountInString(s); n < int(expectedLen) {
				return fmt.Errorf("string length %d is not greater than or equal to %g", n, expectedLen)
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
			if n := utf8.RuneCountInString(s); n > int(expectedLen) {
				return fmt.Errorf("string length %d is not less than or equal to %g", n, expectedLen)
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
			// CheckArgs has compiled and cached the pattern.
			re, err := compilePattern(pattern)
			if err != nil {
				return invalidArgf("invalid regexp pattern: %v", err)
			}
			if !re.MatchString(s) {
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
			if !emailPattern.MatchString(s) {
				return fmt.Errorf("string %q is not a valid email", s)
			}
			return nil
		},
	}),
	"url": define(ConstraintDefinition{
		Name:    "url",
		NumArgs: -2,
		Arg:     schemeArg,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			u, err := url.Parse(s)
			if err != nil || u.Scheme == "" || u.Hostname() == "" {
				return fmt.Errorf("string %q is not a valid URL", s)
			}
			allowed := defaultURLSchemes
			if len(args) > 0 {
				// Scheme arguments replace the default set.
				allowed = make([]string, 0, len(args))
				for _, arg := range args {
					scheme, _ := arg.StringValue()
					allowed = append(allowed, strings.ToLower(scheme))
				}
			}
			// url.Parse lower-cases the scheme.
			if !slices.Contains(allowed, u.Scheme) {
				return fmt.Errorf("URL %q has scheme %q, expected one of %s", s, u.Scheme, strings.Join(allowed, ", "))
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
	"in": define(ConstraintDefinition{
		Name:    "in",
		NumArgs: -1,
		Arg:     memberArg(stringArg, `"a", "b"`),
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if !isMember(val, args) {
				return fmt.Errorf("string %q is not in the set", s)
			}
			return nil
		},
	}),
	"not_in": define(ConstraintDefinition{
		Name:    "not_in",
		NumArgs: -1,
		Arg:     memberArg(stringArg, `"a", "b"`),
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if isMember(val, args) {
				return fmt.Errorf("string %q is in the set", s)
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

// emailPattern is compiled once, not on every check.
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// defaultURLSchemes are the schemes `@url()` accepts with no arguments;
// any scheme arguments replace this set.
var defaultURLSchemes = []string{"http", "https"}

// urlSchemePattern is the RFC 3986 scheme syntax.
var urlSchemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*$`)

// schemeArg requires a string that is a URL scheme, such as "ftp". A value
// such as "ftp://" could never match, so it is rejected here.
func schemeArg(name string, arg box.Value) error {
	if err := stringArg(name, arg); err != nil {
		return err
	}
	scheme, _ := arg.StringValue()
	if !urlSchemePattern.MatchString(scheme) {
		return invalidArgf("invalid URL scheme %q", scheme)
	}
	return nil
}
