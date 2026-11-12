// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestConstraintChecksEvaluate runs the #146 constraint behaviours through `is`.
func (s *RuntimeTestSuite) TestConstraintChecksEvaluate() {
	ctx := s.T().Context()
	cases := []struct {
		name, expr string
		want       trinary.Value
	}{
		{"number in holds", `2 is number@in(1, 2, 3)`, trinary.True},
		{"number in rejects", `5 is number@in(1, 2, 3)`, trinary.False},
		{"number not_in holds", `5 is number@not_in(1, 2)`, trinary.True},
		{"number not_in rejects", `1 is number@not_in(1, 2)`, trinary.False},
		{"string in holds", `"b" is string@in("a", "b")`, trinary.True},
		{"string in rejects", `"z" is string@in("a", "b")`, trinary.False},
		{"string not_in holds", `"z" is string@not_in("a", "b")`, trinary.True},
		{"string not_in rejects", `"a" is string@not_in("a", "b")`, trinary.False},
		{"url default http", `"http://example.com" is string@url()`, trinary.True},
		{"url default https", `"https://example.com" is string@url()`, trinary.True},
		{"url default rejects ftp", `"ftp://example.com" is string@url()`, trinary.False},
		{"url localhost port", `"http://localhost:8080/x" is string@url()`, trinary.True},
		{"url ipv4", `"http://127.0.0.1:80" is string@url()`, trinary.True},
		{"url ipv6", `"http://[::1]:80/" is string@url()`, trinary.True},
		{"url no scheme", `"localhost:8080" is string@url()`, trinary.False},
		{"url no host", `"http://:80" is string@url()`, trinary.False},
		{"url path only", `"/path" is string@url()`, trinary.False},
		{"url ftp only holds", `"ftp://a.co" is string@url("ftp")`, trinary.True},
		{"url ftp only rejects http", `"http://a.co" is string@url("ftp")`, trinary.False},
		{"url three schemes", `"ftp://a.co" is string@url("http", "https", "ftp")`, trinary.True},
		{"url scheme case", `"FTP://a.co" is string@url("ftp")`, trinary.True},
		{"length precomposed one", `"é" is string@length(1)`, trinary.True},
		{"length precomposed not two", `"é" is string@length(2)`, trinary.False},
		{"minlength code points", `"é" is string@minlength(2)`, trinary.False},
		{"maxlength code points", `"é" is string@maxlength(1)`, trinary.True},
		{"min negative holds", `-3 is number@min(-5)`, trinary.True},
		{"min negative rejects", `-6 is number@min(-5)`, trinary.False},
		{"range negative holds", `-3 is number@range(-10, -1)`, trinary.True},
		{"range negative rejects", `0 is number@range(-10, -1)`, trinary.False},
		{"gt negative float", `-0.25 is number@gt(-0.5)`, trinary.True},
		{"gt negative float rejects", `-0.75 is number@gt(-0.5)`, trinary.False},
		{"in negative", `-1 is number@in(-1, 2)`, trinary.True},
		{"multiple_of negative value", `-6 is number@multiple_of(3)`, trinary.True},
		{"multiple_of negative both", `-6 is number@multiple_of(-3)`, trinary.True},
		{"multiple_of negative rejects", `-7 is number@multiple_of(3)`, trinary.False},
		{"multiple_of negative float rejects", `-3.5 is number@multiple_of(2)`, trinary.False},
	}
	var rules, exports string
	for i, tc := range cases {
		rules += "  rule r" + itoa(i) + " = { yield " + tc.expr + " }\n"
		exports += "  export decision of r" + itoa(i) + "\n"
	}
	src := "namespace com/app\npolicy pol {\n" + rules + exports + "}\n"
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "constraint_checks.sentrie", src: src})
	for i, tc := range cases {
		s.Run(tc.name, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", "r"+itoa(i), map[string]any{})
			s.Require().NoError(err)
			s.Equal(tc.want, out.Decision.State)
		})
	}
}

// TestInListFormFailsAtRuntime checks the old list form reports the argument
// error at runtime instead of comparing against the list (#146).
func (s *RuntimeTestSuite) TestInListFormFailsAtRuntime() {
	for _, expr := range []string{
		`1 is number@in([1, 2])`, `1 is not number@in([1, 2])`,
		`1 is number@not_in([1, 2])`, `"a" is string@in(["a"])`, `"a" is string@not_in(["a"])`,
	} {
		src := "namespace com/app\npolicy pol {\n  rule r = { yield " + expr + " }\n  export decision of r\n}\n"
		got, err := s.evalUnvalidatedRule(src, "r")
		s.Require().Error(err, "%s: got %v", expr, got)
		s.Contains(err.Error(), "takes each value as its own argument", expr)
		s.False(isTypeMismatch(err), expr)
	}
}

// TestBadURLSchemeArgumentFailsAtRuntime covers `@url("ftp://")`.
func (s *RuntimeTestSuite) TestBadURLSchemeArgumentFailsAtRuntime() {
	src := "namespace com/app\npolicy pol {\n  rule r = { yield \"ftp://a.co\" is string@url(\"ftp://\") }\n  export decision of r\n}\n"
	_, err := s.evalUnvalidatedRule(src, "r")
	s.Require().Error(err)
	s.Contains(err.Error(), `invalid URL scheme "ftp://"`)
	s.False(isTypeMismatch(err))
}
