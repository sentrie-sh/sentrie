// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) TestParseTypeRefConstraintErrorBranches() {
	noAt := NewParserFromString("minlength(1)", "test.sentra")
	constraint := parseTypeRefConstraint(s.T().Context(), noAt, nil)
	s.Nil(constraint)
	s.Error(noAt.err)

	missingName := NewParserFromString("@(1)", "test.sentra")
	constraint = parseTypeRefConstraint(s.T().Context(), missingName, nil)
	s.Nil(constraint)
	s.Error(missingName.err)

	missingArgs := NewParserFromString("@minlength()", "test.sentra")
	constraint = parseTypeRefConstraint(s.T().Context(), missingArgs, nil)
	s.NotNil(constraint)
	s.Equal("minlength", constraint.Name)
	s.Len(constraint.Args, 0)
}

func (s *ParserTestSuite) TestParseTypeRefConstraintAdditionalErrorBranches() {
	missingParen := NewParserFromString("@minlength", "test.sentra")
	constraint := parseTypeRefConstraint(s.T().Context(), missingParen, nil)
	s.Nil(constraint)
	s.Error(missingParen.err)

	nonLiteralArg := NewParserFromString("@minlength(user.name)", "test.sentra")
	constraint = parseTypeRefConstraint(s.T().Context(), nonLiteralArg, nil)
	s.Nil(constraint)
	s.Error(nonLiteralArg.err)
	s.Contains(nonLiteralArg.err.Error(), "constraint arguments must be literals")
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
		ref := parseTypeRef(s.T().Context(), p)
		s.Require().NoError(p.err)
		s.Require().NotNil(ref)
		tc.assertFn(ref)
	}
}

func (s *ParserTestSuite) TestParseTypeRefDirectCollectionErrorBranches() {
	listOnly := NewParserFromString("list", "test.sentra")
	listRef := parseTypeRef(s.T().Context(), listOnly)
	s.Nil(listRef)
	s.Error(listOnly.err)

	dictOnly := NewParserFromString("dict", "test.sentra")
	dictRef := parseTypeRef(s.T().Context(), dictOnly)
	s.Nil(dictRef)
	s.Error(dictOnly.err)

	recordOnly := NewParserFromString("record", "test.sentra")
	recordRef := parseTypeRef(s.T().Context(), recordOnly)
	s.Nil(recordRef)
	s.Error(recordOnly.err)
}

// TestParseSetTypeRef covers `set[T]` (#162): `set` is a type only before `[`,
// so a bare `set` stays a shape name and `set(...)` stays a call.
func (s *ParserTestSuite) TestParseSetTypeRef() {
	p := NewParserFromString("set[string]?@not_empty()", "test.sentrie")
	ref := parseTypeRef(s.T().Context(), p)
	s.Require().NoError(p.err)
	s.Require().NotNil(ref)
	nullable, ok := ref.(*ast.NullableTypeRef)
	s.Require().True(ok)
	set, ok := nullable.Inner.(*ast.SetTypeRef)
	s.Require().True(ok)
	s.Equal("set[string]", set.String())
	s.IsType(&ast.StringTypeRef{}, set.ElemType)
	s.Len(nullable.GetConstraints(), 1)

	nested := NewParserFromString("set[list[number]]", "test.sentrie")
	ref = parseTypeRef(s.T().Context(), nested)
	s.Require().NoError(nested.err)
	s.Equal("set[list[number]]", ref.String())

	bare := NewParserFromString("set", "test.sentrie")
	ref = parseTypeRef(s.T().Context(), bare)
	s.Require().NoError(bare.err)
	s.IsType(&ast.ShapeTypeRef{}, ref)

	unclosed := NewParserFromString("set[string", "test.sentrie")
	s.Nil(parseTypeRef(s.T().Context(), unclosed))
	s.Error(unclosed.err)

	badConstraint := NewParserFromString("set[string]@minlength(1)", "test.sentrie")
	s.Nil(parseTypeRef(s.T().Context(), badConstraint))
	s.Error(badConstraint.err)

	is, ok := s.parseIsExpr("x is set[number]").(*ast.IsExpression)
	s.Require().True(ok)
	s.Nil(is.Right)
	s.IsType(&ast.SetTypeRef{}, is.Type)

	call := NewParserFromString(`set(["a"])`, "test.sentrie")
	expr := call.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(call.err)
	s.IsType(&ast.CallExpression{}, expr)
}
