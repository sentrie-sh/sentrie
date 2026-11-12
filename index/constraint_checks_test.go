// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"

	"github.com/sentrie-sh/sentrie/constraints"
)

// TestValidateRejectsListFormOfIn checks the old `@in([1, 2])` list form is an
// invalid argument at validate time, for numbers and strings (#146).
func (s *IndexTestSuite) TestValidateRejectsListFormOfIn() {
	cases := []struct{ name, typ, value, msg string }{
		{"number in", `number@in([1, 2])`, `1`, "invalid argument to constraint 'in': in constraint takes each value as its own argument, as in @in(1, 2), not a list"},
		{"number not_in", `number@not_in([1, 2])`, `1`, "invalid argument to constraint 'not_in': not_in constraint takes each value as its own argument, as in @not_in(1, 2), not a list"},
		{"string in", `string@in(["a", "b"])`, `"a"`, "takes each value as its own argument"},
		{"string not_in", `string@not_in(["a"])`, `"a"`, "takes each value as its own argument"},
	}
	for position := range constraintArgPositions {
		for _, tc := range cases {
			s.Run(position+"/"+tc.name, func() {
				err := s.parseAndValidate(s.constraintArgSource(position, tc.typ, tc.value), "c.sentrie")
				s.Require().Error(err)
				s.Contains(err.Error(), tc.msg)
				s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
			})
		}
	}
}

// TestValidateAcceptsVariadicInAndNegativeArguments checks the new valid forms
// validate in every type position (#146).
func (s *IndexTestSuite) TestValidateAcceptsVariadicInAndNegativeArguments() {
	cases := []struct{ name, typ, value string }{
		{"number in", `number@in(1, 2)`, `1`},
		{"number not_in", `number@not_in(1, -2)`, `1`},
		{"string in", `string@in("a", "b")`, `"a"`},
		{"string not_in", `string@not_in("a")`, `"b"`},
		{"min negative", `number@min(-5)`, `-3`},
		{"range negative", `number@range(-10, -1)`, `-3`},
		{"gt negative float", `number@gt(-0.5)`, `0`},
		{"multiple_of negative", `number@multiple_of(-3)`, `6`},
		{"url default", `string@url()`, `"http://localhost:8080"`},
		{"url schemes", `string@url("http", "https", "ftp")`, `"ftp://a.co"`},
	}
	for position := range constraintArgPositions {
		for _, tc := range cases {
			s.Run(position+"/"+tc.name, func() {
				s.NoError(s.parseAndValidate(s.constraintArgSource(position, tc.typ, tc.value), "c.sentrie"))
			})
		}
	}
}

// TestValidateRejectsBadURLSchemeArgument checks `@url("ftp://")` is an
// invalid argument (#146).
func (s *IndexTestSuite) TestValidateRejectsBadURLSchemeArgument() {
	for _, typ := range []string{`string@url("ftp://")`, `string@url("")`, `string@url(1)`} {
		err := s.parseAndValidate(s.constraintArgSource("fact", typ, `"a"`), "c.sentrie")
		s.Require().Error(err, typ)
		s.Contains(err.Error(), "invalid argument to constraint 'url'", typ)
		s.True(errors.Is(err, constraints.ErrInvalidArgument), typ)
	}
}
