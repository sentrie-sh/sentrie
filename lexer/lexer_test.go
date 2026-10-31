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

func TestLexerNumberExponent(t *testing.T) {
	cases := []struct {
		input string
		kind  tokens.Kind
		value string
	}{
		{"42", tokens.Int, "42"},
		{"3.14", tokens.Float, "3.14"},
		{"1e5", tokens.Float, "1e5"},
		{"1E5", tokens.Float, "1E5"},
		{"1e+5", tokens.Float, "1e+5"},
		{"1e-5", tokens.Float, "1e-5"},
		{"2.5e-3", tokens.Float, "2.5e-3"},
		{"2.5E+10", tokens.Float, "2.5E+10"},
		{"0e0", tokens.Float, "0e0"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			l := NewLexer(strings.NewReader(tc.input), "test.sent")
			tok := l.NextToken()
			if tok.Kind != tc.kind || tok.Value != tc.value {
				t.Fatalf("got %s %q, want %s %q", tok.Kind, tok.Value, tc.kind, tc.value)
			}
			if tok.Range.To.Column != len(tc.input)+1 {
				t.Fatalf("token ends at column %d, want %d", tok.Range.To.Column, len(tc.input)+1)
			}
			if next := l.NextToken(); next.Kind != tokens.EOF {
				t.Fatalf("expected EOF after %q, got %s %q", tc.input, next.Kind, next.Value)
			}
		})
	}
}

func TestLexerNumberExponentInExpression(t *testing.T) {
	got := collectKinds("1e5+2.5e-3*x")
	want := []tokens.Kind{tokens.Float, tokens.TokenPlus, tokens.Float, tokens.TokenMul, tokens.Ident, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %s, want %s (all: %v)", i, got[i], want[i], got)
		}
	}
}

func TestLexerRejectsMalformedExponent(t *testing.T) {
	cases := []struct {
		input   string
		literal string
		message string
	}{
		{"1e", `"1e"`, "the exponent needs at least one digit"},
		{"1E", `"1E"`, "the exponent needs at least one digit"},
		{"1e+", `"1e+"`, "the exponent needs at least one digit"},
		{"1e-", `"1e-"`, "the exponent needs at least one digit"},
		{"1ex", `"1e"`, "the exponent needs at least one digit"},
		{"2.5e", `"2.5e"`, "the exponent needs at least one digit"},
		{"1e+ 5", `"1e+"`, "the exponent needs at least one digit"},
		{"1.e5", `"1.e5"`, "a digit must follow the decimal point"},
		{"1.E-5", `"1.E-5"`, "a digit must follow the decimal point"},
		{"1.e", `"1.e"`, "a digit must follow the decimal point"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			l := NewLexer(strings.NewReader(tc.input), "test.sent")
			tok := l.NextToken()
			if tok.Kind != tokens.Error {
				t.Fatalf("expected error token, got %s %q", tok.Kind, tok.Value)
			}
			if !strings.Contains(tok.Value, "invalid number literal "+tc.literal) || !strings.Contains(tok.Value, tc.message) {
				t.Fatalf("unexpected message %q", tok.Value)
			}
			if tok.Range.To.Column != len(strings.Trim(tc.literal, `"`))+1 {
				t.Fatalf("error token ends at column %d, want it to cover %s", tok.Range.To.Column, tc.literal)
			}
		})
	}
}

// TestLexerExponentMarkerOutsideNumbers tests that the exponent rule applies
// only straight after digits: `x.e5` is a field access, and `1.foo` is still an
// integer, a dot and an identifier.
func TestLexerExponentMarkerOutsideNumbers(t *testing.T) {
	cases := []struct {
		input string
		want  []tokens.Kind
	}{
		{"x.e5", []tokens.Kind{tokens.Ident, tokens.TokenDot, tokens.Ident, tokens.EOF}},
		{"e5", []tokens.Kind{tokens.Ident, tokens.EOF}},
		{"1.foo", []tokens.Kind{tokens.Int, tokens.TokenDot, tokens.Ident, tokens.EOF}},
		{"1 e5", []tokens.Kind{tokens.Int, tokens.Ident, tokens.EOF}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := collectKinds(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("token %d: got %s, want %s (all: %v)", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestLexerUnknownCharacterNamesTheCharacter(t *testing.T) {
	l := NewLexer(strings.NewReader("$"), "test.sent")
	tok := l.NextToken()
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s %q", tok.Kind, tok.Value)
	}
	if tok.Value != `unexpected character "$"` {
		t.Fatalf("unexpected message %q", tok.Value)
	}
	if tok.Range.From.Column != 1 || tok.Range.To.Column != 2 {
		t.Fatalf("error token spans columns %d-%d, want 1-2", tok.Range.From.Column, tok.Range.To.Column)
	}
}
