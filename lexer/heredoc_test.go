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
