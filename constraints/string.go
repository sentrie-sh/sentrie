package constraints

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/sentrie-sh/sentrie/index"
)

var StringContraintCheckers map[string]ConstraintDefinition[string] = map[string]ConstraintDefinition[string]{
	"length": {
		Name:    "length",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("length constraint requires 1 argument")
			}
			expectedLen := args[0].(int64)
			if len(val) != int(expectedLen) {
				return fmt.Errorf("string length %d is not equal to %d", len(val), expectedLen)
			}
			return nil
		},
	},
	"minlength": {
		Name:    "minlength",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("minlength constraint requires 1 argument")
			}
			expectedLen := args[0].(int64)
			if len(val) < int(expectedLen) {
				return fmt.Errorf("string length %d is not greater than or equal to %d", len(val), expectedLen)
			}
			return nil
		},
	},
	"maxlength": {
		Name:    "maxlength",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("maxlength constraint requires 1 argument")
			}
			expectedLen := args[0].(int64)
			if len(val) > int(expectedLen) {
				return fmt.Errorf("string length %d is not less than or equal to %d", len(val), expectedLen)
			}
			return nil
		},
	},
	"regexp": {
		Name:    "regexp",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
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
	},
	"starts_with": {
		Name:    "starts_with",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("starts_with constraint requires 1 argument")
			}
			prefix := args[0].(string)
			if !strings.HasPrefix(val, prefix) {
				return fmt.Errorf("string %q does not start with %q", val, prefix)
			}
			return nil
		},
	},
	"ends_with": {
		Name:    "ends_with",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("ends_with constraint requires 1 argument")
			}
			suffix := args[0].(string)
			if !strings.HasSuffix(val, suffix) {
				return fmt.Errorf("string %q does not end with %q", val, suffix)
			}
			return nil
		},
	},
	"has_substring": {
		Name:    "has_substring",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("has_substring constraint requires 1 argument")
			}
			substring := args[0].(string)
			if !strings.Contains(val, substring) {
				return fmt.Errorf("string %q does not contain %q", val, substring)
			}
			return nil
		},
	},
	"not_has_substring": {
		Name:    "not_has_substring",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if len(args) != 1 {
				return fmt.Errorf("not_has_substring constraint requires 1 argument")
			}
			substring := args[0].(string)
			if strings.Contains(val, substring) {
				return fmt.Errorf("string %q contains %q", val, substring)
			}
			return nil
		},
	},
	"email": {
		Name:    "email",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(val) {
				return fmt.Errorf("string %q is not a valid email", val)
			}
			return nil
		},
	},
	"url": {
		Name:    "url",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`)
			if !urlRegex.MatchString(val) {
				return fmt.Errorf("string %q is not a valid URL", val)
			}
			return nil
		},
	},
	"uuid": {
		Name:    "uuid",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			err := uuid.Validate(val)
			if err != nil {
				return fmt.Errorf("string %q is not a valid UUID: %v", val, err)
			}
			return nil
		},
	},
	"alphanumeric": {
		Name:    "alphanumeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			for _, r := range val {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return fmt.Errorf("string %q contains non-alphanumeric characters", val)
				}
			}
			return nil
		},
	},
	"alpha": {
		Name:    "alpha",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			for _, r := range val {
				if !unicode.IsLetter(r) {
					return fmt.Errorf("string %q contains non-letter characters", val)
				}
			}
			return nil
		},
	},
	"numeric": {
		Name:    "numeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			// Check if it's a valid integer or float
			// First try to parse as float64 to support both integers and floats
			if _, err := strconv.ParseFloat(val, 64); err != nil {
				return fmt.Errorf("string %q is not a valid numeric value", val)
			}
			return nil
		},
	},
	"lowercase": {
		Name:    "lowercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if val != strings.ToLower(val) {
				return fmt.Errorf("string %q is not lowercase", val)
			}
			return nil
		},
	},
	"uppercase": {
		Name:    "uppercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if val != strings.ToUpper(val) {
				return fmt.Errorf("string %q is not uppercase", val)
			}
			return nil
		},
	},
	"trimmed": {
		Name:    "trimmed",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if val != strings.TrimSpace(val) {
				return fmt.Errorf("string %q has leading or trailing whitespace", val)
			}
			return nil
		},
	},
	"not_empty": {
		Name:    "not_empty",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
			if val == "" {
				return fmt.Errorf("string is empty")
			}
			return nil
		},
	},
	"one_of": {
		Name:    "one_of",
		NumArgs: -1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
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
	},
	"not_one_of": {
		Name:    "not_one_of",
		NumArgs: -1,
		Checker: func(ctx context.Context, p *index.Policy, val string, args []any) error {
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
	},
}
