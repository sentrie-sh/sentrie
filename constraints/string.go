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
	"github.com/sentrie-sh/sentrie/index"
)

var StringContraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"length": {
		Name:    "length",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("length constraint requires 1 argument")
			}
			expectedLen, okn := args[0].NumberValue()
			if !okn {
				return fmt.Errorf("expected number, got %s", args[0].Kind())
			}
			if len(s) != int(expectedLen) {
				return fmt.Errorf("string length %d is not equal to %g", len(s), expectedLen)
			}
			return nil
		},
	},
	"minlength": {
		Name:    "minlength",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("minlength constraint requires 1 argument")
			}
			expectedLen, okn := args[0].NumberValue()
			if !okn {
				return fmt.Errorf("expected number, got %s", args[0].Kind())
			}
			if len(s) < int(expectedLen) {
				return fmt.Errorf("string length %d is not greater than or equal to %g", len(s), expectedLen)
			}
			return nil
		},
	},
	"maxlength": {
		Name:    "maxlength",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("maxlength constraint requires 1 argument")
			}
			expectedLen, okn := args[0].NumberValue()
			if !okn {
				return fmt.Errorf("expected number, got %s", args[0].Kind())
			}
			if len(s) > int(expectedLen) {
				return fmt.Errorf("string length %d is not less than or equal to %g", len(s), expectedLen)
			}
			return nil
		},
	},
	"regexp": {
		Name:    "regexp",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("regexp constraint requires 1 argument")
			}
			pattern, okp := args[0].StringValue()
			if !okp {
				return fmt.Errorf("expected string, got %s", args[0].Kind())
			}
			matched, err := regexp.MatchString(pattern, s)
			if err != nil {
				return fmt.Errorf("invalid regexp pattern: %v", err)
			}
			if !matched {
				return fmt.Errorf("string %q does not match pattern %q", s, pattern)
			}
			return nil
		},
	},
	"starts_with": {
		Name:    "starts_with",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("starts_with constraint requires 1 argument")
			}
			prefix, okp := args[0].StringValue()
			if !okp {
				return fmt.Errorf("expected string, got %s", args[0].Kind())
			}
			if !strings.HasPrefix(s, prefix) {
				return fmt.Errorf("string %q does not start with %q", s, prefix)
			}
			return nil
		},
	},
	"ends_with": {
		Name:    "ends_with",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("ends_with constraint requires 1 argument")
			}
			suffix, okp := args[0].StringValue()
			if !okp {
				return fmt.Errorf("expected string, got %s", args[0].Kind())
			}
			if !strings.HasSuffix(s, suffix) {
				return fmt.Errorf("string %q does not end with %q", s, suffix)
			}
			return nil
		},
	},
	"has_substring": {
		Name:    "has_substring",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("has_substring constraint requires 1 argument")
			}
			substring, okp := args[0].StringValue()
			if !okp {
				return fmt.Errorf("expected string, got %s", args[0].Kind())
			}
			if !strings.Contains(s, substring) {
				return fmt.Errorf("string %q does not contain %q", s, substring)
			}
			return nil
		},
	},
	"not_has_substring": {
		Name:    "not_has_substring",
		NumArgs: 1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) != 1 {
				return fmt.Errorf("not_has_substring constraint requires 1 argument")
			}
			substring, okp := args[0].StringValue()
			if !okp {
				return fmt.Errorf("expected string, got %s", args[0].Kind())
			}
			if strings.Contains(s, substring) {
				return fmt.Errorf("string %q contains %q", s, substring)
			}
			return nil
		},
	},
	"email": {
		Name:    "email",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"url": {
		Name:    "url",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"uuid": {
		Name:    "uuid",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"alphanumeric": {
		Name:    "alphanumeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"alpha": {
		Name:    "alpha",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"numeric": {
		Name:    "numeric",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
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
	},
	"lowercase": {
		Name:    "lowercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.ToLower(s) {
				return fmt.Errorf("string %q is not lowercase", s)
			}
			return nil
		},
	},
	"uppercase": {
		Name:    "uppercase",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.ToUpper(s) {
				return fmt.Errorf("string %q is not uppercase", s)
			}
			return nil
		},
	},
	"trimmed": {
		Name:    "trimmed",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s != strings.TrimSpace(s) {
				return fmt.Errorf("string %q has leading or trailing whitespace", s)
			}
			return nil
		},
	},
	"not_empty": {
		Name:    "not_empty",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if s == "" {
				return fmt.Errorf("string is empty")
			}
			return nil
		},
	},
	"one_of": {
		Name:    "one_of",
		NumArgs: -1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) < 1 {
				return fmt.Errorf("one_of constraint requires at least 1 argument")
			}
			for _, arg := range args {
				argString, oka := arg.StringValue()
				if !oka {
					return fmt.Errorf("expected string, got %s", arg.Kind())
				}
				if s == argString {
					return nil
				}
			}
			return fmt.Errorf("string %q is not one of the allowed values", s)
		},
	},
	"not_one_of": {
		Name:    "not_one_of",
		NumArgs: -1,
		Checker: func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error {
			s, ok := val.StringValue()
			if !ok {
				return fmt.Errorf("expected string, got %s", val.Kind())
			}
			if len(args) < 1 {
				return fmt.Errorf("not_one_of constraint requires at least 1 argument")
			}
			for _, arg := range args {
				argString, oka := arg.StringValue()
				if !oka {
					return fmt.Errorf("expected string, got %s", arg.Kind())
				}
				if s == argString {
					return fmt.Errorf("string %q is one of the allowed values", s)
				}
			}

			return nil
		},
	},
}
