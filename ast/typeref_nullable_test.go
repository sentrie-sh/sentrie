// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *AstTestSuite) TestNullableTypeRefHelpersAndConstraintPropagation() {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 6, Offset: 5},
	}
	inner := NewStringTypeRef(r)
	nullable := NewNullableTypeRef(inner, r)

	s.True(IsNullableTypeRef(nullable))
	s.False(IsNullableTypeRef(inner))
	s.Equal(inner, UnwrapNullableTypeRef(nullable))
	s.Equal(inner, UnwrapNullableTypeRef(inner))
	s.Equal("string?", nullable.String())
}

func (s *AstTestSuite) TestNullableTypeRefAddConstraintUpdatesOuterSpan() {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 6, Offset: 5},
	}
	constraintRange := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 8, Offset: 7},
		To:   tokens.Pos{Line: 1, Column: 20, Offset: 19},
	}
	nullable := NewNullableTypeRef(NewStringTypeRef(r), r)
	constraint := NewTypeRefConstraint("maxlength", []Expression{NewIntegerLiteral(32, constraintRange)}, constraintRange)

	err := nullable.AddConstraint(constraint)
	s.Require().NoError(err)
	s.Require().Len(nullable.GetConstraints(), 1)
	s.Equal(constraintRange.To, nullable.Span().To)
}

func (s *AstTestSuite) TestNullableTypeRefAddConstraintPropagatesInnerError() {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 6, Offset: 5},
	}
	constraint := NewTypeRefConstraint("unknown", []Expression{NewIntegerLiteral(1, r)}, r)
	nullable := NewNullableTypeRef(NewStringTypeRef(r), r)

	err := nullable.AddConstraint(constraint)
	s.Require().Error(err)
	s.Require().Len(nullable.GetConstraints(), 0)
}
