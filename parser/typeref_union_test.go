// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) parseTypeRefSrc(src string) ast.TypeRef {
	p := NewParserFromString(src, "union.sentrie")
	ref := parseTypeRef(s.T().Context(), p)
	s.Require().NoError(p.err, src)
	s.Require().NotNil(ref, src)
	return ref
}

// TestParseUnionTypeRef covers `|` precedence: looser than the `?` and
// constraint suffixes and than brackets.
func (s *ParserTestSuite) TestParseUnionTypeRef() {
	s.Run("two members", func() {
		u, ok := s.parseTypeRefSrc("number | list[number]").(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.Require().Len(u.Members, 2)
		s.IsType(&ast.NumberTypeRef{}, u.Members[0])
		s.IsType(&ast.ListTypeRef{}, u.Members[1])
		s.Equal("number | list[number]", u.String())
	})
	s.Run("nullable suffix binds to its member", func() {
		u, ok := s.parseTypeRefSrc("string | number?").(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.IsType(&ast.StringTypeRef{}, u.Members[0])
		s.IsType(&ast.NullableTypeRef{}, u.Members[1])
		s.Equal("string | number?", u.String())
		s.True(ast.AcceptsNull(u))
	})
	s.Run("constraints bind to the member they follow", func() {
		u, ok := s.parseTypeRefSrc("string@minlength(3) | number@min(10) | trinary").(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.Require().Len(u.Members, 3)
		s.Require().Len(u.Members[0].GetConstraints(), 1)
		s.Equal("minlength", u.Members[0].GetConstraints()[0].Name)
		s.Require().Len(u.Members[1].GetConstraints(), 1)
		s.Equal("min", u.Members[1].GetConstraints()[0].Name)
		s.Empty(u.Members[2].GetConstraints())
		s.Empty(u.GetConstraints())
		s.False(ast.AcceptsNull(u))
	})
	s.Run("union inside brackets", func() {
		l, ok := s.parseTypeRefSrc("list[string | number]").(*ast.ListTypeRef)
		s.Require().True(ok)
		s.IsType(&ast.UnionTypeRef{}, l.ElemType)
		r, ok := s.parseTypeRefSrc("record[string | number, trinary]").(*ast.RecordTypeRef)
		s.Require().True(ok)
		s.Require().Len(r.Fields, 2)
		s.IsType(&ast.UnionTypeRef{}, r.Fields[0])
	})
	s.Run("shape members", func() {
		u, ok := s.parseTypeRefSrc("User | com/ex/Admin").(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.IsType(&ast.ShapeTypeRef{}, u.Members[0])
		s.IsType(&ast.ShapeTypeRef{}, u.Members[1])
	})
	s.Run("missing member", func() {
		p := NewParserFromString("string |", "union.sentrie")
		s.Nil(parseTypeRef(s.T().Context(), p))
		s.Error(p.err)
	})
}

func (s *ParserTestSuite) TestParseAnyTypeRef() {
	s.Run("bare any is the top type", func() {
		ref := s.parseTypeRefSrc("any")
		s.IsType(&ast.AnyTypeRef{}, ref)
		s.Equal("any", ref.String())
		s.True(ast.AcceptsNull(ref))
	})
	s.Run("qualified any is a shape", func() {
		s.IsType(&ast.ShapeTypeRef{}, s.parseTypeRefSrc("com/ex/any"))
	})
	s.Run("any in a union and a list", func() {
		u, ok := s.parseTypeRefSrc("list[any] | any").(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.IsType(&ast.AnyTypeRef{}, u.Members[0].(*ast.ListTypeRef).ElemType)
		s.IsType(&ast.AnyTypeRef{}, u.Members[1])
	})
	s.Run("any takes no constraints", func() {
		p := NewParserFromString("any@minlength(1)", "union.sentrie")
		s.Nil(parseTypeRef(s.T().Context(), p))
		s.ErrorContains(p.err, "minlength")
	})
	s.Run("shape cannot be named any", func() {
		p := NewParserFromString("namespace com/ex\nshape any { a: string }\n", "union.sentrie")
		_, err := p.ParseProgram(s.T().Context())
		s.ErrorContains(err, `cannot declare a shape named "any"`)
	})
	s.Run("any quantifier still parses", func() {
		call, ok := s.parseIsExpr("any(xs, (x) => { yield x > 1 })").(*ast.CallExpression)
		s.Require().True(ok)
		s.NotNil(call)
	})
}

func (s *ParserTestSuite) TestParseIsUnionAndAny() {
	s.Run("type keyword union", func() {
		is, ok := s.parseIsExpr("x is string | number").(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(is.Right)
		s.IsType(&ast.UnionTypeRef{}, is.Type)
	})
	s.Run("name union is a type", func() {
		is, ok := s.parseIsExpr("x is User | com/ex/Admin").(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(is.Right)
		s.IsType(&ast.UnionTypeRef{}, is.Type)
	})
	s.Run("name with suffix then union", func() {
		is, ok := s.parseIsExpr("x is User? | string").(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(is.Right)
		u, ok := is.Type.(*ast.UnionTypeRef)
		s.Require().True(ok)
		s.IsType(&ast.NullableTypeRef{}, u.Members[0])
	})
	s.Run("any is a type, not a value", func() {
		is, ok := s.parseIsExpr("x is any").(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(is.Right)
		s.IsType(&ast.AnyTypeRef{}, is.Type)
	})
	s.Run("union then ternary", func() {
		t, ok := s.parseIsExpr("x is string | number ? 1 : 2").(*ast.TernaryExpression)
		s.Require().True(ok)
		s.IsType(&ast.IsExpression{}, t.Condition)
	})
	s.Run("bare pipe in an expression is an error", func() {
		p := NewParserFromString("a | b", "union.sentrie")
		p.parseExpression(s.T().Context(), LOWEST)
		s.ErrorContains(p.err, "only '|>' is supported")
	})
}
