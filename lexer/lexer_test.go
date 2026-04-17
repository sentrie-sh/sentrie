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
