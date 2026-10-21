// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
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
