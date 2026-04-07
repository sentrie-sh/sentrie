// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package lexer

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

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
