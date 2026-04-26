// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

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
