// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// TestParseUnionInEveryTypePosition parses a program that puts a union and
// `any` in each place a type reference may appear.
func (s *ParserTestSuite) TestParseUnionInEveryTypePosition() {
	src := `namespace com/ex
shape S { a: string | number  b: any }
shape Alias string | number

derive d = (p: string | number, q?: any): number | list[number] => { yield 1 }

policy pol {
  fact f?: number | list[number]
  fact g: any
  let x: string | number = 1
  let y = x as number | string
  let z = [1] is number | list[number]
  rule r = { yield any([1], (n) => { yield n is any }) }
  export decision of r
}
`
	p := NewParserFromString(src, "positions.sentrie")
	prog, err := p.ParseProgram(s.T().Context())
	s.Require().NoError(err)
	s.Require().NotNil(prog)
	var shapes []*ast.ShapeStatement
	var derive *ast.DeriveStatement
	var policy *ast.PolicyStatement
	for _, st := range prog.Statements {
		switch n := st.(type) {
		case *ast.ShapeStatement:
			shapes = append(shapes, n)
		case *ast.DeriveStatement:
			derive = n
		case *ast.PolicyStatement:
			policy = n
		}
	}
	s.Require().Len(shapes, 2)
	s.Require().NotNil(derive)
	s.Require().NotNil(policy)

	s.Require().NotNil(shapes[0].Complex)
	s.Equal("string | number", shapes[0].Complex.Fields["a"].Type.String())
	s.IsType(&ast.AnyTypeRef{}, shapes[0].Complex.Fields["b"].Type)
	s.IsType(&ast.UnionTypeRef{}, shapes[1].Simple)

	s.Require().Len(derive.Value.ParamTypes, 2)
	s.IsType(&ast.UnionTypeRef{}, derive.Value.ParamTypes[0])
	s.IsType(&ast.AnyTypeRef{}, derive.Value.ParamTypes[1])
	s.Equal("number | list[number]", derive.Value.ReturnType.String())

	facts := map[string]ast.TypeRef{}
	for _, st := range policy.Statements {
		if f, ok := st.(*ast.FactStatement); ok {
			facts[f.Name] = f.Type
		}
	}
	s.Require().Len(facts, 2)
	s.Equal("number | list[number]", facts["f"].String())
	s.IsType(&ast.AnyTypeRef{}, facts["g"])
}

func (s *ParserTestSuite) TestParseUnionRejectsMalformed() {
	cases := map[string]string{
		"trailing bar":           "string |",
		"double bar":             "string | | number",
		"leading bar":            "| string",
		"bar then close bracket": "list[string |]",
		"bar then constraint":    "string | @min(1)",
	}
	for name, src := range cases {
		s.Run(name, func() {
			p := NewParserFromString(src, "union.sentrie")
			ref := parseTypeRef(s.T().Context(), p)
			s.True(ref == nil || p.err != nil, "%q should not parse cleanly", src)
			s.Error(p.err, src)
		})
	}
}

// TestBarePipeStillRejectedInExpressions: `|` is a type operator only; the
// expression grammar still has just `|>`.
func (s *ParserTestSuite) TestBarePipeStillRejectedInExpressions() {
	for _, src := range []string{"a | b", "1 | 2", "(a | b)", "f(a | b)", "a |> f() | g()"} {
		s.Run(src, func() {
			p := NewParserFromString(src, "union.sentrie")
			p.parseExpression(s.T().Context(), LOWEST)
			s.Require().Error(p.err, src)
			s.Contains(p.err.Error(), "only '|>' is supported")
		})
	}
	s.Run("let initialiser", func() {
		p := NewParserFromString("namespace a\npolicy p {\n  let x = 1 | 2\n  rule r = { yield true }\n  export decision of r\n}\n", "union.sentrie")
		_, err := p.ParseProgram(s.T().Context())
		s.Require().Error(err)
		s.Contains(err.Error(), "only '|>' is supported")
	})
	s.Run("pipeline forward still lowers", func() {
		p := NewParserFromString("value |> len()", "union.sentrie")
		expr := p.parseExpression(s.T().Context(), LOWEST)
		s.Require().NoError(p.err)
		s.Equal("len(value)", expr.String())
	})
	s.Run("pipeline then is union", func() {
		is, ok := s.parseIsExpr("(value |> len()) is number | string").(*ast.IsExpression)
		s.Require().True(ok)
		s.IsType(&ast.UnionTypeRef{}, is.Type)
	})
}
