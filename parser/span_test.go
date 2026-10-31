// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// TestStatementSpanEndsAtLastToken checks that every single-line statement
// kind spans from its first token to just past its last token, on the first
// token's line, whatever follows it: a newline, CRLF, a space, a semicolon,
// another token on the same line, a comment or the end of the input (#236).
// Before the fix a statement followed by a newline ended at column 0 of the
// next line, and one at EOF ended one column short.
func (s *ParserTestSuite) TestStatementSpanEndsAtLastToken() {
	type stmtCase struct {
		name   string
		policy bool // parsed as a policy-level statement, else a top-level one
		src    string
		// endsInExpr is set when the statement ends in an expression, which
		// wraps a trailing comment into itself and so spans it.
		endsInExpr bool
	}
	statements := []stmtCase{
		{"namespace", false, "namespace com/example", false},
		{"shape", false, "shape Person { name: string }", false},
		{"shape alias", false, "shape Name string", false},
		{"export shape", false, "export shape Person", false},
		{"export derive", false, "export derive double", false},
		{"title", true, `title "Pricing"`, false},
		{"description", true, `description "Prices things"`, false},
		{"version", true, `version "1.0.0"`, false},
		{"tag", true, `tag "team" = "billing"`, false},
		{"fact", true, "fact amount: number", false},
		{"fact optional", true, "fact amount: number?", false},
		{"fact default", true, "fact amount: number default 5", true},
		{"let", true, "let x = 1", true},
		{"let typed", true, "let x: number = 1", true},
		{"let list", true, "let x = [1, 2]", true},
		{"let dict", true, `let x = {"a": 1}`, true},
		{"rule", true, "rule r = default false true", true},
		{"export decision", true, "export decision of r", false},
		{"export decision attach", true, "export decision of r attach a as 1", true},
		{"export decision attach list", true, "export decision of r attach a as [1, 2]", true},
		{"use", true, "use { x } from @sentrie/time as t", false},
		// Statements that end in a block expression end just past its `}`.
		{"let block", true, "let x = { yield 1 }", true},
		{"let lambda", true, "let f = (x) => { yield x }", true},
		{"rule block", true, "rule r = default false { yield true }", true},
		{"derive", true, "derive d = (a: number) => { yield a }", true},
		{"derive top-level", false, "derive d = (a: number) => { yield a }", true},
		{"export decision attach block", true, "export decision of r attach a as { yield 1 }", true},
	}
	// What follows the statement. The statement itself is never read past its
	// last token, so each suffix only changes what the lexer sees after it.
	suffixes := []struct{ name, text string }{
		{"newline", "\n  rule"},
		{"CRLF", "\r\n  rule"},
		{"space", " "},
		{"semicolon", ";"},
		{"same line", " ; rule"},
		{"trailing comment", " -- note"},
		{"EOF", ""},
	}

	for _, stmt := range statements {
		for _, suffix := range suffixes {
			if stmt.endsInExpr && suffix.name == "trailing comment" {
				continue
			}
			// The statement sits on the second line, indented by two spaces.
			input := "\n  " + stmt.src + suffix.text
			label := fmt.Sprintf("%s followed by %s: %q", stmt.name, suffix.name, input)
			p := NewParserFromString(input, "t.sentrie")
			var parsed ast.Statement
			if stmt.policy {
				parsed = parsePolicyStatement(s.T().Context(), p)
			} else {
				parsed = parseStatement(s.T().Context(), p)
			}
			s.Require().NoError(p.err, label)
			s.Require().NotNil(parsed, label)

			want := tokens.NewRange("t.sentrie",
				tokens.Pos{Line: 2, Column: 3, Offset: 3},
				tokens.Pos{Line: 2, Column: 3 + utf8.RuneCountInString(stmt.src), Offset: 3 + len(stmt.src)},
			)
			s.Equal(want, parsed.Span(), label)
			s.Equal(stmt.src, input[want.From.Offset:want.To.Offset], label)
		}
	}
}

// TestStatementSpanCountsColumnsInRunesAndOffsetsInBytes pins how a multi-byte
// string literal moves a statement's end: the column advances by one per rune
// and the offset by one per byte (#236).
func (s *ParserTestSuite) TestStatementSpanCountsColumnsInRunesAndOffsetsInBytes() {
	testCases := []struct {
		name string
		src  string
		to   tokens.Pos
	}{
		// "café" is 4 runes and 5 bytes, so the end sits one byte further than
		// one column per byte would put it.
		{"2-byte rune", `title "café"`, tokens.Pos{Line: 1, Column: 13, Offset: 13}},
		{"3-byte rune", `title "日本"`, tokens.Pos{Line: 1, Column: 11, Offset: 14}},
		{"4-byte rune", `title "😀"`, tokens.Pos{Line: 1, Column: 10, Offset: 12}},
		{"keys and values", `tag "é" = "ü"`, tokens.Pos{Line: 1, Column: 14, Offset: 15}},
	}
	for _, tc := range testCases {
		for _, suffix := range []string{"", "\n", "\r\n"} {
			input := tc.src + suffix
			p := NewParserFromString(input, "t.sentrie")
			stmt := parsePolicyStatement(s.T().Context(), p)
			s.Require().NoError(p.err, "%s: %q", tc.name, input)
			s.Equal(tokens.Pos{Line: 1, Column: 1, Offset: 0}, stmt.Span().From, "%s: %q", tc.name, input)
			s.Equal(tc.to, stmt.Span().To, "%s: %q", tc.name, input)
		}
	}
}

// spanCase is one statement of the program in TestProgramStatementSpans: its
// exact source text, where it starts and ends (1-based line and column, the
// end just past the last rune) and, for spans that cross lines, the printed
// range.
type spanCase struct {
	text                             string
	fromLine, fromCol, toLine, toCol int
	printed                          string
}

// TestProgramStatementSpans parses a whole program, with LF and with CRLF line
// endings, and checks every statement's span through the real statement
// dispatchers. Statements sit at the start of a line, indented, after a
// multi-byte string and next to another statement on the same line, and some
// span several lines (#236). A heredoc whose terminator has trailing
// whitespace ends just past its tag, in a CRLF file too (#270). Offsets are checked by slicing the source: the
// span must cover exactly the statement's text.
func (s *ParserTestSuite) TestProgramStatementSpans() {
	lines := []string{
		/*  1 */ "namespace com/example",
		/*  2 */ "",
		/*  3 */ "shape Person {",
		/*  4 */ "  name: string",
		/*  5 */ "  age: number",
		/*  6 */ "}",
		/*  7 */ "export shape Person",
		/*  8 */ "",
		/*  9 */ "policy p {",
		/* 10 */ `  title "café"`,
		/* 11 */ `  description "d"`,
		/* 12 */ `  version "1.0"`,
		/* 13 */ `  tag "é" = "ü"; fact f: string`,
		/* 14 */ "  let x = 1",
		/* 15 */ "  let y = [",
		/* 16 */ "    1,",
		/* 17 */ "    2",
		/* 18 */ "  ]",
		/* 19 */ "  use { x } from @sentrie/time as t",
		/* 20 */ "  rule r = default false true",
		/* 21 */ "  let h = <<<EOT",
		/* 22 */ "hello",
		/* 23 */ "EOT \t",
		/* 24 */ "  export decision of r",
		/* 25 */ "}",
	}
	const file = "t.sentrie"
	want := []spanCase{
		{"namespace com/example", 1, 1, 1, 22, ""},
		{"shape Person {\n  name: string\n  age: number\n}", 3, 1, 6, 2, "t.sentrie:3:1-6:2"},
		{"export shape Person", 7, 1, 7, 20, "t.sentrie:7:1-20"},
		{"policy p {\n" + `  title "café"` + "\n" + `  description "d"` + "\n" + `  version "1.0"` + "\n" +
			`  tag "é" = "ü"; fact f: string` + "\n  let x = 1\n  let y = [\n    1,\n    2\n  ]\n" +
			"  use { x } from @sentrie/time as t\n  rule r = default false true\n" +
			"  let h = <<<EOT\nhello\nEOT \t\n  export decision of r\n}",
			9, 1, 25, 2, "t.sentrie:9:1-25:2"},
	}
	inPolicy := []spanCase{
		{`title "café"`, 10, 3, 10, 15, "t.sentrie:10:3-15"},
		{`description "d"`, 11, 3, 11, 18, ""},
		{`version "1.0"`, 12, 3, 12, 16, ""},
		{`tag "é" = "ü"`, 13, 3, 13, 16, ""},
		{"fact f: string", 13, 18, 13, 32, "t.sentrie:13:18-32"},
		{"let x = 1", 14, 3, 14, 12, ""},
		{"let y = [\n    1,\n    2\n  ]", 15, 3, 18, 4, "t.sentrie:15:3-18:4"},
		{"use { x } from @sentrie/time as t", 19, 3, 19, 36, ""},
		{"rule r = default false true", 20, 3, 20, 30, ""},
		{"let h = <<<EOT\nhello\nEOT", 21, 3, 23, 4, "t.sentrie:21:3-23:4"},
		{"export decision of r", 24, 3, 24, 23, "t.sentrie:24:3-23"},
	}

	for _, eol := range []string{"\n", "\r\n"} {
		name := map[string]string{"\n": "LF", "\r\n": "CRLF"}[eol]
		input := strings.Join(lines, eol) + eol
		program, err := NewParserFromString(input, file).ParseProgram(s.T().Context())
		s.Require().NoError(err, name)

		var topLevel, nested []ast.Statement
		for _, stmt := range program.Statements {
			topLevel = append(topLevel, stmt)
			if policy, ok := stmt.(*ast.PolicyStatement); ok {
				nested = policy.Statements
			}
		}
		s.Require().Len(topLevel, len(want), name)
		s.Require().Len(nested, len(inPolicy), name)

		check := func(got ast.Statement, c spanCase) {
			label := fmt.Sprintf("%s: %q", name, c.text)
			span := got.Span()
			s.Equal(tokens.Pos{Line: c.fromLine, Column: c.fromCol}, tokens.Pos{Line: span.From.Line, Column: span.From.Column}, label)
			s.Equal(tokens.Pos{Line: c.toLine, Column: c.toCol}, tokens.Pos{Line: span.To.Line, Column: span.To.Column}, label)
			// The offsets cover exactly the statement's own text, line endings
			// included, so a multi-line span is one byte longer per line under CRLF.
			s.Require().LessOrEqual(span.To.Offset, len(input), label)
			s.Equal(strings.ReplaceAll(c.text, "\n", eol), input[span.From.Offset:span.To.Offset], label)
			if c.printed != "" {
				s.Equal(c.printed, span.String(), label)
			}
		}
		for i, c := range want {
			check(topLevel[i], c)
		}
		for i, c := range inPolicy {
			check(nested[i], c)
		}
	}
}

// TestStatementSpansAroundHeredoc checks a heredoc value: the statement ends
// just past the terminator tag on its own line, and the statement after it
// starts where it is written. The heredoc used to swallow the newline after
// its terminator, which moved the end onto the next line (#236).
func (s *ParserTestSuite) TestStatementSpansAroundHeredoc() {
	input := "namespace a\npolicy p {\n  let x = <<<T\nhello\nT\n  fact f: string\n}\n"
	program, err := NewParserFromString(input, "t.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
	s.Require().Len(program.Statements, 2)
	policy, ok := program.Statements[1].(*ast.PolicyStatement)
	s.Require().True(ok)
	s.Require().Len(policy.Statements, 2)

	letStmt := policy.Statements[0]
	s.Equal(tokens.Pos{Line: 3, Column: 3, Offset: 25}, letStmt.Span().From)
	s.Equal(tokens.Pos{Line: 5, Column: 2, Offset: 45}, letStmt.Span().To)
	s.Equal("let x = <<<T\nhello\nT", input[letStmt.Span().From.Offset:letStmt.Span().To.Offset])
	s.Equal("t.sentrie:3:3-5:2", letStmt.Span().String())

	factStmt := policy.Statements[1]
	s.Equal(tokens.Pos{Line: 6, Column: 3, Offset: 48}, factStmt.Span().From)
	s.Equal(tokens.Pos{Line: 6, Column: 17, Offset: 62}, factStmt.Span().To)

	// A heredoc that ends the input has no newline after its terminator.
	p := NewParserFromString("\n  let x = <<<T\nhello\nT", "t.sentrie")
	stmt := parsePolicyStatement(s.T().Context(), p)
	s.Require().NoError(p.err)
	s.Equal(tokens.Pos{Line: 2, Column: 3, Offset: 3}, stmt.Span().From)
	s.Equal(tokens.Pos{Line: 4, Column: 2, Offset: 23}, stmt.Span().To)
}

// TestStatementSpanOfMultiLineStatementsPrintsBothLines checks that a statement
// crossing lines ends on the line of its last token and that the printed range
// names that line, not the one before it (#236). A block expression ends just
// past its `}`, so the rule does too.
func (s *ParserTestSuite) TestStatementSpanOfMultiLineStatementsPrintsBothLines() {
	p := NewParserFromString("\n  rule r = default false {\n    yield true\n  }\n", "t.sentrie")
	stmt := parsePolicyStatement(s.T().Context(), p)
	s.Require().NoError(p.err)
	s.Equal(tokens.Pos{Line: 2, Column: 3, Offset: 3}, stmt.Span().From)
	s.Equal(tokens.Pos{Line: 4, Column: 4, Offset: 46}, stmt.Span().To)
	s.Equal("t.sentrie:2:3-4:4", stmt.Span().String())
}

// TestCommentStatementSpanCoversTheComment checks that a comment statement
// spans from its `--` to just past its last character, not zero width at its
// start. A comment runs to the end of its line, so only a newline or the end
// of the input can follow it.
func (s *ParserTestSuite) TestCommentStatementSpanCoversTheComment() {
	for _, input := range []string{"\n  -- a note\n  rule", "\n  -- a note"} {
		for _, policy := range []bool{true, false} {
			p := NewParserFromString(input, "t.sentrie")
			var stmt ast.Statement
			if policy {
				stmt = parsePolicyStatement(s.T().Context(), p)
			} else {
				stmt = parseStatement(s.T().Context(), p)
			}
			s.Require().NoError(p.err, "input: %q", input)
			s.Equal(tokens.NewRange("t.sentrie",
				tokens.Pos{Line: 2, Column: 3, Offset: 3},
				tokens.Pos{Line: 2, Column: 12, Offset: 12},
			), stmt.Span(), "input: %q policy: %v", input, policy)
		}
	}
}

// TestParseErrorsPrintOneBasedPositions pins the position users see in a
// diagnostic: lines and columns are 1-based on the first line and on later
// lines alike, and a range that crosses lines prints both of its lines as
// stored (#236).
func (s *ParserTestSuite) TestParseErrorsPrintOneBasedPositions() {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			"parse error on line 1",
			"shape 5",
			"parsing error at t.sentrie:1:7-8: expected Ident, got Int at t.sentrie:1:7-8",
		},
		{
			"parse error on a later line",
			"namespace a\n\nshape 5",
			"parsing error at t.sentrie:3:7-8: expected Ident, got Int at t.sentrie:3:7-8",
		},
		{
			"lexer error on line 1",
			"$",
			"parsing error at t.sentrie:1:1-2: unexpected token 'Error'",
		},
		{
			"lexer error on a later line",
			"namespace a\n  $",
			"parsing error at t.sentrie:2:3-4: unexpected token 'Error'",
		},
		{
			"statement at the end of a line is followed by no position on the next line",
			"namespace a\nnamespace b\n",
			"namespace cannot be declared after namespace declaration at t.sentrie:2:1-12",
		},
		{
			"multi-line range prints its end line",
			"shape A {\n  a: string\n}\n",
			"program must start with namespace, got *ast.ShapeStatement at t.sentrie:1:1-3:2",
		},
		{
			"multi-line range on a later line",
			"namespace a\nshape A {\n  a: string\n}\nnamespace b",
			"namespace cannot be declared after namespace declaration at t.sentrie:5:1-12",
		},
	}
	for _, tc := range testCases {
		_, err := NewParserFromString(tc.input, "t.sentrie").ParseProgram(s.T().Context())
		s.Require().Error(err, tc.name)
		s.Equal(tc.want, err.Error(), tc.name)
	}
}
