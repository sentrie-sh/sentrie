// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) TestParseShapeFieldNullabilityMatrix() {
	parser := NewParserFromString("shape Person { name:string age?:number middle_name:string? nickname?:string? }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
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
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseShapeStatement(context.Background(), parser)
		s.Require().Nil(stmt)
		s.Require().Error(parser.err)
		s.Contains(parser.err.Error(), tc.message)
	}
}

func (s *ParserTestSuite) TestParseFactNullableTypeRef() {
	parser := NewParserFromString("fact input?: string?", "test.sentra")
	stmt := parseFactStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	factStmt, ok := stmt.(*ast.FactStatement)
	s.Require().True(ok)
	s.True(factStmt.Optional)
	s.True(ast.IsNullableTypeRef(factStmt.Type))
}

func (s *ParserTestSuite) TestParseNullableTypeRefWithConstraint() {
	parser := NewParserFromString("shape Person { middle_name: string? @minlength(1) }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	typ := shapeStmt.Complex.Fields["middle_name"].Type
	s.True(ast.IsNullableTypeRef(typ))
	s.Len(ast.UnwrapNullableTypeRef(typ).GetConstraints(), 1)
}

func (s *ParserTestSuite) TestParseTypeRefCollectionKindsAndErrors() {
	parser := NewParserFromString("shape T { names:list[string]? metadata:dict[number] tuple:record[string, number] }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
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

	badParser := NewParserFromString("shape T { names:list[string }", "test.sentra")
	badStmt := parseShapeStatement(context.Background(), badParser)
	s.Require().Nil(badStmt)
	s.Require().Error(badParser.err)
}

func (s *ParserTestSuite) TestParseComplexShapeWithClauseAndInvalidWith() {
	parser := NewParserFromString("shape Child with app/Base { id:string -- trailing\n name:string }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	s.Require().NotNil(shapeStmt.Complex)
	s.Require().NotNil(shapeStmt.Complex.With)
	s.Equal("app/Base", shapeStmt.Complex.With.String())
	s.Len(shapeStmt.Complex.Fields, 2)

	invalid := NewParserFromString("shape Child with { name:string }", "test.sentra")
	invalidStmt := parseShapeStatement(context.Background(), invalid)
	s.Nil(invalidStmt)
	s.Error(invalid.err)
}

func (s *ParserTestSuite) TestParseTypeRefRejectsInvalidStartToken() {
	parser := NewParserFromString("shape Person { name: ? }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
	s.Contains(parser.err.Error(), "expected one of")
}

func (s *ParserTestSuite) TestParseTypeRefConstraintErrorBranches() {
	noAt := NewParserFromString("minlength(1)", "test.sentra")
	constraint := parseTypeRefConstraint(context.Background(), noAt, nil)
	s.Nil(constraint)
	s.Error(noAt.err)

	missingName := NewParserFromString("@(1)", "test.sentra")
	constraint = parseTypeRefConstraint(context.Background(), missingName, nil)
	s.Nil(constraint)
	s.Error(missingName.err)

	missingArgs := NewParserFromString("@minlength()", "test.sentra")
	constraint = parseTypeRefConstraint(context.Background(), missingArgs, nil)
	s.NotNil(constraint)
	s.Equal("minlength", constraint.Name)
	s.Len(constraint.Args, 0)
}

func (s *ParserTestSuite) TestParseTypeRefConstraintAdditionalErrorBranches() {
	missingParen := NewParserFromString("@minlength", "test.sentra")
	constraint := parseTypeRefConstraint(context.Background(), missingParen, nil)
	s.Nil(constraint)
	s.Error(missingParen.err)

	nonLiteralArg := NewParserFromString("@minlength(user.name)", "test.sentra")
	constraint = parseTypeRefConstraint(context.Background(), nonLiteralArg, nil)
	s.Nil(constraint)
	s.Error(nonLiteralArg.err)
	s.Contains(nonLiteralArg.err.Error(), "constraint arguments must be literals")
}

func (s *ParserTestSuite) TestParseTypeRefAdditionalCollectionBranches() {
	dictMissingBracket := NewParserFromString("shape T { meta:dict[string }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), dictMissingBracket)
	s.Nil(stmt)
	s.Error(dictMissingBracket.err)

	recordTrailingComma := NewParserFromString("shape T { tuple:record[string, number,] }", "test.sentra")
	stmt = parseShapeStatement(context.Background(), recordTrailingComma)
	s.Require().NotNil(stmt)
	s.Require().NoError(recordTrailingComma.err)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	recordRef, ok := shapeStmt.Complex.Fields["tuple"].Type.(*ast.RecordTypeRef)
	s.Require().True(ok)
	s.Len(recordRef.Fields, 2)
}

func (s *ParserTestSuite) TestParseTypeRefConstraintValidationFailureBubblesAsParserError() {
	parser := NewParserFromString("shape T { name:string @minlength() }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
	s.Contains(parser.err.Error(), "cannot add constraint minlength")
}

func (s *ParserTestSuite) TestParseTypeRefDirectKindCoverage() {
	cases := []struct {
		input    string
		assertFn func(ref ast.TypeRef)
	}{
		{
			input: "boolean",
			assertFn: func(ref ast.TypeRef) {
				_, ok := ref.(*ast.TrinaryTypeRef)
				s.True(ok)
			},
		},
		{
			input: "trinary",
			assertFn: func(ref ast.TypeRef) {
				_, ok := ref.(*ast.TrinaryTypeRef)
				s.True(ok)
			},
		},
		{
			input: "document",
			assertFn: func(ref ast.TypeRef) {
				_, ok := ref.(*ast.DocumentTypeRef)
				s.True(ok)
			},
		},
		{
			input: "app/User",
			assertFn: func(ref ast.TypeRef) {
				shapeRef, ok := ref.(*ast.ShapeTypeRef)
				s.True(ok)
				s.Equal("app/User", shapeRef.Ref.String())
			},
		},
	}

	for _, tc := range cases {
		p := NewParserFromString(tc.input, "test.sentra")
		ref := parseTypeRef(context.Background(), p)
		s.Require().NoError(p.err)
		s.Require().NotNil(ref)
		tc.assertFn(ref)
	}
}

func (s *ParserTestSuite) TestParseTypeRefDirectCollectionErrorBranches() {
	listOnly := NewParserFromString("list", "test.sentra")
	listRef := parseTypeRef(context.Background(), listOnly)
	s.Nil(listRef)
	s.Error(listOnly.err)

	dictOnly := NewParserFromString("dict", "test.sentra")
	dictRef := parseTypeRef(context.Background(), dictOnly)
	s.Nil(dictRef)
	s.Error(dictOnly.err)

	recordOnly := NewParserFromString("record", "test.sentra")
	recordRef := parseTypeRef(context.Background(), recordOnly)
	s.Nil(recordRef)
	s.Error(recordOnly.err)
}

func (s *ParserTestSuite) TestParseTypeRefMalformedRecordDoesNotLoop() {
	parser := NewParserFromString("shape T { tuple:record[string, number }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Nil(stmt)
	s.Error(parser.err)
}
