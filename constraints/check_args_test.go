// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"errors"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/trinary"
)

func allConstraintTables() map[string]map[string]constraints.ConstraintDefinition {
	return map[string]map[string]constraints.ConstraintDefinition{
		"string":   constraints.StringContraintCheckers,
		"number":   constraints.NumberContraintCheckers,
		"trinary":  constraints.TrinaryConstraintCheckers,
		"list":     constraints.ListContraintCheckers,
		"dict":     constraints.DictContraintCheckers,
		"document": constraints.DocumentContraintCheckers,
		"record":   constraints.RecordContraintCheckers,
		"shape":    constraints.ShapeContraintCheckers,
	}
}

// TestDefinitionsDescribeTheirArguments checks the argument metadata every
// definition carries: its key is its name, and a constraint that takes
// arguments says how each one is checked.
func (s *ConstraintsTestSuite) TestDefinitionsDescribeTheirArguments() {
	for kind, table := range allConstraintTables() {
		for name, def := range table {
			s.Equal(name, def.Name, "%s@%s", kind, name)
			s.GreaterOrEqual(def.NumArgs, -1, "%s@%s", kind, name)
			s.Equal(def.NumArgs != 0, def.Arg != nil, "%s@%s: Arg is set exactly when the constraint takes arguments", kind, name)
			s.NotNil(def.Checker, "%s@%s", kind, name)
		}
	}
}

// TestCheckArgs checks the count, kind and per-constraint rules CheckArgs
// enforces, all without a value, and that every failure is an argument
// failure with the message the checker has always reported.
func (s *ConstraintsTestSuite) TestCheckArgs() {
	str := constraints.StringContraintCheckers
	num := constraints.NumberContraintCheckers
	tri := constraints.TrinaryConstraintCheckers
	list := constraints.ListContraintCheckers
	n := func(f float64) box.Value { return box.Number(f) }
	sv := func(v string) box.Value { return box.String(v) }
	args := func(vs ...box.Value) []box.Value { return vs }

	invalid := []struct {
		name string
		def  constraints.ConstraintDefinition
		args []box.Value
		msg  string
	}{
		{"regexp invalid pattern", str["regexp"], args(sv("[")), "invalid regexp pattern: error parsing regexp: missing closing ]: `[`"},
		{"regexp wrong kind", str["regexp"], args(n(1)), "expected string, got number"},
		{"minlength wrong kind", str["minlength"], args(sv("q")), "expected number, got string"},
		{"minlength no argument", str["minlength"], nil, "minlength constraint requires 1 argument"},
		{"email with an argument", str["email"], args(sv("x")), "email constraint takes no arguments"},
		{"one_of no argument", str["one_of"], nil, "one_of constraint requires at least 1 argument"},
		{"one_of second argument wrong kind", str["one_of"], args(sv("a"), n(1)), "expected string, got number"},
		{"range one argument", num["range"], args(n(1)), "range constraint requires 2 arguments"},
		{"range wrong kind", num["range"], args(n(1), sv("x")), "expected number, got string"},
		{"range minimum above maximum", num["range"], args(n(5), n(1)), "range minimum 5 is greater than maximum 1"},
		{"multiple_of zero", num["multiple_of"], args(n(0)), "divisor cannot be zero"},
		{"multiple_of wrong kind", num["multiple_of"], args(sv("2")), "expected number, got string"},
		{"in wrong set item", num["in"], args(box.List(args(n(1), sv("x")))), "expected number, got string"},
		{"trinary eq wrong kind", tri["eq"], args(n(1)), "eq constraint expects a boolean argument"},
		{"list not_empty with an argument", list["not_empty"], args(n(1)), "not_empty constraint takes no arguments"},
	}
	for _, tc := range invalid {
		s.Run(tc.name, func() {
			err := tc.def.CheckArgs(tc.args)
			s.Require().Error(err)
			s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
			s.Equal(tc.msg, err.Error())
		})
	}

	valid := []struct {
		name string
		def  constraints.ConstraintDefinition
		args []box.Value
	}{
		{"regexp", str["regexp"], args(sv("^[a-z]+$"))},
		{"minlength", str["minlength"], args(n(2))},
		{"email", str["email"], nil},
		{"one_of", str["one_of"], args(sv("a"), sv("b"))},
		{"range", num["range"], args(n(1), n(1))},
		{"multiple_of", num["multiple_of"], args(n(0.5))},
		{"in number", num["in"], args(n(1))},
		{"in list", num["in"], args(box.List(args(n(1), n(2))))},
		{"trinary eq", tri["eq"], args(box.Trinary(trinary.Unknown))},
	}
	for _, tc := range valid {
		s.Run(tc.name, func() {
			s.NoError(tc.def.CheckArgs(tc.args))
		})
	}
}

// TestCheckerReportsBrokenArgumentForAnyValue checks that a checker reports a
// broken argument before it looks at the value, so a value of the wrong kind
// does not hide it (#273).
func (s *ConstraintsTestSuite) TestCheckerReportsBrokenArgumentForAnyValue() {
	ctx := s.T().Context()
	cases := map[string]error{
		"regexp on a number":     constraints.StringContraintCheckers["regexp"].Checker(ctx, box.Number(1), []box.Value{box.String("[")}),
		"minlength on null":      constraints.StringContraintCheckers["minlength"].Checker(ctx, box.Null(), []box.Value{box.String("q")}),
		"multiple_of on string":  constraints.NumberContraintCheckers["multiple_of"].Checker(ctx, box.String("x"), []box.Value{box.Number(0)}),
		"trinary eq on a number": constraints.TrinaryConstraintCheckers["eq"].Checker(ctx, box.Number(1), []box.Value{box.Number(1)}),
	}
	for name, err := range cases {
		s.Run(name, func() {
			s.Require().Error(err)
			s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
		})
	}
}

// argClass describes the arguments a constraint takes, by what CheckArgs must
// accept and reject. valid is one well-formed argument list; wrong, when set,
// is a list of the right length whose argument has the wrong kind (or, for a
// divisor, a zero).
type argClass struct {
	valid []box.Value
	wrong []box.Value
}

// TestCheckArgsCoversEveryConstraint walks every constraint table and checks
// that CheckArgs accepts a well-formed argument list and rejects a wrong
// count and a wrong-kind argument. Each constraint must be classified in
// argClasses, so a constraint added without argument metadata (or without a
// class here) fails this test instead of silently skipping the check (#273).
func (s *ConstraintsTestSuite) TestCheckArgsCoversEveryConstraint() {
	n := func(f float64) box.Value { return box.Number(f) }
	str := func(v string) box.Value { return box.String(v) }
	one := func(v box.Value) []box.Value { return []box.Value{v} }

	none := argClass{}
	number := argClass{valid: one(n(-1.5)), wrong: one(str("x"))}
	text := argClass{valid: one(str("a")), wrong: one(n(1))}
	classes := map[string]argClass{
		"none":     none,
		"number":   number,
		"string":   text,
		"regexp":   {valid: one(str("^[a-z]+$")), wrong: one(n(1))},
		"divisor":  {valid: one(n(-2)), wrong: one(n(0))},
		"trinary":  {valid: one(box.Trinary(trinary.Unknown)), wrong: one(n(1))},
		"numset":   {valid: one(box.List([]box.Value{n(1), n(2.5)})), wrong: one(str("x"))},
		"range":    {valid: []box.Value{n(-5), n(-1)}, wrong: []box.Value{n(1), str("x")}},
		"variadic": {valid: []box.Value{str("a"), str("b")}, wrong: []box.Value{str("a"), n(1)}},
	}
	classOf := map[string]string{
		"string/length": "number", "string/minlength": "number", "string/maxlength": "number",
		"string/regexp":      "regexp",
		"string/starts_with": "string", "string/ends_with": "string",
		"string/has_substring": "string", "string/not_has_substring": "string",
		"string/email": "none", "string/url": "none", "string/uuid": "none",
		"string/alphanumeric": "none", "string/alpha": "none", "string/numeric": "none",
		"string/lowercase": "none", "string/uppercase": "none", "string/trimmed": "none",
		"string/not_empty": "none",
		"string/one_of":    "variadic", "string/not_one_of": "variadic",
		"number/min": "number", "number/max": "number", "number/eq": "number",
		"number/neq": "number", "number/gt": "number", "number/lt": "number",
		"number/in": "numset", "number/not_in": "numset",
		"number/range":       "range",
		"number/multiple_of": "divisor",
		"number/even":        "none", "number/odd": "none", "number/positive": "none",
		"number/negative": "none", "number/non_negative": "none", "number/non_positive": "none",
		"number/finite": "none", "number/infinite": "none", "number/nan": "none",
		"trinary/eq": "trinary", "trinary/neq": "trinary",
		"trinary/not_unknown": "none", "trinary/is_true": "none", "trinary/is_false": "none",
		"list/not_empty": "none",
	}

	seen := map[string]bool{}
	for kind, table := range allConstraintTables() {
		for name, def := range table {
			key := kind + "/" + name
			seen[key] = true
			className, ok := classOf[key]
			s.Require().True(ok, "%s has no argument class in this test: classify it", key)
			class := classes[className]
			s.Run(key, func() {
				if className == "none" {
					s.NoError(def.CheckArgs(nil))
					s.Error(def.CheckArgs([]box.Value{n(1)}), "a constraint without arguments rejects one")
					return
				}
				s.Require().NotNil(def.Arg, "a constraint with arguments needs Arg metadata")
				s.NoError(def.CheckArgs(class.valid))
				err := def.CheckArgs(class.wrong)
				s.Require().Error(err)
				s.True(errors.Is(err, constraints.ErrInvalidArgument), err.Error())
				// The runtime path (Checker) reports the same failure for a
				// value of any kind, here one that is never the right kind.
				cerr := def.Checker(s.T().Context(), box.Null(), class.wrong)
				s.Require().Error(cerr)
				s.Equal(err.Error(), cerr.Error())
				// Count: no arguments always fails; one too many fails unless variadic.
				s.Error(def.CheckArgs(nil))
				if def.NumArgs > 0 {
					s.Error(def.CheckArgs(append(append([]box.Value{}, class.valid...), class.valid[0])))
				} else {
					s.NoError(def.CheckArgs(append(append([]box.Value{}, class.valid...), class.valid[0])), "a variadic constraint takes more")
				}
			})
		}
	}
	for key := range classOf {
		s.True(seen[key], "%s is classified but is not in a table", key)
	}
}
