// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package ast

import (
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

func TestLambdaExpressionStringAndKind(t *testing.T) {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 30, Offset: 29},
	}
	body := NewBlockExpression(nil, NewIdentifier("x", r), r)
	lam := NewLambdaExpression([]string{"x", "idx"}, body, r)

	if got := lam.Kind(); got != "lambda" {
		t.Fatalf("expected lambda kind, got %q", got)
	}
	if got := lam.String(); got != "(x, idx) => {yield x}" {
		t.Fatalf("unexpected lambda string: %q", got)
	}
}
