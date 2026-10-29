// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"unicode/utf8"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// TestParseUseStatementKeywordPathSegments checks that keywords are accepted as
// segments of an `@scope/name` module path (#158).
func (s *ParserTestSuite) TestParseUseStatementKeywordPathSegments() {
	testCases := []struct {
		input   string
		libFrom []string
		alias   string
	}{
		{"use { lower, upper } from @sentrie/string as str", []string{"sentrie", "string"}, "str"},
		{"use { abs } from @sentrie/math", []string{"sentrie", "math"}, "math"},
		{"use { f } from @sentrie/number/helpers", []string{"sentrie", "number", "helpers"}, "helpers"},
		{"use { f } from @policy/string/util as u", []string{"policy", "string", "util"}, "u"},
		{"use { f } from @local/document/rule as r", []string{"local", "document", "rule"}, "r"},
	}

	for _, tc := range testCases {
		p := NewParserFromString(tc.input, "test.sentrie")
		stmt := parseUseStatement(s.T().Context(), p)
		s.Require().NoError(p.err, "input: %s", tc.input)
		useStmt, ok := stmt.(*ast.UseStatement)
		s.Require().True(ok, "input: %s", tc.input)
		s.Equal(tc.libFrom, useStmt.LibFrom, "input: %s", tc.input)
		s.Equal(tc.alias, useStmt.As, "input: %s", tc.input)
		s.Empty(useStmt.RelativeFrom, "input: %s", tc.input)
	}
}

// TestParseUseStatementKeywordDefaultAliasRejected checks that a keyword last
// segment without `as` is a clear parse error, since the default alias would
// not be usable as an identifier at call sites.
func (s *ParserTestSuite) TestParseUseStatementKeywordDefaultAliasRejected() {
	p := NewParserFromString("use { lower } from @sentrie/string", "test.sentrie")
	stmt := parseUseStatement(s.T().Context(), p)
	s.Nil(stmt)
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), `default alias "string" is a keyword; add 'as <alias>'`)
	// The diagnostic spans `string`, not the token after the statement.
	s.Contains(p.err.Error(), "parsing error at test.sentrie:1:28-33:")
}

// TestParseUseStatementTrailingSlashBeforeAs pins how `@sentrie/ as x` fails:
// `as` is read as a keyword path segment, so the statement is rejected by the
// keyword default-alias check. A future change must not let it parse as module
// `@sentrie/as` with alias `x`.
func (s *ParserTestSuite) TestParseUseStatementTrailingSlashBeforeAs() {
	p := NewParserFromString("use { f } from @sentrie/ as x", "test.sentrie")
	stmt := parseUseStatement(s.T().Context(), p)
	s.Nil(stmt)
	s.Require().Error(p.err)
	s.Contains(p.err.Error(), `module @sentrie/as: default alias "as" is a keyword; add 'as <alias>'`)
}

// TestParseUseStatementKeywordsStillReservedElsewhere checks that accepting
// keywords in the module path does not leak into the other positions of a use
// statement, which still require identifiers.
func (s *ParserTestSuite) TestParseUseStatementKeywordsStillReservedElsewhere() {
	testCases := []string{
		"use { string } from @sentrie/string as str",   // keyword in the import list
		"use { lower } from @sentrie/string as string", // keyword as the explicit alias
		"use { lower } from @sentrie/",                 // missing segment after '/'
		"use { lower } from @123",                      // a number is not a segment
	}

	for _, tc := range testCases {
		p := NewParserFromString(tc, "test.sentrie")
		stmt := parseUseStatement(s.T().Context(), p)
		s.Error(p.err, "input: %s", tc)
		s.Nil(stmt, "input: %s", tc)
	}
}

// TestParseProgramWithKeywordModulePath checks the whole-program path, so the
// policy statement dispatcher reaches the use statement with a keyword segment.
func (s *ParserTestSuite) TestParseProgramWithKeywordModulePath() {
	input := `namespace com/example

policy p {
  use { lower } from @sentrie/string as str
  rule r = default false {
    yield str.lower("A") == "a"
  }
  export decision of r
}
`
	program, err := NewParserFromString(input, "test.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
	s.Require().NotNil(program)
}

// TestParseUseStatementSpanEndsAtLastToken checks that a use statement's span
// runs from `use` to the end of its last token (the alias, or the module path
// when there is no `as`) for every `from` form, whatever follows the statement
// (#216). It used to end at column 0 of the next line when a newline followed.
func (s *ParserTestSuite) TestParseUseStatementSpanEndsAtLastToken() {
	statements := []string{
		"use { x } from @sentrie/time as t",
		"use { x } from @sentrie/time",
		"use { x, y } from @local/a/b as ab",
		"use { x } from @local",
		"use { lower } from @sentrie/string as str",
		`use { y } from "./m.ts" as m`,
		`use { y } from "./m.ts"`,
		`use { y } from "./sub\/m.ts"`,
		`use { y } from "./é/m.ts"`,
	}
	// What follows the statement: a newline, a space, or the end of the input.
	suffixes := []string{"\n  rule", " ", ""}

	for _, stmtSrc := range statements {
		for _, suffix := range suffixes {
			// The statement sits on the second line, indented by two spaces.
			input := "\n  " + stmtSrc + suffix
			p := NewParserFromString(input, "t.sentrie")
			stmt := parseUseStatement(s.T().Context(), p)
			s.Require().NoError(p.err, "input: %q", input)
			useStmt, ok := stmt.(*ast.UseStatement)
			s.Require().True(ok, "input: %q", input)

			start := tokens.Pos{Line: 1, Column: 3, Offset: 3}
			end := tokens.Pos{
				Line:   1,
				Column: start.Column + utf8.RuneCountInString(stmtSrc),
				Offset: start.Offset + len(stmtSrc),
			}
			s.Equal(tokens.NewRange("t.sentrie", start, end), useStmt.Span(), "input: %q", input)
		}
	}
}

// TestParseUseStatementSpanInProgram checks the span through the policy
// statement dispatcher, with use statements followed by further statements.
func (s *ParserTestSuite) TestParseUseStatementSpanInProgram() {
	input := `namespace com/example

policy p {
  use { x } from @sentrie/time as t
  use { y } from "./m.ts" as m
  rule r = default false {
    yield true
  }
  export decision of r
}
`
	program, err := NewParserFromString(input, "t.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)

	var uses []*ast.UseStatement
	for _, stmt := range program.Statements {
		if policy, ok := stmt.(*ast.PolicyStatement); ok {
			for _, inner := range policy.Statements {
				if use, ok := inner.(*ast.UseStatement); ok {
					uses = append(uses, use)
				}
			}
		}
	}
	s.Require().Len(uses, 2)

	// Lines are 0-based; on every line after the first, columns are 1-based.
	s.Equal(tokens.Pos{Line: 3, Column: 3, Offset: 36}, uses[0].Span().From)
	s.Equal(tokens.Pos{Line: 3, Column: 36, Offset: 69}, uses[0].Span().To)
	s.Equal(tokens.Pos{Line: 4, Column: 3, Offset: 72}, uses[1].Span().From)
	s.Equal(tokens.Pos{Line: 4, Column: 31, Offset: 100}, uses[1].Span().To)
}
