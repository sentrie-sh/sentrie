// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// allKeywordWords is every reserved word the lexer turns into a keyword token
// (tokens.keywords), grouped by what they stand for: statement words, operators,
// literals, type names and metadata names.
var allKeywordWords = []string{
	// statements and clauses
	"decision", "yield", "transform", "shape", "derive", "of", "attach", "namespace",
	"policy", "let", "rule", "when", "default", "fact", "export", "use", "from", "as",
	"with", "import",
	// operators
	"and", "or", "xor", "not", "in", "is", "matches", "contains", "defined", "empty",
	// metadata
	"title", "description", "version", "tag",
	// literals
	"true", "false", "unknown", "null",
	// types
	"string", "number", "boolean", "trinary", "list", "dict", "record", "document",
}

// TestParseNamespaceEveryKeywordAsLaterSegment checks that every reserved word,
// not just the ones named in the issue, is accepted as the second and as a
// middle segment of a namespace path (#297).
func (s *ParserTestSuite) TestParseNamespaceEveryKeywordAsLaterSegment() {
	for _, kw := range allKeywordWords {
		for _, tc := range []struct {
			input string
			parts []string
		}{
			{"namespace a/" + kw, []string{"a", kw}},
			{"namespace a/" + kw + "/b", []string{"a", kw, "b"}},
		} {
			p := NewParserFromString(tc.input, "test.sentrie")
			stmt := parseNamespaceStatement(s.T().Context(), p)
			s.Require().NoError(p.err, "input: %s", tc.input)
			ns, ok := stmt.(*ast.NamespaceStatement)
			s.Require().True(ok, "input: %s", tc.input)
			s.Equal(tc.parts, ns.Name.Parts, "input: %s", tc.input)
		}
	}
}

// TestParseNamespaceEveryKeywordAsFirstSegmentRejected checks that the first
// segment is still an identifier for every reserved word, with the keyword
// named in the error.
func (s *ParserTestSuite) TestParseNamespaceEveryKeywordAsFirstSegmentRejected() {
	for _, kw := range allKeywordWords {
		input := "namespace " + kw + "/a"
		p := NewParserFromString(input, "test.sentrie")
		stmt := parseNamespaceStatement(s.T().Context(), p)
		s.Nil(stmt, "input: %s", input)
		s.Require().Error(p.err, "input: %s", input)
		s.Contains(p.err.Error(), "the first segment of a name cannot be a keyword", "input: %s", input)
		s.Contains(p.err.Error(), "'"+kw+"'", "input: %s", input)
	}
}

// TestParseNamespaceKeywordLikeIdentifiersStayIdentifiers checks that a word
// that merely starts with, or differs in case from, a keyword is an ordinary
// segment: only the exact lower-case reserved word is a keyword.
func (s *ParserTestSuite) TestParseNamespaceKeywordLikeIdentifiersStayIdentifiers() {
	p := NewParserFromString("namespace inside/String/Policy/rules/trueish", "test.sentrie")
	ns, ok := parseNamespaceStatement(s.T().Context(), p).(*ast.NamespaceStatement)
	s.Require().NoError(p.err)
	s.Require().True(ok)
	s.Equal([]string{"inside", "String", "Policy", "rules", "trueish"}, ns.Name.Parts)
}

// TestParseNamespaceMalformedPathAfterSlash checks that accepting keywords after
// a slash did not loosen the segment grammar: anything that is neither an
// identifier nor a keyword is still an error, and the namespace is not built.
func (s *ParserTestSuite) TestParseNamespaceMalformedPathAfterSlash() {
	for _, input := range []string{
		"namespace a/",
		"namespace a//b",
		"namespace a/1",
		`namespace a/"x"`,
		"namespace a/(b)",
		"namespace a/in/",
		"namespace a/in//b",
	} {
		p := NewParserFromString(input, "test.sentrie")
		stmt := parseNamespaceStatement(s.T().Context(), p)
		s.Nil(stmt, "input: %s", input)
		s.Require().Error(p.err, "input: %s", input)
	}
}

// TestParseTypeRefEveryKeywordAsLaterSegment checks the same for a named type,
// the position every keyword-segment shape is referenced from.
func (s *ParserTestSuite) TestParseTypeRefEveryKeywordAsLaterSegment() {
	for _, kw := range allKeywordWords {
		input := "let x: a/" + kw + "/Item = {}"
		p := NewParserFromString(input, "test.sentrie")
		decl, ok := parseLetsStatement(s.T().Context(), p).(*ast.VarDeclaration)
		s.Require().NoError(p.err, "input: %s", input)
		s.Require().True(ok, "input: %s", input)
		s.Equal("a/"+kw+"/Item", decl.Type.String(), "input: %s", input)
	}
}

// TestParseExpressionOperandFailureShapes extends the failed-operand guard to
// the other ways an operand can fail: every shape must end with an error and a
// nil expression, never a panic from a nil left operand reaching an infix,
// call, index, field or ternary handler.
func (s *ParserTestSuite) TestParseExpressionOperandFailureShapes() {
	for _, src := range []string{
		// the right operand fails
		"1 + in",
		"1 + )",
		"x and )",
		"x or in",
		"x == ]",
		"x ? in : 1",
		// a failed prefix is followed by every kind of postfix and infix
		"in(1)",
		"in[0]",
		"in.b",
		"in + 1",
		"in ? 1 : 2",
		"in is defined",
		"in as string",
		"in | f",
		// a failed operand inside a failed operand
		"a/not/f(1) + 2",
		"a/not/b/f(1)",
		"f(a/not/g(1))",
		"[a/not/f(1)]",
		"(a/not/f(1))",
		"x and a/not/f(1)",
	} {
		parser := NewParserFromString(src, "test.sentrie")
		var expr any
		s.NotPanics(func() { expr = parser.parseExpression(s.T().Context(), LOWEST) }, "input: %s", src)
		s.Nil(expr, "input: %s", src)
		s.Error(parser.err, "input: %s", src)
	}
}
