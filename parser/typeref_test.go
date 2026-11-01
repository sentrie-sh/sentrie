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

// TestParseTypeRefKeywordConstraintName checks that a constraint can be named
// after a keyword (#285): `in` lexes as a keyword, and before the fix
// `number @in([1, 2])` failed with "expected Ident, got in".
func (s *ParserTestSuite) TestParseTypeRefKeywordConstraintName() {
	cases := []struct {
		input      string
		name       string
		constraint string // the source text the constraint spans
	}{
		{"number @in([1, 2])", "in", "@in([1, 2])"},
		{"number @not_in([1, 2])", "not_in", "@not_in([1, 2])"},
		{"number @in([1, 2]) @not_in([3])", "in", "@in([1, 2])"},
	}

	for _, tc := range cases {
		p := NewParserFromString(tc.input, "test.sentra")
		ref := parseTypeRef(s.T().Context(), p)
		s.Require().NoError(p.err, tc.input)
		s.Require().NotNil(ref, tc.input)

		constraints := ref.GetConstraints()
		s.Require().NotEmpty(constraints, tc.input)
		first := constraints[0]
		s.Equal(tc.name, first.Name, tc.input)
		s.Len(first.Args, 1, tc.input)
		s.Equal(tc.constraint, tc.input[first.Span().From.Offset:first.Span().To.Offset], tc.input)

		// The type's span runs to the end of its last constraint.
		s.Equal(tc.input, tc.input[ref.Span().From.Offset:ref.Span().To.Offset], tc.input)
	}
}

// TestParseTypeRefUnknownKeywordConstraintName checks that a keyword which is
// not a constraint of the type is reported as an unknown constraint, the same
// way an unknown identifier is, rather than as a token error.
func (s *ParserTestSuite) TestParseTypeRefUnknownKeywordConstraintName() {
	for _, input := range []string{"number @policy(1)", "string @in([\"a\"])", "number @nope(1)"} {
		p := NewParserFromString(input, "test.sentra")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		s.Contains(p.err.Error(), "cannot add constraint", input)
		s.Contains(p.err.Error(), "not found", input)
		s.NotContains(p.err.Error(), "expected Ident", input)
	}
}

// TestParseTypeRefKeywordConstraintNameWithoutArgs checks that a keyword
// constraint name still needs its parenthesised arguments, before `=` or `)`.
func (s *ParserTestSuite) TestParseTypeRefKeywordConstraintNameWithoutArgs() {
	for _, input := range []string{"number @in = 1", "number @in)"} {
		p := NewParserFromString(input, "test.sentra")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		s.Contains(p.err.Error(), "expected 'LeftParen'", input)
	}

	// `@` followed by a token that is neither a name nor a keyword names what
	// was expected.
	for _, input := range []string{"number @ = 1", "number @)"} {
		p := NewParserFromString(input, "test.sentra")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		s.Contains(p.err.Error(), "expected constraint name after '@', got", input)
	}
}
