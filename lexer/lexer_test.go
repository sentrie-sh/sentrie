// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

func collectKinds(input string) []tokens.Kind {
	l := NewLexer(strings.NewReader(input), "test.sent")
	kinds := []tokens.Kind{}
	for {
		tok := l.NextToken()
		kinds = append(kinds, tok.Kind)
		if tok.Kind == tokens.EOF || tok.Kind == tokens.Error {
			break
		}
	}
	return kinds
}

func mustNextToken(t *testing.T, l *Lexer) tokens.Instance {
	t.Helper()
	tok := l.NextToken()
	return tok
}

func TestLexerHereDocRejectsInvalidTagStart(t *testing.T) {
	l := NewLexer(strings.NewReader("<<<1TAG\nbody\nTAG\n"), "test.sent")

	tok := l.NextToken()
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s", tok.Kind)
	}
	if !strings.Contains(tok.Value, "heredoc requires identifier tag") {
		t.Fatalf("unexpected literal: %q", tok.Value)
	}
}

func TestLexerPipelineToken(t *testing.T) {
	l := NewLexer(strings.NewReader("value |> len"), "test.sent")

	tok := mustNextToken(t, l)
	if tok.Kind != tokens.Ident || tok.Value != "value" {
		t.Fatalf("expected first ident token, got %s(%q)", tok.Kind, tok.Value)
	}

	tok = mustNextToken(t, l)
	if tok.Kind != tokens.TokenPipeForward || tok.Value != "|>" {
		t.Fatalf("expected pipeline token, got %s(%q)", tok.Kind, tok.Value)
	}

	tok = mustNextToken(t, l)
	if tok.Kind != tokens.Ident || tok.Value != "len" {
		t.Fatalf("expected rhs ident token, got %s(%q)", tok.Kind, tok.Value)
	}
}

func TestLexerPipelineNoWhitespace(t *testing.T) {
	got := collectKinds("a|>b")
	want := []tokens.Kind{tokens.Ident, tokens.TokenPipeForward, tokens.Ident, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestLexerPipelineMultilineSequence(t *testing.T) {
	got := collectKinds("value\n|> string.trim\n|> len")
	want := []tokens.Kind{
		tokens.Ident,
		tokens.TokenPipeForward,
		tokens.KeywordString,
		tokens.TokenDot,
		tokens.Ident,
		tokens.TokenPipeForward,
		tokens.Ident,
		tokens.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestLexerRejectsBarePipe(t *testing.T) {
	l := NewLexer(strings.NewReader("value | len"), "test.sent")
	_ = mustNextToken(t, l) // value
	tok := mustNextToken(t, l)
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s", tok.Kind)
	}
	if !strings.Contains(tok.Value, "only '|>' is supported") {
		t.Fatalf("expected pipe error message, got %q", tok.Value)
	}
}

func TestLexerRejectsTrailingPipeAtEOF(t *testing.T) {
	l := NewLexer(strings.NewReader("value |"), "test.sent")
	_ = mustNextToken(t, l) // value
	tok := mustNextToken(t, l)
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s", tok.Kind)
	}
}

func TestLexerRejectsDoublePipe(t *testing.T) {
	l := NewLexer(strings.NewReader("value || len"), "test.sent")
	_ = mustNextToken(t, l) // value
	tok := mustNextToken(t, l)
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s", tok.Kind)
	}
}

func TestLexerPipelineHoleToken(t *testing.T) {
	got := collectKinds("#")
	want := []tokens.Kind{tokens.TokenPipelineHole, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestLexerPipelineHoleInsideCall(t *testing.T) {
	got := collectKinds("f(a, #, b)")
	want := []tokens.Kind{
		tokens.Ident,
		tokens.PunctLeftParentheses,
		tokens.Ident,
		tokens.PunctComma,
		tokens.TokenPipelineHole,
		tokens.PunctComma,
		tokens.Ident,
		tokens.PunctRightParentheses,
		tokens.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

// TestLexerTokenPositions checks that lines and columns are 1-based on every
// line and that a token ends just past its last rune on its own line, whether
// a space, a newline or EOF follows it (#236).
func TestLexerTokenPositions(t *testing.T) {
	type span struct{ fromLine, fromCol, toLine, toCol, toOff int }
	testCases := []struct {
		input string
		want  []span
	}{
		{"ab cd", []span{{1, 1, 1, 3, 2}, {1, 4, 1, 6, 5}}},
		{"ab\ncd\n", []span{{1, 1, 1, 3, 2}, {2, 1, 2, 3, 5}}},
		{"x == y\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}, {1, 6, 1, 7, 6}}},
		{"\"é\"\n-- note\n", []span{{1, 1, 1, 4, 4}, {2, 1, 2, 8, 12}}},
		{"<<<T\na\nT\nx", []span{{1, 1, 3, 2, 8}, {4, 1, 4, 2, 10}}},
		// These operators took their end before their last read and ended one
		// column short; check them mid-line, at the end of a line and at EOF.
		{"a-b", []span{{1, 1, 1, 2, 1}, {1, 2, 1, 3, 2}, {1, 3, 1, 4, 3}}},
		{"x -\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 4, 3}}},
		{"x -", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 4, 3}}},
		{"x ==\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x ==", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x =>\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x =>", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x |>\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x |>", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
	}
	for _, tc := range testCases {
		l := NewLexer(strings.NewReader(tc.input), "t.sentrie")
		for i, want := range tc.want {
			tok := l.NextToken()
			got := span{tok.Range.From.Line, tok.Range.From.Column, tok.Range.To.Line, tok.Range.To.Column, tok.Range.To.Offset}
			if got != want {
				t.Fatalf("input %q token %d (%s %q): got %+v, want %+v", tc.input, i, tok.Kind, tok.Value, got, want)
			}
		}
		if eof := l.NextToken(); eof.Kind != tokens.EOF {
			t.Fatalf("input %q: expected EOF, got %s", tc.input, eof.Kind)
		}
	}
}

// TestLexerStringLiteralLineBreaks checks that a line break inside a string
// literal is \n whether the file uses LF or CRLF, while an escaped \r and a
// lone raw \r are kept (#280).
func TestLexerStringLiteralLineBreaks(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{"LF", "\"a\nb\"", "a\nb"},
		{"CRLF", "\"a\r\nb\"", "a\nb"},
		{"two CRLF breaks", "\"a\r\n\r\nb\"", "a\n\nb"},
		{"escaped CR before an LF break", "\"a\\r\nb\"", "a\r\nb"},
		{"escaped CR before a CRLF break", "\"a\\r\r\nb\"", "a\r\nb"},
		{"escaped CR alone", "\"a\\rb\"", "a\rb"},
		{"lone raw CR", "\"a\rb\"", "a\rb"},
		{"lone raw CR before the closing quote", "\"a\r\"", "a\r"},
		{"escaped line break, LF", "\"a\\\nb\"", "a\nb"},
		{"escaped line break, CRLF", "\"a\\\r\nb\"", "a\nb"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok := NewLexer(strings.NewReader(tc.input), "t.sentrie").NextToken()
			if tok.Kind != tokens.String || tok.Value != tc.want {
				t.Fatalf("input %q: got %s %q, want String %q", tc.input, tok.Kind, tok.Value, tc.want)
			}
		})
	}

	// The token's span still covers the source text, including the \r bytes.
	tok := NewLexer(strings.NewReader("\"a\r\nb\" x"), "t.sentrie").NextToken()
	if tok.Range.From.Line != 1 || tok.Range.To.Line != 2 || tok.Range.To.Column != 3 || tok.Range.To.Offset != 6 {
		t.Fatalf("CRLF string span: got %+v", tok.Range)
	}
}
