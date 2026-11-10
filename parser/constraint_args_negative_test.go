// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) constraintArgs(input string) []ast.Expression {
	p := NewParserFromString(input, "test.sentrie")
	ref := parseTypeRef(s.T().Context(), p)
	s.Require().NoError(p.err, input)
	s.Require().NotNil(ref, input)
	cs := ref.GetConstraints()
	s.Require().NotEmpty(cs, input)
	return cs[0].Args
}

// TestConstraintNegativeNumberArguments covers `@min(-5)` and friends (#146).
func (s *ParserTestSuite) TestConstraintNegativeNumberArguments() {
	s.Run("integers and floats", func() {
		args := s.constraintArgs("number @min(-5)")
		i, ok := args[0].(*ast.IntegerLiteral)
		s.Require().True(ok, "%T", args[0])
		s.Equal(int64(-5), i.Value)

		args = s.constraintArgs("number @gt(-0.5)")
		f, ok := args[0].(*ast.FloatLiteral)
		s.Require().True(ok, "%T", args[0])
		s.Equal(-0.5, f.Value)
	})
	s.Run("range", func() {
		args := s.constraintArgs("number @range(-10, -1)")
		s.Require().Len(args, 2)
		a, ok := args[0].(*ast.IntegerLiteral)
		s.Require().True(ok)
		b, ok := args[1].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(int64(-10), a.Value)
		s.Equal(int64(-1), b.Value)
	})
	s.Run("in a list", func() {
		args := s.constraintArgs("number @not_in([-1, 2, -3.5])")
		list, ok := args[0].(*ast.ListLiteral)
		s.Require().True(ok, "%T", args[0])
		s.Require().Len(list.Values, 3)
		v0, ok := list.Values[0].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(int64(-1), v0.Value)
		v2, ok := list.Values[2].(*ast.FloatLiteral)
		s.Require().True(ok)
		s.Equal(-3.5, v2.Value)
	})
	s.Run("span covers minus and number", func() {
		input := "number @min(-5)"
		args := s.constraintArgs(input)
		sp := args[0].Span()
		s.Equal("-5", input[sp.From.Offset:sp.To.Offset])
	})
	s.Run("positive literals unchanged", func() {
		args := s.constraintArgs("number @min(5)")
		i, ok := args[0].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(int64(5), i.Value)
	})
}

// TestConstraintMinusMustPrecedeNumberLiteral checks `-x`, `-"a"` and `-(1)`
// stay parse errors, with a position and no garbled formatting (#146).
func (s *ParserTestSuite) TestConstraintMinusMustPrecedeNumberLiteral() {
	for _, input := range []string{`number @min(-x)`, `number @min(-"a")`, `number @min(-(1))`, `number @min(-true)`} {
		p := NewParserFromString(input, "test.sentrie")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		msg := p.err.Error()
		s.Contains(msg, "followed by a number literal", input)
		s.Contains(msg, "test.sentrie:1:", input)
		s.NotContains(msg, "%!", input)
	}
}

// TestConstraintArgumentErrorShowsPosition checks the non-literal error prints
// file:line:col, not a Go formatting verb error (#146).
func (s *ParserTestSuite) TestConstraintArgumentErrorShowsPosition() {
	cases := []string{"number @min(foo)", "number @min(x + 1)", "number @eq({foo: 1})"}
	for _, input := range cases {
		p := NewParserFromString(input, "test.sentrie")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		msg := p.err.Error()
		s.NotContains(msg, "%!", input)
		s.NotContains(msg, "{%", input)
		s.Contains(msg, "test.sentrie:1:", input)
	}
}

// TestConstraintNameAnyKeyword checks every keyword after `@` is read as a
// name; an unknown one is reported as an unknown constraint, never as a token
// error (#146).
func (s *ParserTestSuite) TestConstraintNameAnyKeyword() {
	for _, kw := range []string{"policy", "rule", "let", "if", "namespace", "shape", "true", "null", "is", "not", "and", "or", "export", "fact", "use", "from", "as", "yield", "default"} {
		input := "number @" + kw + "(1)"
		p := NewParserFromString(input, "test.sentrie")
		ref := parseTypeRef(s.T().Context(), p)
		s.Nil(ref, input)
		s.Require().Error(p.err, input)
		s.Contains(p.err.Error(), "cannot add constraint "+kw, input)
		s.Contains(p.err.Error(), "not found", input)
		s.NotContains(p.err.Error(), "expected Ident", input)
	}
}
