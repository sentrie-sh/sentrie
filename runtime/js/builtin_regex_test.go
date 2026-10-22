// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

// TestRegexReplaceFirstAndAll pins #209: replace rewrites only the leftmost
// match, replaceAll rewrites every match, and both expand $n references.
func (s *JSTestSuite) TestRegexReplaceFirstAndAll() {
	cases := []struct {
		expr string
		want string
	}{
		{`m.replace("a", "aaa", "b")`, "baa"},
		{`m.replaceAll("a", "aaa", "b")`, "bbb"},
		{`m.replace("a", "xyz", "b")`, "xyz"},
		{`m.replaceAll("a", "xyz", "b")`, "xyz"},
		{`m.replace("o+", "foo boo", "0")`, "f0 boo"},
		{`m.replaceAll("o+", "foo boo", "0")`, "f0 b0"},
		{`m.replace("(\\w+)@(\\w+)", "a@b c@d", "$2@$1")`, "b@a c@d"},
		{`m.replaceAll("(\\w+)@(\\w+)", "a@b c@d", "$2@$1")`, "b@a d@c"},
		{`m.replace("^", "abc", ">")`, ">abc"},
		{`m.replace("x*", "abc", "-")`, "-abc"},
		{`m.replaceAll("x*", "abc", "-")`, "-a-b-c-"},
	}
	for _, tc := range cases {
		s.Run(tc.expr, func() {
			got, err := s.runBuiltinScript(BuiltinRegexGo, tc.expr)
			s.Require().NoError(err)
			s.Equal(tc.want, got.String())
		})
	}
}

// TestRegexReplaceThrowsOnBadPattern keeps both replace forms on the
// throwing convention from #111.
func (s *JSTestSuite) TestRegexReplaceThrowsOnBadPattern() {
	for _, expr := range []string{`m.replace("(", "x", "y")`, `m.replaceAll("(", "x", "y")`} {
		s.Run(expr, func() {
			_, err := s.runBuiltinScript(BuiltinRegexGo, expr)
			s.Require().Error(err)
			s.Contains(err.Error(), "missing closing )")
		})
	}
}
