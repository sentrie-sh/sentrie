// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// inPolicy wraps body lines in a namespace and policy, indented one level.
func inPolicy(lines ...string) string {
	body := "  " + strings.Join(lines, "\n  ")
	return "namespace t\n\npolicy p {\n" + strings.ReplaceAll(body, "\n  \n", "\n\n") + "\n}\n"
}

// assertFormats checks that src formats to want and that want is stable.
func (s *FormatTestSuite) assertFormats(src, want string) {
	s.T().Helper()
	got, err := Source(s.T().Context(), "case.sentrie", []byte(src))
	s.Require().NoError(err)
	s.Equal(want, string(got))
	again, err := Source(s.T().Context(), "case.sentrie", got)
	s.Require().NoError(err)
	s.Equal(want, string(again), "output is not stable")
}

func (s *FormatTestSuite) TestStatements() {
	cases := map[string][2]string{
		"empty file":     {"\n\n", ""},
		"namespace only": {"namespace a/b;", "namespace a/b\n"},
		"empty policy":   {"namespace t\npolicy p { }", "namespace t\n\npolicy p {}\n"},
		"metadata": {
			"namespace t\npolicy p { title \"T \\\"q\\\"\" description \"D\" version \"1.0.0\" tag \"k\" = \"v\" }",
			inPolicy(`title "T \"q\""`, `description "D"`, `version "1.0.0"`, `tag "k" = "v"`),
		},
		"facts": {
			"namespace t\npolicy p {\nfact a:string as a\nfact b?:  number as c default 1\nfact d: string? @length(3)\n}",
			inPolicy("fact a: string", "fact b?: number as c default 1", "fact d: string? @length(3)"),
		},
		"use": {
			"namespace t\npolicy p {\nuse {now} from @sentrie/time as time\nuse { a,b } from \"./x/y\" as z\n}",
			inPolicy("use { now } from @sentrie/time", `use { a, b } from "./x/y" as z`),
		},
		"lets": {
			"namespace t\npolicy p {\nlet a:list[number] = [1,2]\nlet b : record[string,boolean] = [\"x\", true]\nlet c: dict[number @min(1)] = {}\n}",
			inPolicy("let a: list[number] = [1, 2]", `let b: record[string, boolean] = ["x", true]`, "let c: dict[number @min(1)] = {}"),
		},
		"rules": {
			"namespace t\npolicy p {\nrule a = default false when ( x ) { yield true }\nrule b = import decision r from q/s with x as 1 with y as 2\nrule c = import decision r from s\n  with x as 1\n}",
			inPolicy("rule a = default false when x { yield true }", "rule b = import decision r from q/s with x as 1 with y as 2",
				"rule c = import decision r from s", "  with x as 1"),
		},
		"nullable shape guard": {
			"namespace t\npolicy p {\nrule a = when (x is Foo?) { yield true }\n}",
			inPolicy("rule a = when (x is Foo?) { yield true }"),
		},
		"exports": {
			"namespace t\npolicy p {\nexport decision of a attach x as 1 attach y as 2\nexport decision of b\n  attach x as 1\n}",
			inPolicy("export decision of a attach x as 1 attach y as 2", "export decision of b", "  attach x as 1"),
		},
		"shapes": {
			"namespace t\nshape A { b: string\n c?: number? }\nshape B with A {}\nshape C number @min(0)\nexport shape A\npolicy p { shape D string }",
			"namespace t\n\nshape A {\n  b: string\n  c?: number?\n}\n\nshape B with A {}\n\nshape C number @min(0)\n\nexport shape A\n\npolicy p {\n  shape D string\n}\n",
		},
		"derive": {
			"namespace t\nderive d = (a:number, b?:string):boolean => { yield a }\nexport derive d",
			"namespace t\n\nderive d = (a: number, b?: string): boolean => { yield a }\n\nexport derive d\n",
		},
		"blank lines collapse": {
			"namespace t\npolicy p {\n\n  let a = 1\n\n\n\n  let b = 2\n  let c = 3\n\n}",
			inPolicy("let a = 1", "", "let b = 2", "let c = 3"),
		},
	}
	for name, c := range cases {
		s.Run(name, func() { s.assertFormats(c[0], c[1]) })
	}
}

func (s *FormatTestSuite) TestExpressions() {
	cases := map[string][2]string{
		"literals":         {`let a = [1, 2.50, "s\t\"", true, unknown, null, #]`, `let a = [1, 2.5, "s\t\"", true, unknown, null, #]`},
		"float keeps dot":  {"let a = 3.0", "let a = 3.0"},
		"map":              {`let a = { "k": 1, ["x" + y]: 2 }`, `let a = {"k": 1, ["x" + y]: 2}`},
		"precedence":       {"let a = (a + b) * c - (d - e) / (f.g)", "let a = (a + b) * c - (d - e) / (f.g)"},
		"redundant parens": {"let a = (a * b) + ((c))", "let a = a * b + c"},
		"unary":            {"let a = - (-x) + !y + not (z)", "let a = -(-x) + !y + not z"},
		"negated infix":    {"let a = not (x in y) and x not matches z", "let a = x not in y and x not matches z"},
		"negated is":       {"let a = not (x is defined) or x is not empty", "let a = x is not defined or x is not empty"},
		"ternary":          {"let a = c ? ((d) ? 1 : 2) : 3", "let a = c ? ((d) ? 1 : 2) : 3"},
		"left ternary":     {"let a = (1 ? b : c) ? d : e", "let a = (1 ? b : c) ? d : e"},
		"elvis":            {"let a = x ?: y + 1", "let a = x ?: y + 1"},
		"elvis then eq":    {"let a = x ?: y == z", "let a = ((x) ?: y) == z"},
		"cast":             {"let a = (x as number).y + z as string", "let a = (x as number).y + z as string"},
		"is forms":         {"let a = [x is string?, x is a / b, x is list[number], x is (1)]", "let a = [x is string?, x is a/b, x is list[number], x is 1]"},
		"calls":            {"let a = f(1,2)!60 + g()! + ns/fn(x)", "let a = f(1, 2)!60 + g()! + ns/fn(x)"},
		"slash quirk":      {"let a = a / b.c + (a / b)[0]", "let a = a / b.c + a / b[0]"},
		"access":           {"let a = x.y[0](1).z", "let a = x.y[0](1).z"},
		"pipeline":         {"let a = x |> f(#, 1) |> g()!", "let a = x |> f(#, 1) |> g()!"},
		"pipeline in op":   {"let a = (x |> f()) + 1", "let a = (x |> f()) + 1"},
		"transform":        {`let a = transform x with "t"`, `let a = transform x with "t"`},
		"lambdas":          {"let a = f(() => { yield 1 }, (x: number, y?) => { yield x })", "let a = f(() => { yield 1 }, (x: number, y?) => { yield x })"},
		"heredoc":          {"let a = <<<TAG\nline\nTAG", "let a = <<<TAG\nline\nTAG"},
	}
	for name, c := range cases {
		s.Run(name, func() {
			s.assertFormats("namespace t\npolicy p {\n"+c[0]+"\n}", inPolicy(c[1]))
		})
	}
}

func (s *FormatTestSuite) TestLineBreaks() {
	cases := map[string][2]string{
		"block":       {"let a = { let b = 1 yield b }", "let a = {\n    let b = 1\n    yield b\n  }"},
		"broken list": {"let a = [1,\n 2, 3]", "let a = [\n    1,\n    2,\n    3\n  ]"},
		"broken args": {"let a = f(\n x)", "let a = f(\n    x\n  )"},
		"lambda arg stays inline": {
			"let a = reduce(xs, 0, (acc, x) => {\n yield acc + x\n})",
			"let a = reduce(xs, 0, (acc, x) => {\n    yield acc + x\n  })",
		},
		"pipeline chain":  {"let a = x\n |> f()\n |> g()", "let a = x\n    |> f()\n    |> g()"},
		"infix after op":  {"let a = x and\n y", "let a = x and\n    y"},
		"infix before op": {"let a = x\n and y", "let a = x and\n    y"},
	}
	for name, c := range cases {
		s.Run(name, func() {
			s.assertFormats("namespace t\npolicy p {\n"+c[0]+"\n}", inPolicy(c[1]))
		})
	}
}

func (s *FormatTestSuite) TestComments() {
	cases := map[string][2]string{
		"top level": {
			"-- header\n\nnamespace t -- ns\n-- about p\npolicy p { -- open\n}",
			"-- header\n\nnamespace t -- ns\n\n-- about p\npolicy p { -- open\n}\n",
		},
		"policy body": {
			"namespace t\npolicy p {\n  -- lead\n  fact a: string -- after fact\n  -- between\n\n  -- before let\n  let b = 1\n}",
			inPolicy("-- lead", "fact a: string -- after fact", "-- between", "", "-- before let", "let b = 1"),
		},
		"empty comment": {"namespace t\npolicy p {\n--\nlet a = 1\n}", inPolicy("--", "let a = 1")},
		"block body": {
			"namespace t\npolicy p {\nrule r = { -- open\n -- lead\n let a = 1 -- ta\n yield a -- ty\n}\n}",
			inPolicy("rule r = { -- open", "  -- lead", "  let a = 1 -- ta", "  yield a -- ty", "}"),
		},
		"shape body": {
			"namespace t\nshape S { -- open\n  -- lead\n  a: string -- ta\n\n  b: number\n}",
			"namespace t\n\nshape S { -- open\n  -- lead\n  a: string -- ta\n\n  b: number\n}\n",
		},
		"expressions": {
			"namespace t\npolicy p {\nlet a = 5 -- t5\n + -- op\n -- line\n 3 -- t3\nlet b = c ? -- q\n 1 : 2\n}",
			inPolicy("let a = 5 -- t5", "  + -- op", "  -- line", "  3 -- t3", "let b = c ? -- q", "  1 : 2"),
		},
		"trailing comment on operator expression": {
			"namespace t\npolicy p {\nlet a = (x + y) -- sum\nlet b = f(x) -- call\n}",
			inPolicy("let a = (x + y) -- sum", "let b = f(x) -- call"),
		},
		"list items": {
			"namespace t\npolicy p {\nlet a = [1 -- one\n, 2, -- before three\n 3, f(x) -- fx\n, -4]\n}",
			inPolicy("let a = [", "  1 -- one", "  2, -- before three", "  3,", "  f(x) -- fx", "  , -4", "]"),
		},
	}
	for name, c := range cases {
		s.Run(name, func() { s.assertFormats(c[0], c[1]) })
	}
}

func (s *FormatTestSuite) TestUnprintableNodeFails() {
	rng := tokens.Range{}
	prg := &ast.Program{Statements: []ast.Statement{
		ast.NewNamespaceStatement(ast.NewFQN([]string{"t"}, rng), rng),
		ast.NewVarDeclaration("a", nil, ast.NewWithClause("w", ast.NewIdentifier("x", rng), rng), rng),
	}}
	_, err := Program(prg, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot print expression *ast.WithClause")

	prg.Statements[1] = ast.NewVarDeclaration("a", nil, ast.NewInfixExpression(ast.NewIdentifier("x", rng), ast.NewIdentifier("y", rng), "**", rng), rng)
	_, err = Program(prg, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), `cannot print operator "**"`)
}

func (s *FormatTestSuite) TestSourceRejectsUnparsableInput() {
	_, err := Source(s.T().Context(), "bad.sentrie", []byte("namespace t\npolicy p { let = }"))
	s.Error(err)
}

func (s *FormatTestSuite) TestEqual() {
	parse := func(src string) *ast.Program {
		prg, err := parse(s.T().Context(), "eq.sentrie", []byte(src))
		s.Require().NoError(err)
		return prg
	}
	base := parse("namespace t\npolicy p {\n  -- note\n  let a = 1 + 2\n}")
	s.NoError(Equal(base, parse("namespace t policy p {\n-- note\nlet a=1+2 }")), "positions and layout are ignored")
	s.ErrorContains(Equal(base, parse("namespace t\npolicy p {\n  -- other\n  let a = 1 + 2\n}")), `"note" differs from "other"`)
	s.Error(Equal(base, parse("namespace t\npolicy p {\n  let a = 1 + 2\n}")), "a dropped comment is a difference")
	s.Error(Equal(base, parse("namespace t\npolicy p {\n  -- note\n  let a = 1 - 2\n}")))
	s.NoError(Equal(nil, nil))
	s.Error(Equal(base, nil))
}
