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

func TestLexerMapIsIdentifierAndDictIsKeyword(t *testing.T) {
	l := NewLexer(strings.NewReader("map[list]"), "test.sent")
	first := l.NextToken()
	if first.Kind != tokens.Ident {
		t.Fatalf("expected map identifier token, got %v (%q)", first.Kind, first.Value)
	}

	l2 := NewLexer(strings.NewReader("dict[list]"), "test.sent")
	first2 := l2.NextToken()
	if first2.Kind != tokens.KeywordDict {
		t.Fatalf("expected dict keyword token, got %v (%q)", first2.Kind, first2.Value)
	}
}
