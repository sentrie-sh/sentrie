// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import "strings"

// assertStable formats src, requires that the result parses back to an equal
// program (Source checks that itself) and that formatting it again changes
// nothing. It returns the formatted text.
func (s *FormatTestSuite) assertStable(src string) string {
	s.T().Helper()
	got, err := Source(s.T().Context(), "case.sentrie", []byte(src))
	s.Require().NoError(err)
	again, err := Source(s.T().Context(), "case.sentrie", got)
	s.Require().NoError(err)
	s.Equal(string(got), string(again), "output is not stable")
	return string(got)
}

func (s *FormatTestSuite) TestCommentPlacement() {
	cases := map[string][2]string{
		"comment-only file keeps runs and collapses blank lines": {
			"-- a\n-- b\n\n\n\n-- c\n",
			"-- a\n-- b\n\n-- c\n",
		},
		"comments at the end of the file": {
			"namespace t\n-- tail one\n\n\n-- tail two\n",
			"namespace t\n\n-- tail one\n\n-- tail two\n",
		},
		"comment between policies": {
			"namespace t\npolicy p {\n}\n-- after policy\npolicy q {}",
			"namespace t\n\npolicy p {}\n\n-- after policy\npolicy q {}\n",
		},
		"comment-only policy body": {
			"namespace t\npolicy p {\n  -- only comment\n}",
			inPolicy("-- only comment"),
		},
		"export shape with trailing and following comment": {
			"namespace t\nshape S { a: number -- c\n}\nexport shape S -- done\n-- end",
			"namespace t\n\nshape S {\n  a: number -- c\n}\n\nexport shape S -- done\n\n-- end\n",
		},
		"rule closing brace and export": {
			"namespace t\npolicy p {\nrule r = { yield true } -- close\nexport decision of r -- exp\n-- last\n}",
			inPolicy("rule r = { yield true } -- close", "export decision of r -- exp", "-- last"),
		},
	}
	for name, c := range cases {
		s.Run(name, func() { s.assertFormats(c[0], c[1]) })
	}
}

// TestRoundTripConstructs formats constructs that are easy to print wrongly.
// Source already requires the output to parse to an equal program, so these
// check that no construct is rejected and that the output is stable.
func (s *FormatTestSuite) TestRoundTripConstructs() {
	exprs := map[string]string{
		"left grouped ternary chain":  "c ? 1 : d ? 2 : 3",
		"elvis chain":                 "x ?: y ?: z",
		"elvis in comparison":         "x ?: y == z",
		"divide chain":                "a / b / c",
		"divide by group":             "a / (b + c)",
		"subtract group":              "a - (b - c)",
		"negated membership":          "not (x in y) or not (x is defined)",
		"negated matches":             "not (x matches y)",
		"pipeline of lambdas":         "xs |> map((x) => { yield x + 1 }) |> filter((x: number) => { yield x > 1 })",
		"typed nullable lambda":       "f((x: string?, y?: list[number]) => { yield x })",
		"cast then access":            "(x as document).y[0]",
		"heredoc argument":            "f(<<<T\nbody \"q\" \\n\nT\n, 2)",
		"heredoc after text":          "[1, <<<TAG\nline\nTAG\n]",
		"quoted newline stays quoted": `"line\nbreak"`,
		"import with expressions":     "[(1 + 2) * 3, -(a + b), !(c or d)]",
	}
	for name, e := range exprs {
		s.Run(name, func() {
			out := s.assertStable("namespace t\npolicy p {\nlet a = " + e + "\n}")
			if strings.Contains(e, "<<<") {
				s.Contains(out, "<<<"+map[bool]string{true: "T", false: "TAG"}[strings.Contains(e, "<<<T\n")], "heredoc tag lost")
			}
		})
	}

	decls := map[string]string{
		"typed lets":                   "let a: string? = null\nlet b: list[string?] = []\nlet c: record[string, number?] = {}",
		"rule import with expressions": "rule r = import decision d from a/b with x as (1 + 2) with y as f(z)\nexport decision of r attach a as [1, 2]",
	}
	for name, d := range decls {
		s.Run(name, func() { s.assertStable("namespace t\npolicy p {\n" + d + "\n}") })
	}
}
