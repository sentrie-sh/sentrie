// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import "github.com/sentrie-sh/sentrie/ast"

// TestParseIsExpressionIncompleteRightOperand does not panic when the right side is missing.
func (s *ParserTestSuite) TestParseIsExpressionIncompleteRightOperand() {
	parser := NewParserFromString("x is", "test.sentrie")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.Error(parser.err)
}

func (s *ParserTestSuite) parseIsExpr(src string) ast.Expression {
	parser := NewParserFromString(src, "is.sentrie")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(parser.err, src)
	s.Require().NotNil(expr, src)
	return expr
}

// TestParseIsTypeForms covers how the right-hand side of `is` is classified.
func (s *ParserTestSuite) TestParseIsTypeForms() {
	s.Run("type keyword is a type", func() {
		is, ok := s.parseIsExpr("x is list[number]").(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(is.Right)
		s.IsType(&ast.ListTypeRef{}, is.Type)
	})
	s.Run("bare name carries both readings", func() {
		is, ok := s.parseIsExpr("x is User").(*ast.IsExpression)
		s.Require().True(ok)
		s.True(is.Ambiguous())
		s.Equal("User", is.Type.String())
	})
	s.Run("qualified name carries both readings", func() {
		is, ok := s.parseIsExpr("x is com/ex/User").(*ast.IsExpression)
		s.Require().True(ok)
		s.True(is.Ambiguous())
		s.Equal("com/ex/User", is.Type.String())
	})
	s.Run("name with a constraint is read as a type", func() {
		// Shapes declare no constraints today, so the constraint is rejected
		// by the type reference, which shows the name was read as a type
		// rather than as a value followed by a stray `@`.
		parser := NewParserFromString(`x is User@minlength(1)`, "is.sentrie")
		parser.parseExpression(s.T().Context(), LOWEST)
		s.Require().Error(parser.err)
		s.Contains(parser.err.Error(), "cannot add constraint minlength")
	})
	s.Run("name with nullable suffix is a type", func() {
		is, ok := s.parseIsExpr("x is User? and y").(*ast.InfixExpression)
		s.Require().True(ok)
		left, ok := is.Left.(*ast.IsExpression)
		s.Require().True(ok)
		s.Nil(left.Right)
		s.IsType(&ast.NullableTypeRef{}, left.Type)
	})
	s.Run("question before an expression is a ternary", func() {
		t, ok := s.parseIsExpr(`x is string ? "a" : "b"`).(*ast.TernaryExpression)
		s.Require().True(ok)
		is, ok := t.Condition.(*ast.IsExpression)
		s.Require().True(ok)
		s.IsType(&ast.StringTypeRef{}, is.Type)
	})
	s.Run("question before a name is a ternary", func() {
		_, ok := s.parseIsExpr(`x is User ? a : b`).(*ast.TernaryExpression)
		s.True(ok)
	})
	s.Run("question at the end is nullable", func() {
		is, ok := s.parseIsExpr("x is string?").(*ast.IsExpression)
		s.Require().True(ok)
		s.IsType(&ast.NullableTypeRef{}, is.Type)
	})
	s.Run("is not wraps the type check", func() {
		u, ok := s.parseIsExpr("x is not number").(*ast.UnaryExpression)
		s.Require().True(ok)
		s.Equal("not", u.Operator)
		s.IsType(&ast.IsExpression{}, u.Right)
	})
	s.Run("other right-hand sides stay equality", func() {
		in, ok := s.parseIsExpr("x is y.z").(*ast.InfixExpression)
		s.Require().True(ok)
		s.Equal("is", in.Operator)
	})
	s.Run("is defined and is empty are unchanged", func() {
		s.IsType(&ast.IsDefinedExpression{}, s.parseIsExpr("x is defined"))
		s.IsType(&ast.IsEmptyExpression{}, s.parseIsExpr("x is empty"))
		u, ok := s.parseIsExpr("x is not defined").(*ast.UnaryExpression)
		s.Require().True(ok)
		s.IsType(&ast.IsDefinedExpression{}, u.Right)
		u, ok = s.parseIsExpr("x is not empty").(*ast.UnaryExpression)
		s.Require().True(ok)
		s.IsType(&ast.IsEmptyExpression{}, u.Right)
	})
}

// TestParseIsQuestionBeforeBrace covers `?` followed by `{`: nullable after a
// type keyword (a rule body follows), the ternary after a name, which may be
// a value and has always been a ternary there.
func (s *ParserTestSuite) TestParseIsQuestionBeforeBrace() {
	s.Run("name then map literal is a ternary", func() {
		t, ok := s.parseIsExpr(`role is admin ? { "x": 1 } : { "x": 2 }`).(*ast.TernaryExpression)
		s.Require().True(ok)
		is, ok := t.Condition.(*ast.IsExpression)
		s.Require().True(ok)
		s.True(is.Ambiguous())
	})

	whenRules := map[string]string{
		"type keyword nullable before a rule body": `rule r = default false when u is string? { yield true }`,
		"parenthesised nullable shape":             `rule r = default false when (u is User?) { yield true }`,
		"shape without the nullable suffix":        `rule r = default false when u is User { yield true }`,
	}
	for name, rule := range whenRules {
		s.Run(name, func() {
			src := "namespace com/ex\nshape User { name: string }\npolicy p {\n  fact u?: document\n  " + rule + "\n  export decision of r\n}\n"
			_, err := NewParserFromString(src, "when.sentrie").ParseProgram(s.T().Context())
			s.Require().NoError(err)
		})
	}

	s.Run("name nullable before a rule body needs parentheses", func() {
		src := "namespace com/ex\nshape User { name: string }\npolicy p {\n  fact u?: document\n  rule r = default false when u is User? { yield true }\n  export decision of r\n}\n"
		_, err := NewParserFromString(src, "when.sentrie").ParseProgram(s.T().Context())
		s.Error(err)
	})
}

// TestParseIsNotAndCompositeTypes covers `is not` over a name and over
// composite or constrained types, which must stay a single type reference.
func (s *ParserTestSuite) TestParseIsNotAndCompositeTypes() {
	notIs := func(src string) *ast.IsExpression {
		u, ok := s.parseIsExpr(src).(*ast.UnaryExpression)
		s.Require().True(ok, src)
		s.Equal("not", u.Operator)
		is, ok := u.Right.(*ast.IsExpression)
		s.Require().True(ok, src)
		return is
	}
	s.Run("is not a name is ambiguous", func() {
		s.True(notIs("x is not User").Ambiguous())
	})
	s.Run("is not a nullable type", func() {
		s.IsType(&ast.NullableTypeRef{}, notIs("x is not string?").Type)
	})
	s.Run("record keeps all its element types", func() {
		is, ok := s.parseIsExpr("x is record[number, string]").(*ast.IsExpression)
		s.Require().True(ok)
		rec, ok := is.Type.(*ast.RecordTypeRef)
		s.Require().True(ok)
		s.Len(rec.Fields, 2)
	})
	s.Run("list element constraint stays on the element", func() {
		is, ok := s.parseIsExpr("x is list[string@minlength(2)]").(*ast.IsExpression)
		s.Require().True(ok)
		list, ok := is.Type.(*ast.ListTypeRef)
		s.Require().True(ok)
		s.Len(list.ElemType.GetConstraints(), 1)
	})
	s.Run("constraint after a nullable type", func() {
		is, ok := s.parseIsExpr("x is string?@minlength(2)").(*ast.IsExpression)
		s.Require().True(ok)
		s.NotNil(is.Type)
		s.Len(is.Type.GetConstraints(), 1)
	})
}
