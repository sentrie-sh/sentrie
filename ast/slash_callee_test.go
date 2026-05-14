// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func stubRng() tokens.Range {
	return tokens.Range{File: "t.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 1}}
}

func TestFlattenSlashIdentChain(t *testing.T) {
	a := NewIdentifier("a", stubRng())
	b := NewIdentifier("b", stubRng())
	c := NewIdentifier("c", stubRng())

	parts, ok := FlattenSlashIdentChain(a)
	require.True(t, ok)
	require.Equal(t, []string{"a"}, parts)

	chain := NewInfixExpression(NewInfixExpression(a, b, "/", stubRng()), c, "/", stubRng())
	parts, ok = FlattenSlashIdentChain(chain)
	require.True(t, ok)
	require.Equal(t, []string{"a", "b", "c"}, parts)

	_, ok = FlattenSlashIdentChain(NewInfixExpression(a, b, "+", stubRng()))
	require.False(t, ok)

	_, ok = FlattenSlashIdentChain(NewIntegerLiteral(1, stubRng()))
	require.False(t, ok)
}

func TestSlashCalleeFQNS(t *testing.T) {
	chain := NewInfixExpression(
		NewInfixExpression(NewIdentifier("com", stubRng()), NewIdentifier("ex", stubRng()), "/", stubRng()),
		NewIdentifier("fn", stubRng()),
		"/",
		stubRng(),
	)
	require.Equal(t, "com/ex/fn", SlashCalleeFQNS(chain))
	require.Equal(t, "solo", SlashCalleeFQNS(NewIdentifier("solo", stubRng())))
	require.Equal(t, "", SlashCalleeFQNS(NewIntegerLiteral(1, stubRng())))
}
