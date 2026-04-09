// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package lexer

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

func TestLexerFatArrowTokenization(t *testing.T) {
	l := NewLexer(strings.NewReader("(x)=>{yield x}"), "test.sent")

	seen := []tokens.Kind{
		l.NextToken().Kind, // (
		l.NextToken().Kind, // ident
		l.NextToken().Kind, // )
		l.NextToken().Kind, // =>
	}

	if seen[3] != tokens.TokenFatArrow {
		t.Fatalf("expected fat arrow token, got %v", seen[3])
	}
}

func TestLexerMapIsKeywordAgain(t *testing.T) {
	l := NewLexer(strings.NewReader("map[list]"), "test.sent")
	first := l.NextToken()
	if first.Kind != tokens.KeywordMap {
		t.Fatalf("expected map keyword token, got %v (%q)", first.Kind, first.Value)
	}
}
