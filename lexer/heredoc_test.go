// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

// lexAll returns every token of input up to, not including, EOF.
func lexAll(t *testing.T, input string) []tokens.Instance {
	t.Helper()
	l := NewLexer(strings.NewReader(input), "t.sentrie")
	var out []tokens.Instance
	for range 100 {
		tok := l.NextToken()
		if tok.Kind == tokens.EOF {
			return out
		}
		out = append(out, tok)
		if tok.Kind == tokens.Error {
			return out
		}
	}
	t.Fatalf("input %q: no EOF after 100 tokens", input)
	return nil
}

// TestLexerHereDocLineEndings checks that a heredoc reads the same in LF and
// CRLF files: the CRLF terminator line ends it, the value has LF line endings
// with no '\r', and the heredoc and the token after it sit on the same lines
// and columns (#270).
func TestLexerHereDocLineEndings(t *testing.T) {
	lines := []string{"let s = <<<EOT", "hello", "  world", "EOT", "x"}
	for _, eol := range []string{"\n", "\r\n"} {
		input := strings.Join(lines, eol) + eol
		toks := lexAll(t, input)
		require.Len(t, toks, 5, "%q", input)

		doc := toks[3]
		require.Equal(t, tokens.String, doc.Kind, "%q: %s", input, doc.Value)
		require.Equal(t, "hello\n  world\n", doc.Value, "%q", input)
		require.Equal(t, tokens.Pos{Line: 1, Column: 9, Offset: 8}, doc.Range.From, "%q", input)
		require.Equal(t, 4, doc.Range.To.Line, "%q", input)
		require.Equal(t, 4, doc.Range.To.Column, "%q", input)
		require.Equal(t, "<<<EOT"+eol+"hello"+eol+"  world"+eol+"EOT", input[doc.Range.From.Offset:doc.Range.To.Offset], "%q", input)

		x := toks[4]
		require.Equal(t, tokens.Ident, x.Kind, "%q", input)
		require.Equal(t, 5, x.Range.From.Line, "%q", input)
		require.Equal(t, 1, x.Range.From.Column, "%q", input)
		require.Equal(t, "x", input[x.Range.From.Offset:x.Range.To.Offset], "%q", input)
	}
}

// TestLexerHereDocTerminatorTrailingWhitespace checks that spaces, tabs and a
// '\r' after the terminator tag still end the heredoc, and that the heredoc
// ends just past the tag rather than past the whitespace (#270).
func TestLexerHereDocTerminatorTrailingWhitespace(t *testing.T) {
	for _, input := range []string{
		"<<<T\na\nT  \nx",
		"<<<T\na\nT\t\nx",
		"<<<T\na\nT \t \nx",
		"<<<T\r\na\r\nT \t\r\nx",
	} {
		toks := lexAll(t, input)
		require.Len(t, toks, 2, "%q", input)
		require.Equal(t, tokens.String, toks[0].Kind, "%q: %s", input, toks[0].Value)
		require.Equal(t, "a\n", toks[0].Value, "%q", input)
		require.Equal(t, 3, toks[0].Range.To.Line, "%q", input)
		require.Equal(t, 2, toks[0].Range.To.Column, "%q", input)
		require.True(t, strings.HasSuffix(input[:toks[0].Range.To.Offset], "\nT"), "%q", input)
		require.Equal(t, tokens.Ident, toks[1].Kind, "%q", input)
		require.Equal(t, 4, toks[1].Range.From.Line, "%q", input)
		require.Equal(t, 1, toks[1].Range.From.Column, "%q", input)
	}
}

// TestLexerHereDocTerminatorAtEOF checks a terminator that ends the input with
// no newline after it, with and without trailing whitespace and in a CRLF file.
func TestLexerHereDocTerminatorAtEOF(t *testing.T) {
	for _, input := range []string{
		"<<<T\na\nT",
		"<<<T\r\na\r\nT",
		"<<<T\na\nT \t",
	} {
		toks := lexAll(t, input)
		require.Len(t, toks, 1, "%q", input)
		require.Equal(t, tokens.String, toks[0].Kind, "%q: %s", input, toks[0].Value)
		require.Equal(t, "a\n", toks[0].Value, "%q", input)
		require.Equal(t, 3, toks[0].Range.To.Line, "%q", input)
		require.Equal(t, 2, toks[0].Range.To.Column, "%q", input)
	}
}

// TestLexerHereDocBodyKeepsOtherWhitespace checks that only a CRLF's '\r' is
// dropped from a body line: trailing spaces and tabs stay in the value.
func TestLexerHereDocBodyKeepsOtherWhitespace(t *testing.T) {
	toks := lexAll(t, "<<<T\r\na \t\r\nT\r\n")
	require.Len(t, toks, 1)
	require.Equal(t, "a \t\n", toks[0].Value)
}

// TestLexerHereDocUnterminated checks that a heredoc with no terminator line,
// including one whose only candidate is indented, reports the heredoc's own
// start and tag rather than the end of the file (#270).
func TestLexerHereDocUnterminated(t *testing.T) {
	for _, input := range []string{
		"x = <<<EOT\nbody\n",
		"x = <<<EOT\r\nbody\r\n",
		"x = <<<EOT\nbody\n  EOT\n",
		"x = <<<EOT\nbody\nEOTX\n",
	} {
		toks := lexAll(t, input)
		last := toks[len(toks)-1]
		require.Equal(t, tokens.Error, last.Kind, "%q", input)
		require.Equal(t, `heredoc <<<EOT has no terminator line "EOT": unterminated string literal: at t.sentrie:1:5`, last.Value, "%q", input)
		require.Equal(t, tokens.Pos{Line: 1, Column: 5, Offset: 4}, last.Range.From, "%q", input)
	}
}

// TestLexerHereDocCRLFBodyLines checks how a CRLF body line becomes a value
// line: an empty CRLF line stays an empty line, a '\r' that is not at the end
// of a line stays, and LF and CRLF lines can mix in one heredoc (#270).
func TestLexerHereDocCRLFBodyLines(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty CRLF line", "<<<T\r\na\r\n\r\nb\r\nT\r\n", "a\n\nb\n"},
		{"empty CRLF line before the terminator", "<<<T\r\na\r\n\r\nT\r\n", "a\n\n"},
		{"only an empty CRLF line", "<<<T\r\n\r\nT\r\n", "\n"},
		{"no body", "<<<T\r\nT\r\n", ""},
		{"lone CR inside a line", "<<<T\na\rb\nT\n", "a\rb\n"},
		{"lone CR inside a CRLF line", "<<<T\r\na\rb\r\nT\r\n", "a\rb\n"},
		{"two CRs at a line end drop only one", "<<<T\r\na\r\r\nT\r\n", "a\r\n"},
		{"mixed endings, CRLF first", "<<<T\r\na\r\nb\nc\r\nT\n", "a\nb\nc\n"},
		{"mixed endings, LF first", "<<<T\na\nb\r\nc\nT\r\n", "a\nb\nc\n"},
		{"CRLF body, LF terminator", "<<<T\r\na\r\nT\nx", "a\n"},
		{"LF body, CRLF terminator", "<<<T\na\nT\r\nx", "a\n"},
		{"padded tag line, CRLF", "<<<T \t\r\na\r\nT\r\n", "a\n"},
	}
	for _, tc := range cases {
		toks := lexAll(t, tc.input)
		require.NotEmpty(t, toks, "%s: %q", tc.name, tc.input)
		require.Equal(t, tokens.String, toks[0].Kind, "%s: %q: %s", tc.name, tc.input, toks[0].Value)
		require.Equal(t, tc.want, toks[0].Value, "%s: %q", tc.name, tc.input)
	}
}

// TestLexerHereDocLinesThatOnlyContainTheTag checks that a line ends a
// heredoc only when it is the tag and nothing else but trailing whitespace: a
// line with other text before or after the tag, or with the tag repeated, is
// body (#270).
func TestLexerHereDocLinesThatOnlyContainTheTag(t *testing.T) {
	for _, eol := range []string{"\n", "\r\n"} {
		body := []string{"T x", "TT", "xT", "T.", "T\rx", "T;", "t"}
		input := "<<<T" + eol + strings.Join(body, eol) + eol + "T" + eol + "y"
		toks := lexAll(t, input)
		require.Len(t, toks, 2, "%q", input)
		require.Equal(t, tokens.String, toks[0].Kind, "%q: %s", input, toks[0].Value)
		require.Equal(t, "T x\nTT\nxT\nT.\nT\rx\nT;\nt\n", toks[0].Value, "%q", input)
		require.Equal(t, tokens.Ident, toks[1].Kind, "%q", input)
		require.Equal(t, "y", toks[1].Value, "%q", input)
	}
}

// TestLexerTwoHereDocsInACRLFFile checks that the second heredoc starts and
// ends where it is written when the first one is in a CRLF file, so the first
// leaves nothing behind (#270).
func TestLexerTwoHereDocsInACRLFFile(t *testing.T) {
	lines := []string{"a = <<<ONE", "1", "ONE  ", "b = <<<TWO", "", "2", "TWO", "c"}
	input := strings.Join(lines, "\r\n") + "\r\n"
	toks := lexAll(t, input)
	require.Len(t, toks, 7, "%q", input)

	one, two := toks[2], toks[5]
	require.Equal(t, tokens.String, one.Kind, "%s", one.Value)
	require.Equal(t, "1\n", one.Value)
	require.Equal(t, "<<<ONE\r\n1\r\nONE", input[one.Range.From.Offset:one.Range.To.Offset])
	require.Equal(t, 3, one.Range.To.Line)
	require.Equal(t, 4, one.Range.To.Column)

	require.Equal(t, tokens.Ident, toks[3].Kind)
	require.Equal(t, 4, toks[3].Range.From.Line)
	require.Equal(t, 1, toks[3].Range.From.Column)

	require.Equal(t, tokens.String, two.Kind, "%s", two.Value)
	require.Equal(t, "\n2\n", two.Value)
	require.Equal(t, 4, two.Range.From.Line)
	require.Equal(t, 5, two.Range.From.Column)
	require.Equal(t, "<<<TWO\r\n\r\n2\r\nTWO", input[two.Range.From.Offset:two.Range.To.Offset])
	require.Equal(t, 7, two.Range.To.Line)
	require.Equal(t, 4, two.Range.To.Column)

	c := toks[6]
	require.Equal(t, tokens.Ident, c.Kind)
	require.Equal(t, 8, c.Range.From.Line)
	require.Equal(t, 1, c.Range.From.Column)
}
