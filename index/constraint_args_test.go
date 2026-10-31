// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/xerr"
)

// constraintArgPositions wraps a type with a broken or valid constraint in
// every position a type ref can appear. %[1]s is the constrained type, such
// as `string@regexp("[")`; %[2]s is a value that conforms to its kind.
var constraintArgPositions = map[string]string{
	"fact":              "policy p {\n  fact x: %[1]s as x\n  rule r = { yield true }\n  export decision of r\n}",
	"policy let":        "policy p {\n  let v: %[1]s = %[2]s\n  rule r = { yield true }\n  export decision of r\n}",
	"block let":         "policy p {\n  rule r = {\n    let v: %[1]s = %[2]s\n    yield true\n  }\n  export decision of r\n}",
	"shape alias":       "shape S %[1]s\npolicy p {\n  rule r = { yield true }\n  export decision of r\n}",
	"shape field":       "shape S { f: %[1]s }\npolicy p {\n  rule r = { yield true }\n  export decision of r\n}",
	"policy shape":      "policy p {\n  shape S { f: %[1]s }\n  rule r = { yield true }\n  export decision of r\n}",
	"derive param":      "derive d = (a: %[1]s): trinary => { yield true }\npolicy p {\n  rule r = { yield true }\n  export decision of r\n}",
	"derive return":     "derive d = (): %[1]s => { yield %[2]s }\npolicy p {\n  rule r = { yield true }\n  export decision of r\n}",
	"lambda param":      "policy p {\n  rule r = { yield any([1], (a: %[1]s): trinary => { yield true }) }\n  export decision of r\n}",
	"lambda return":     "policy p {\n  rule r = { yield any([1], (a: number): %[1]s => { yield %[2]s }) }\n  export decision of r\n}",
	"cast":              "policy p {\n  rule r = { yield (%[2]s as %[1]s) is defined }\n  export decision of r\n}",
	"is":                "policy p {\n  rule r = { yield 1 is not %[1]s }\n  export decision of r\n}",
	"list element":      "policy p {\n  rule r = { yield [] is list[%[1]s] }\n  export decision of r\n}",
	"dict value":        "policy p {\n  rule r = { yield {} is dict[%[1]s] }\n  export decision of r\n}",
	"record field":      "policy p {\n  rule r = { yield [1, %[2]s] is record[number, %[1]s] }\n  export decision of r\n}",
	"nested list":       "policy p {\n  fact x?: list[list[%[1]s]]? as x\n  rule r = { yield true }\n  export decision of r\n}",
	"rule default":      "policy p {\n  rule r = default (1 is %[1]s) { yield true }\n  export decision of r\n}",
	"derive body is":    "derive d = (v: document): trinary => { yield v.a is %[1]s }\npolicy p {\n  rule r = { yield true }\n  export decision of r\n}",
	"lambda in lambda":  "policy p {\n  rule r = { yield any([1], (a) => { yield any([2], (b: %[1]s) => { yield true }) }) }\n  export decision of r\n}",
	"is inside a block": "policy p {\n  rule r = {\n    let y = 1\n    yield y is %[1]s\n  }\n  export decision of r\n}",
}

func (s *IndexTestSuite) constraintArgSource(position, typ, value string) string {
	tmpl, ok := constraintArgPositions[position]
	s.Require().True(ok, position)
	return "namespace com/ex\n" + fmt.Sprintf(tmpl, typ, value) + "\n"
}

// TestValidateRejectsBrokenConstraintArguments checks that validation reports
// a broken constraint argument, with the constraint's span, in every position
// a type ref can appear (#273).
func (s *IndexTestSuite) TestValidateRejectsBrokenConstraintArguments() {
	broken := []struct {
		name, typ, value, msg string
	}{
		{"invalid regexp", `string@regexp("[")`, `"a"`, "invalid argument to constraint 'regexp': invalid regexp pattern: error parsing regexp: missing closing ]"},
		{"wrong kind", `string@minlength("q")`, `"a"`, "invalid argument to constraint 'minlength': expected number, got string"},
		{"zero divisor", `number@multiple_of(0)`, `1`, "invalid argument to constraint 'multiple_of': divisor cannot be zero"},
		{"range minimum above maximum", `number@range(5, 1)`, `1`, "invalid argument to constraint 'range': range minimum 5 is greater than maximum 1"},
		{"wrong set item kind", `number@not_in([1, "x"])`, `1`, "invalid argument to constraint 'not_in': expected number, got string"},
		{"wrong trinary kind", `trinary@eq(1)`, `true`, "invalid argument to constraint 'eq': eq constraint expects a boolean argument"},
		{"variadic item kind", `string@one_of("a", 1)`, `"a"`, "invalid argument to constraint 'one_of': expected string, got number"},
	}
	for position := range constraintArgPositions {
		for _, tc := range broken {
			s.Run(position+"/"+tc.name, func() {
				err := s.parseAndValidate(s.constraintArgSource(position, tc.typ, tc.value), "c.sentrie")
				s.Require().Error(err)
				s.Contains(err.Error(), tc.msg)
				s.True(errors.Is(err, xerr.ErrIndex), err.Error())
				s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())

				var argErr xerr.ConstraintArgError
				s.Require().True(errors.As(err, &argErr), err.Error())
				s.Equal("c.sentrie", argErr.At.File)
				s.Positive(argErr.At.From.Line)
			})
		}
	}
}

// TestValidateAcceptsValidConstraintArguments checks that well-formed
// constraint arguments still validate in every position.
func (s *IndexTestSuite) TestValidateAcceptsValidConstraintArguments() {
	valid := []struct {
		name, typ, value string
	}{
		{"regexp", `string@regexp("^[a-z]+$")`, `"a"`},
		{"minlength", `string@minlength(1)`, `"a"`},
		{"multiple_of", `number@multiple_of(0.5)`, `1`},
		{"range", `number@range(1, 1)`, `1`},
		{"not_in list", `number@not_in([3, 4])`, `1`},
		{"trinary eq", `trinary@eq(unknown)`, `unknown`},
		{"one_of", `string@one_of("a", "b")`, `"a"`},
		{"no arguments", `string@email()`, `"a@b.co"`},
	}
	for position := range constraintArgPositions {
		for _, tc := range valid {
			s.Run(position+"/"+tc.name, func() {
				s.NoError(s.parseAndValidate(s.constraintArgSource(position, tc.typ, tc.value), "c.sentrie"))
			})
		}
	}
}

// TestValidateReportsEveryBrokenConstraintArgument checks that every broken
// argument is reported, in a stable order.
func (s *IndexTestSuite) TestValidateReportsEveryBrokenConstraintArgument() {
	src := "namespace com/ex\npolicy p {\n  fact a: string@regexp(\"[\") as a\n  fact b: number@multiple_of(0) as b\n  rule r = { yield true }\n  export decision of r\n}\n"
	first := s.parseAndValidate(src, "c.sentrie")
	s.Require().Error(first)
	s.Contains(first.Error(), "invalid regexp pattern")
	s.Contains(first.Error(), "divisor cannot be zero")
	for range 5 {
		s.Equal(first.Error(), s.parseAndValidate(src, "c.sentrie").Error())
	}
}
