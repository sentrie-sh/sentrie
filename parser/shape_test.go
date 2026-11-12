// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) TestParseShapeFieldNullabilityMatrix() {
	parser := NewParserFromString("shape Person { name:string age?:number middle_name:string? nickname?:string? }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	s.Require().NotNil(shapeStmt.Complex)

	s.False(shapeStmt.Complex.Fields["name"].Optional)
	s.False(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["name"].Type))

	s.True(shapeStmt.Complex.Fields["age"].Optional)
	s.False(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["age"].Type))

	s.False(shapeStmt.Complex.Fields["middle_name"].Optional)
	s.True(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["middle_name"].Type))

	s.True(shapeStmt.Complex.Fields["nickname"].Optional)
	s.True(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["nickname"].Type))
}

func (s *ParserTestSuite) TestParseShapeFieldRejectsLegacyBangSyntax() {
	testCases := []struct {
		input   string
		message string
	}{
		{"shape Person { name!: string }", "name!: T"},
		{"shape Person { phone!?: string }", "phone!?: T"},
		{"shape Person { phone?!: string }", "phone?!: T"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentrie")
		stmt := parseShapeStatement(s.T().Context(), parser)
		s.Require().Nil(stmt)
		s.Require().Error(parser.err)
		s.Contains(parser.err.Error(), tc.message)
	}
}

func (s *ParserTestSuite) TestParseNullableTypeRefWithConstraint() {
	parser := NewParserFromString("shape Person { middle_name: string? @minlength(1) }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	typ := shapeStmt.Complex.Fields["middle_name"].Type
	s.True(ast.IsNullableTypeRef(typ))
	s.Len(ast.UnwrapNullableTypeRef(typ).GetConstraints(), 1)
}

func (s *ParserTestSuite) TestParseTypeRefCollectionKindsAndErrors() {
	parser := NewParserFromString("shape T { names:list[string]? metadata:dict[number] tuple:record[string, number] }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)

	names := shapeStmt.Complex.Fields["names"].Type
	s.True(ast.IsNullableTypeRef(names))
	_, ok = ast.UnwrapNullableTypeRef(names).(*ast.ListTypeRef)
	s.True(ok)

	_, ok = shapeStmt.Complex.Fields["metadata"].Type.(*ast.DictTypeRef)
	s.True(ok)

	recordRef, ok := shapeStmt.Complex.Fields["tuple"].Type.(*ast.RecordTypeRef)
	s.True(ok)
	s.Len(recordRef.Fields, 2)

	badParser := NewParserFromString("shape T { names:list[string }", "test.sentrie")
	badStmt := parseShapeStatement(s.T().Context(), badParser)
	s.Require().Nil(badStmt)
	s.Require().Error(badParser.err)
}

func (s *ParserTestSuite) TestParseComplexShapeWithClauseAndInvalidWith() {
	parser := NewParserFromString("shape Child with app/Base { id:string -- trailing\n name:string }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	s.Require().NotNil(shapeStmt.Complex)
	s.Require().NotNil(shapeStmt.Complex.With)
	s.Equal("app/Base", shapeStmt.Complex.With.String())
	s.Len(shapeStmt.Complex.Fields, 2)

	invalid := NewParserFromString("shape Child with { name:string }", "test.sentrie")
	invalidStmt := parseShapeStatement(s.T().Context(), invalid)
	s.Nil(invalidStmt)
	s.Error(invalid.err)
}

func (s *ParserTestSuite) TestParseTypeRefRejectsInvalidStartToken() {
	parser := NewParserFromString("shape Person { name: ? }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
	s.Contains(parser.err.Error(), "expected one of")
}

func (s *ParserTestSuite) TestParseTypeRefAdditionalCollectionBranches() {
	dictMissingBracket := NewParserFromString("shape T { meta:dict[string }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), dictMissingBracket)
	s.Nil(stmt)
	s.Error(dictMissingBracket.err)

	recordTrailingComma := NewParserFromString("shape T { tuple:record[string, number,] }", "test.sentrie")
	stmt = parseShapeStatement(s.T().Context(), recordTrailingComma)
	s.Require().NotNil(stmt)
	s.Require().NoError(recordTrailingComma.err)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	recordRef, ok := shapeStmt.Complex.Fields["tuple"].Type.(*ast.RecordTypeRef)
	s.Require().True(ok)
	s.Len(recordRef.Fields, 2)
}

func (s *ParserTestSuite) TestParseTypeRefConstraintAcceptsExponentLiterals() {
	parser := NewParserFromString("shape T { amount:number @min(1e-3) @max(2.5E+6) @not_in([1e2, 5]) }", "test.sentra")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	constraints := shapeStmt.Complex.Fields["amount"].Type.GetConstraints()
	s.Require().Len(constraints, 3)

	minArg, ok := constraints[0].Args[0].(*ast.FloatLiteral)
	s.Require().True(ok, "min arg parsed as %T", constraints[0].Args[0])
	s.Equal(1e-3, minArg.Value)
	maxArg, ok := constraints[1].Args[0].(*ast.FloatLiteral)
	s.Require().True(ok, "max arg parsed as %T", constraints[1].Args[0])
	s.Equal(2.5e6, maxArg.Value)
	list, ok := constraints[2].Args[0].(*ast.ListLiteral)
	s.Require().True(ok, "in arg parsed as %T", constraints[2].Args[0])
	first, ok := list.Values[0].(*ast.FloatLiteral)
	s.Require().True(ok, "in element parsed as %T", list.Values[0])
	s.Equal(1e2, first.Value)
}

func (s *ParserTestSuite) TestParseTypeRefConstraintRejectsMalformedExponent() {
	parser := NewParserFromString("shape T { amount:number @min(1e) }", "test.sentra")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Nil(stmt)
	s.Require().Error(parser.err)
	s.Contains(parser.err.Error(), "invalid number literal")
}

func (s *ParserTestSuite) TestParseTypeRefConstraintValidationFailureBubblesAsParserError() {
	parser := NewParserFromString("shape T { name:string @minlength() }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
	s.Contains(parser.err.Error(), "cannot add constraint minlength")
}

func (s *ParserTestSuite) TestParseTypeRefMalformedRecordDoesNotLoop() {
	parser := NewParserFromString("shape T { tuple:record[string, number }", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
}

func (s *ParserTestSuite) TestParseShapeOpenMarker() {
	parse := func(src string) (*ast.ShapeStatement, error) {
		p := NewParserFromString(src, "test.sentrie")
		stmt := parseShapeStatement(s.T().Context(), p)
		if p.err != nil {
			return nil, p.err
		}
		shapeStmt, ok := stmt.(*ast.ShapeStatement)
		s.Require().True(ok)
		return shapeStmt, nil
	}

	for name, src := range map[string]string{
		"after fields":          "shape S { a: string b: number ... }",
		"alone":                 "shape S { ... }",
		"comment after marker":  "shape S { a: string ... -- trailing\n }",
		"comment line after":    "shape S {\n a: string\n ...\n -- note\n}",
		"comment before marker": "shape S {\n a: string\n -- note\n ...\n}",
	} {
		s.Run(name, func() {
			stmt, err := parse(src)
			s.Require().NoError(err)
			s.True(stmt.Complex.Open)
			s.NotZero(stmt.Complex.OpenRange.To.Offset)
		})
	}

	stmt, err := parse("shape S { a: string }")
	s.Require().NoError(err)
	s.False(stmt.Complex.Open)

}

func (s *ParserTestSuite) TestParseShapeOpenMarkerMustBeLast() {
	for name, tc := range map[string]struct{ src, msg string }{
		"field after marker":                {"shape S { a: string ... b: number }", "must be the last entry"},
		"second marker":                     {"shape S { a: string ... ... }", "must be the last entry"},
		"marker then field on its own line": {"shape S {\n ...\n b: number\n}", "must be the last entry"},
	} {
		s.Run(name, func() {
			p := NewParserFromString(tc.src, "test.sentrie")
			stmt := parseShapeStatement(s.T().Context(), p)
			s.Nil(stmt)
			s.Require().Error(p.err)
			s.Contains(p.err.Error(), tc.msg)
		})
	}

	s.Run("missing close brace after marker", func() {
		p := NewParserFromString("shape S { a: string ...", "test.sentrie")
		stmt := parseShapeStatement(s.T().Context(), p)
		s.Nil(stmt)
		s.Require().Error(p.err)
		s.NotContains(p.err.Error(), "must be the last entry")
	})
}

func (s *ParserTestSuite) TestParseShapeDuplicateFieldNamesBothSpans() {
	p := NewParserFromString("shape S {\n  a: string\n  b: number\n  a: number\n}", "test.sentrie")
	stmt := parseShapeStatement(s.T().Context(), p)
	s.Nil(stmt)
	s.Require().Error(p.err)
	msg := p.err.Error()
	s.Contains(msg, "duplicate shape field 'a'")
	s.Contains(msg, "first declared at")
	s.Contains(msg, "test.sentrie:4:", "error is reported at the second declaration")
	s.Contains(msg, "test.sentrie:2:", "error names the first declaration")

	// a keyword field name still counts as a duplicate (#159 stack)
	p = NewParserFromString("shape S { default: string default: number }", "test.sentrie")
	s.Nil(parseShapeStatement(s.T().Context(), p))
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), "duplicate shape field 'default'")
}
