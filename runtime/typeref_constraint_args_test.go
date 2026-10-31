// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strconv"

	"github.com/sentrie-sh/sentrie/trinary"
)

// evalRuleBody evaluates a rule of com/app/pol on idx directly, without
// validating, so a pack the index would reject still reaches the runtime's
// own argument check.
func (s *RuntimeTestSuite) evalUnvalidatedRule(src, rule string) (trinary.Value, error) {
	ctx := s.T().Context()
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "args.sentrie", src: src})
	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	exec, ok := rawExec.(*executorImpl)
	s.Require().True(ok)
	pol, err := idx.ResolvePolicy("com/app", "pol")
	s.Require().NoError(err)
	ec := NewExecutionContext(pol, exec)
	defer ec.Dispose()
	got, _, err := eval(ctx, ec, exec, pol, pol.Rules[rule].Body)
	if err != nil {
		return trinary.Unknown, err
	}
	tv, ok := got.TrinaryValue()
	s.Require().True(ok, "rule %s did not yield a trinary", rule)
	return tv, nil
}

// TestBrokenConstraintArgumentErrorsForDocumentAndListValues extends the #273
// reproduction to the other value kinds: a document and a list reach a string,
// number or trinary constraint with a broken argument, and both `is` and
// `is not` report the argument instead of reading a kind mismatch.
func (s *RuntimeTestSuite) TestBrokenConstraintArgumentErrorsForDocumentAndListValues() {
	values := map[string]string{"document": `{"v": 1}`, "list": `[1, 2]`, "number": `1`, "string": `"a"`, "null": `null`}
	types := map[string]string{
		"regexp":        `string@regexp("[")`,
		"wrong kind":    `string@minlength("q")`,
		"zero divisor":  `number@multiple_of(0)`,
		"range":         `number@range(5, 1)`,
		"trinary":       `trinary@eq(1)`,
		"nullable null": `string?@regexp("[")`,
	}
	for vname, v := range values {
		for tname, t := range types {
			for _, op := range []string{"is", "is not"} {
				s.Run(vname+"/"+tname+"/"+op, func() {
					src := "namespace com/app\npolicy pol {\n  rule r = { yield " + v + " " + op + " " + t + " }\n  export decision of r\n}\n"
					got, err := s.evalUnvalidatedRule(src, "r")
					s.Require().Error(err, "got %v", got)
					s.False(isTypeMismatch(err), err.Error())
				})
			}
		}
	}
}

// TestBrokenConstraintArgumentInNestedShapeFieldThroughDerive checks a broken
// argument on a shape field reached through a derive parameter: the shape
// check reaches the field with a value of the wrong kind, and the field's
// argument is reported rather than the kind mismatch.
func (s *RuntimeTestSuite) TestBrokenConstraintArgumentInNestedShapeFieldThroughDerive() {
	src := `namespace com/app
shape Inner { f: string@regexp("[") }
shape Outer { inner: Inner }
derive d = (a: Outer): trinary => { yield true }
policy pol {
  rule wrong_kind = { yield d({ "inner": { "f": 1 } }) }
  rule right_kind = { yield d({ "inner": { "f": "a" } }) }
  rule is_check = { yield { "inner": { "f": 1 } } is not Outer }
  export decision of wrong_kind
}
`
	for _, rule := range []string{"wrong_kind", "right_kind", "is_check"} {
		s.Run(rule, func() {
			_, err := s.evalUnvalidatedRule(src, rule)
			s.Require().Error(err)
			s.Contains(err.Error(), "invalid regexp pattern")
			s.False(isTypeMismatch(err), err.Error())
		})
	}
}

// TestValidConstraintArgumentsEvaluateUnchanged checks there are no false
// positives: well-formed arguments (fractional, lists, and a bound
// equal on both sides of a range) validate and still accept and reject values
// as they did before arguments were checked up front.
func (s *RuntimeTestSuite) TestValidConstraintArgumentsEvaluateUnchanged() {
	ctx := s.T().Context()
	cases := []struct {
		name, expr string
		want       trinary.Value
	}{
		{"range holds", `3 is number@range(1, 5)`, trinary.True},
		{"range rejects", `0 is number@range(1, 5)`, trinary.False},
		{"equal range bounds", `2 is number@range(2, 2)`, trinary.True},
		{"fractional range", `1.5 is number@range(0.5, 2.5)`, trinary.True},
		{"float min", `0.75 is number@min(0.5)`, trinary.True},
		{"float max rejects", `0.75 is number@max(0.5)`, trinary.False},
		{"float multiple", `0.5 is number@multiple_of(0.25)`, trinary.True},
		{"multiple rejects", `5 is number@multiple_of(2)`, trinary.False},
		{"not_in list with a float", `2.5 is number@not_in([1, 2.5])`, trinary.False},
		{"not_in single number", `2 is number@not_in(1)`, trinary.True},
		{"regexp holds", `"ab" is string@regexp("^a")`, trinary.True},
		{"regexp rejects", `"ba" is string@regexp("^a")`, trinary.False},
		{"one_of", `"b" is string@one_of("a", "b")`, trinary.True},
		{"minlength", `"a" is string@minlength(2)`, trinary.False},
		{"trinary eq unknown", `unknown is trinary@eq(unknown)`, trinary.True},
		{"wrong kind is a mismatch", `1 is not string@regexp("^a")`, trinary.True},
		{"nullable null skips constraints", `null is string?@regexp("^a")`, trinary.True},
		{"list item type", `["a", "ab"] is list[string@minlength(1)]`, trinary.True},
		{"no arguments", `"a@b.co" is string@email()`, trinary.True},
	}
	var rules, exports string
	for i, tc := range cases {
		rules += "  rule r" + itoa(i) + " = { yield " + tc.expr + " }\n"
		exports += "  export decision of r" + itoa(i) + "\n"
	}
	src := "namespace com/app\npolicy pol {\n" + rules + exports + "}\n"
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "valid_args.sentrie", src: src})
	for i, tc := range cases {
		s.Run(tc.name, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", "r"+itoa(i), map[string]any{})
			s.Require().NoError(err)
			s.Equal(tc.want, out.Decision.State)
		})
	}
}

func itoa(i int) string { return strconv.Itoa(i) }
