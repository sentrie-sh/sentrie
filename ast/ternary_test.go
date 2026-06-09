// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func rng() tokens.Range {
	return tokens.Range{File: "t.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 2, Offset: 1}}
}

func TestNewTernaryElvisString(t *testing.T) {
	lhs := NewIntegerLiteral(1, rng())
	rhs := NewIntegerLiteral(2, rng())
	e := NewTernaryElvis(lhs, rhs, rng())
	require.True(t, e.Elvis)
	require.Same(t, e.Condition, e.ThenBranch)
	require.Contains(t, e.String(), "?:")
}

func TestNewTernaryExpressionString(t *testing.T) {
	c := NewTrinaryLiteral(trinary.True, rng())
	th := NewIntegerLiteral(1, rng())
	el := NewIntegerLiteral(0, rng())
	e := NewTernaryExpression(c, th, el, rng())
	require.False(t, e.Elvis)
	require.Contains(t, e.String(), "?")
	require.Contains(t, e.String(), ":")
}
