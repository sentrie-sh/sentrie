// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// TestParseNamespaceKeywordSegments checks that a keyword is accepted as a
// later segment of a namespace path (#297), that the span covers the whole
// path, and that the printed name parses back to the same parts.
func (s *ParserTestSuite) TestParseNamespaceKeywordSegments() {
	testCases := []struct {
		input string
		parts []string
	}{
		{"namespace constraints/in", []string{"constraints", "in"}},
		{"namespace text/string", []string{"text", "string"}},
		{"namespace a/policy/true/rule", []string{"a", "policy", "true", "rule"}},
		{"namespace com/not/null", []string{"com", "not", "null"}},
	}

	for _, tc := range testCases {
		p := NewParserFromString(tc.input, "test.sentrie")
		stmt := parseNamespaceStatement(s.T().Context(), p)
		s.Require().NoError(p.err, "input: %s", tc.input)
		ns, ok := stmt.(*ast.NamespaceStatement)
		s.Require().True(ok, "input: %s", tc.input)
		s.Equal(tc.parts, ns.Name.Parts, "input: %s", tc.input)
		s.Equal(len(tc.input)+1, ns.Name.Rnge.To.Column, "input: %s", tc.input) // columns are 1-based, To is one past the end

		// The printed name is valid source again.
		again := NewParserFromString("namespace "+ns.Name.String(), "test.sentrie")
		reparsed, ok := parseNamespaceStatement(s.T().Context(), again).(*ast.NamespaceStatement)
		s.Require().NoError(again.err, "input: %s", tc.input)
		s.Require().True(ok, "input: %s", tc.input)
		s.Equal(tc.parts, reparsed.Name.Parts, "input: %s", tc.input)
	}
}

// TestParseFQNFirstSegmentKeywordRejected checks that the first segment of a
// name stays an identifier, with an error that names the keyword. A name in a
// type position must still be told apart from a type keyword (`string`).
func (s *ParserTestSuite) TestParseFQNFirstSegmentKeywordRejected() {
	testCases := []string{
		"namespace in/x",
		"namespace string",
		"namespace policy/a",
	}

	for _, tc := range testCases {
		p := NewParserFromString(tc, "test.sentrie")
		stmt := parseNamespaceStatement(s.T().Context(), p)
		s.Nil(stmt, "input: %s", tc)
		s.Require().Error(p.err, "input: %s", tc)
		s.Contains(p.err.Error(), "the first segment of a name cannot be a keyword", "input: %s", tc)
	}

	p := NewParserFromString("import decision ok from in/p", "test.sentrie")
	s.Nil(parseImportExpression(s.T().Context(), p))
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), "parsing error at test.sentrie:1:25-27: expected Ident, got keyword 'in'")
}

// TestParseFQNKeywordSegmentsInEveryDeclarationPosition checks that every
// production that reads a name through parseFQN accepts keyword segments the
// same way: an imported decision's source, a shape's base, and a named type in
// a fact, a let and a cast.
func (s *ParserTestSuite) TestParseFQNKeywordSegmentsInEveryDeclarationPosition() {
	ctx := s.T().Context()

	s.Run("import decision from", func() {
		p := NewParserFromString("import decision ok from constraints/in/p", "test.sentrie")
		imp, ok := parseImportExpression(ctx, p).(*ast.ImportClause)
		s.Require().NoError(p.err)
		s.Require().True(ok)
		s.Equal([]string{"constraints", "in", "p"}, imp.FromPolicyFQN.Parts)
	})

	s.Run("shape with base", func() {
		p := NewParserFromString("shape T with text/string/Item { label: string }", "test.sentrie")
		shape, ok := parseShapeStatement(ctx, p).(*ast.ShapeStatement)
		s.Require().NoError(p.err)
		s.Require().True(ok)
		s.Require().NotNil(shape.Complex)
		s.Equal([]string{"text", "string", "Item"}, shape.Complex.With.Parts)
	})

	s.Run("fact type", func() {
		p := NewParserFromString("fact i: text/string/Item", "test.sentrie")
		stmt := parseFactStatement(ctx, p)
		s.Require().NoError(p.err)
		s.Require().NotNil(stmt)
	})

	s.Run("let type", func() {
		p := NewParserFromString("let x: constraints/in/Item = {}", "test.sentrie")
		decl, ok := parseLetsStatement(ctx, p).(*ast.VarDeclaration)
		s.Require().NoError(p.err)
		s.Require().True(ok)
		s.Equal("constraints/in/Item", decl.Type.String())
	})

	s.Run("cast type", func() {
		p := NewParserFromString("x as constraints/in/Item", "test.sentrie")
		cast, ok := p.parseExpression(ctx, LOWEST).(*ast.CastExpression)
		s.Require().NoError(p.err)
		s.Require().True(ok)
		ref, ok := cast.TargetType.(*ast.ShapeTypeRef)
		s.Require().True(ok)
		s.Equal([]string{"constraints", "in", "Item"}, ref.Ref.Parts)
	})
}

// TestParseSlashChainInExpressionKeywordSegments checks slash-qualified names
// in expression position (#297). A keyword that cannot start an expression is a
// name segment after `/`, so a derive callee or the right of `is` can name a
// keyword-segment namespace. The six keywords that do start an expression keep
// that meaning after `/`, so these stay division and are not names.
func (s *ParserTestSuite) TestParseSlashChainInExpressionKeywordSegments() {
	ctx := s.T().Context()
	parse := func(src string) ast.Expression {
		p := NewParserFromString(src, "test.sentrie")
		expr := p.parseExpression(ctx, LOWEST)
		s.Require().NoError(p.err, src)
		s.Require().NotNil(expr, src)
		return expr
	}

	for src, fqn := range map[string]string{
		"lib/in/double(2)":      "lib/in/double",
		"text/string/f(x)":      "text/string/f",
		"a/matches/policy/f(1)": "a/matches/policy/f",
		"com/ex/f(x)":           "com/ex/f",
	} {
		s.Run(src, func() {
			call, ok := parse(src).(*ast.CallExpression)
			s.Require().True(ok)
			s.Equal(fqn, ast.SlashCalleeFQNS(call.Callee))
		})
	}

	s.Run("is with a keyword-segment shape", func() {
		is, ok := parse("x is lib/in/Item").(*ast.IsExpression)
		s.Require().True(ok)
		s.Equal("lib/in/Item", is.Type.String())
	})

	s.Run("a/not b divides by not b", func() {
		infix, ok := parse("a/not b").(*ast.InfixExpression)
		s.Require().True(ok)
		s.Equal("/", infix.Operator)
		s.IsType(&ast.UnaryExpression{}, infix.Right)
	})

	s.Run("a/true divides by a literal", func() {
		infix, ok := parse("a/true").(*ast.InfixExpression)
		s.Require().True(ok)
		s.IsType(&ast.TrinaryLiteral{}, infix.Right)
	})

	s.Run("a/null divides by a literal", func() {
		infix, ok := parse("a/null").(*ast.InfixExpression)
		s.Require().True(ok)
		s.IsType(&ast.NullLiteral{}, infix.Right)
	})

	s.Run("a/true/f(1) is not a qualified callee", func() {
		call, ok := parse("a/true/f(1)").(*ast.CallExpression)
		s.Require().True(ok)
		s.Empty(ast.SlashCalleeFQNS(call.Callee))
	})

	s.Run("only after a slash", func() {
		for _, src := range []string{"1 + in", "x * string", "a/not/f(1)"} {
			p := NewParserFromString(src, "test.sentrie")
			s.Nil(p.parseExpression(ctx, LOWEST), src)
			s.Error(p.err, src)
		}
	})
}
