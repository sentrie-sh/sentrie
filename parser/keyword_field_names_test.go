// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// TestParseFieldAccessEveryKeywordAsName checks that every reserved word is a
// field name after `.` (#159).
func (s *ParserTestSuite) TestParseFieldAccessEveryKeywordAsName() {
	for _, kw := range allKeywordWords {
		p := NewParserFromString("d."+kw, "test.sentrie")
		expr := p.parseExpression(s.T().Context(), LOWEST)
		s.Require().NoError(p.err, "keyword: %s", kw)
		fa, ok := expr.(*ast.FieldAccessExpression)
		s.Require().True(ok, "keyword: %s", kw)
		s.Equal(kw, fa.Field)
		s.Equal("d", fa.Left.(*ast.Identifier).Value)
	}
}

// TestParseFieldAccessKeywordChain reads `x.true.null` as two accesses.
func (s *ParserTestSuite) TestParseFieldAccessKeywordChain() {
	p := NewParserFromString("x.true.null", "test.sentrie")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(p.err)
	outer, ok := expr.(*ast.FieldAccessExpression)
	s.Require().True(ok)
	s.Equal("null", outer.Field)
	inner, ok := outer.Left.(*ast.FieldAccessExpression)
	s.Require().True(ok)
	s.Equal("true", inner.Field)
}

// TestParseKeywordMethodCall reads `alias.contains(...)` as a call whose callee
// is a field access.
func (s *ParserTestSuite) TestParseKeywordMethodCall() {
	p := NewParserFromString("alias.contains(1)", "test.sentrie")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(p.err)
	call, ok := expr.(*ast.CallExpression)
	s.Require().True(ok)
	callee, ok := call.Callee.(*ast.FieldAccessExpression)
	s.Require().True(ok)
	s.Equal("contains", callee.Field)
	s.Len(call.Arguments, 1)
}

// TestParseKeywordFieldAmbiguities checks that an operator keyword right after
// a keyword field name is still the operator.
func (s *ParserTestSuite) TestParseKeywordFieldAmbiguities() {
	s.Run("x.is is defined", func() {
		p := NewParserFromString("x.is is defined", "test.sentrie")
		expr := p.parseExpression(s.T().Context(), LOWEST)
		s.Require().NoError(p.err)
		def, ok := expr.(*ast.IsDefinedExpression)
		s.Require().True(ok)
		fa, ok := def.Left.(*ast.FieldAccessExpression)
		s.Require().True(ok)
		s.Equal("is", fa.Field)
	})
	s.Run("2 in [x.in]", func() {
		p := NewParserFromString("2 in [x.in]", "test.sentrie")
		expr := p.parseExpression(s.T().Context(), LOWEST)
		s.Require().NoError(p.err)
		in, ok := expr.(*ast.InfixExpression)
		s.Require().True(ok)
		s.Equal("in", in.Operator)
		list, ok := in.Right.(*ast.ListLiteral)
		s.Require().True(ok)
		s.Require().Len(list.Values, 1)
		s.Equal("in", list.Values[0].(*ast.FieldAccessExpression).Field)
	})
	s.Run("x.contains contains 7", func() {
		p := NewParserFromString("x.contains contains 7", "test.sentrie")
		expr := p.parseExpression(s.T().Context(), LOWEST)
		s.Require().NoError(p.err)
		in, ok := expr.(*ast.InfixExpression)
		s.Require().True(ok)
		s.Equal("contains", in.Operator)
		s.Equal("contains", in.Left.(*ast.FieldAccessExpression).Field)
	})
}

// TestParseFieldAccessNonNameStillRejected keeps the error for a non-name.
func (s *ParserTestSuite) TestParseFieldAccessNonNameStillRejected() {
	p := NewParserFromString("x.1", "test.sentrie")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), "expected Ident, got Int")
}

// TestParseShapeEveryKeywordAsFieldName checks required and optional shape
// fields named with any reserved word.
func (s *ParserTestSuite) TestParseShapeEveryKeywordAsFieldName() {
	for _, kw := range allKeywordWords {
		p := NewParserFromString("shape S { "+kw+": string "+kw+"2?: number }", "test.sentrie")
		stmt := parseShapeStatement(s.T().Context(), p)
		s.Require().NoError(p.err, "keyword: %s", kw)
		shape := stmt.(*ast.ShapeStatement)
		s.Require().Contains(shape.Complex.Fields, kw)
		s.False(shape.Complex.Fields[kw].Optional)
	}
	p := NewParserFromString("shape S { version: string default?: number }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), p)
	s.Require().NoError(p.err)
	fields := stmt.(*ast.ShapeStatement).Complex.Fields
	s.False(fields["version"].Optional)
	s.True(fields["default"].Optional)
}

// TestParseAttachKeywordNames checks keyword attachment names, including `as`.
func (s *ParserTestSuite) TestParseAttachKeywordNames() {
	for _, kw := range allKeywordWords {
		p := NewParserFromString("export decision of r attach "+kw+" as 5", "test.sentrie")
		stmt := parseRuleExportStatement(s.T().Context(), p)
		s.Require().NoError(p.err, "keyword: %s", kw)
		exp := stmt.(*ast.RuleExportStatement)
		s.Require().Len(exp.Attachments, 1)
		s.Equal(kw, exp.Attachments[0].What)
	}
	p := NewParserFromString("export decision of r attach version as s.version attach default as 5", "test.sentrie")
	stmt := parseRuleExportStatement(s.T().Context(), p)
	s.Require().NoError(p.err)
	exp := stmt.(*ast.RuleExportStatement)
	s.Require().Len(exp.Attachments, 2)
	s.Equal("version", exp.Attachments[0].What)
	s.Equal("default", exp.Attachments[1].What)
}

// TestParseAttachAsNameThenMissingAs reads `attach as 1` as name `as`.
func (s *ParserTestSuite) TestParseAttachAsNameThenMissingAs() {
	p := NewParserFromString("export decision of r attach as 1", "test.sentrie")
	stmt := parseRuleExportStatement(s.T().Context(), p)
	s.Nil(stmt)
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), "expected 'as', got Int")
}

// TestParseBareDictKeywordKeyStillRejected keeps bare dict keys identifier-only.
func (s *ParserTestSuite) TestParseBareDictKeywordKeyStillRejected() {
	p := NewParserFromString("{title: 1}", "test.sentrie")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.Require().Error(p.err)

	p = NewParserFromString(`{"title": 1}`, "test.sentrie")
	expr = p.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(p.err)
	s.NotNil(expr)
}
