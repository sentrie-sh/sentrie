// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *FormatTestSuite) TestCommentAfterCommaStaysOnItsLine() {
	cases := map[string][2]string{
		"call arguments": {"let a = f(1, -- one\n 2)", "let a = f(\n    1, -- one\n    2\n  )"},
		"list items":     {"let b = [1, -- first item\n 2, 3]", "let b = [\n    1, -- first item\n    2,\n    3\n  ]"},
		"after bracket":  {"let c = [ -- head\n 1]", "let c = [ -- head\n    1\n  ]"},
		"comment run":    {"let d = [1, -- one\n -- two\n 2]", "let d = [\n    1, -- one\n    -- two\n    2\n  ]"},
	}
	for name, c := range cases {
		s.Run(name, func() {
			s.assertFormats("namespace t\npolicy p {\n"+c[0]+"\n}", inPolicy(c[1]))
		})
	}
}

func (s *FormatTestSuite) TestCommentTextKeepsInnerIndentation() {
	src := "namespace t\n--    indented\n--no space\n--   \nshape S {\n  --   in shape\n  a: string\n}\npolicy p {\n  let a = 1 --   two  \n  let b = --  pre\n    2\n}"
	want := "namespace t\n\n--    indented\n-- no space\n--\nshape S {\n  --   in shape\n  a: string\n}\n\npolicy p {\n  let a = 1 --   two\n  let b = --  pre\n    2\n}\n"
	s.assertFormats(src, want)
}

func (s *FormatTestSuite) TestPrecedingCommentKeepsParentheses() {
	s.assertFormats(
		"namespace t\npolicy p {\nlet x = a + (-- c\n b - 1)\nlet y = (-- d\n a) + b\n}",
		inPolicy("let x = a + ( -- c", "  b - 1)", "let y = ( -- d", "  a) + b"),
	)
}

func (s *FormatTestSuite) TestEqualSeesShapeCommentOrder() {
	parse := func(src string) *ast.Program {
		prg, err := parse(s.T().Context(), "eq.sentrie", []byte(src))
		s.Require().NoError(err)
		return prg
	}
	before := parse("namespace t\nshape S {\n  -- c\n  a: string\n  b: number\n}")
	moved := parse("namespace t\nshape S {\n  a: string\n  -- c\n  b: number\n}")
	s.ErrorContains(Equal(before, moved), "shape members are in order")
	s.NoError(Equal(before, parse("namespace t shape S {\n-- c\n a: string b: number }")))
}

func (s *FormatTestSuite) TestMissingNodesFail() {
	rng := tokens.Range{}
	ns := ast.NewNamespaceStatement(ast.NewFQN([]string{"t"}, rng), rng)

	_, err := Program(&ast.Program{Statements: []ast.Statement{ns, ast.NewVarDeclaration("a", nil, nil, rng)}}, nil)
	s.ErrorContains(err, "cannot print a missing expression")

	var missing *ast.FactStatement
	_, err = Program(&ast.Program{Statements: []ast.Statement{ns, missing}}, nil)
	s.ErrorContains(err, "cannot print a missing statement")

	_, err = Program(&ast.Program{Statements: []ast.Statement{ns, ast.NewFactStatement("f", nil, "f", nil, false, rng)}}, nil)
	s.ErrorContains(err, "cannot print a missing type")
}
