// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"strings"
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

func TestLambdaExpressionFullStringOptionalParamsAndReturn(t *testing.T) {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 30, Offset: 29},
	}
	body := NewBlockExpression(nil, NewIntegerLiteral(0, r), r)
	types := []TypeRef{NewNumberTypeRef(r), NewStringTypeRef(r)}
	opts := []bool{true, false}
	ret := NewNumberTypeRef(r)
	lam := NewLambdaExpressionFull([]string{"a", "b"}, types, opts, ret, body, r)
	s := lam.String()
	for _, sub := range []string{"a?", "b:", "number", "string", "=>", ": number =>"} {
		if !strings.Contains(s, sub) {
			t.Fatalf("lambda string %q missing %q", s, sub)
		}
	}
}
