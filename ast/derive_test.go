// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func TestDeriveStatementString(t *testing.T) {
	rng := tokens.Range{File: "t.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 2, Column: 1, Offset: 10}}
	lam := NewLambdaExpression([]string{"x"}, NewBlockExpression(nil, NewIntegerLiteral(1, rng), rng), rng)
	ds := NewDeriveStatement("d", lam, rng)
	require.Contains(t, ds.String(), "derive d")
	require.Equal(t, rng, ds.Span())

	ed := NewExportDeriveStatement("d", rng)
	require.Equal(t, "export derive d", ed.String())
}
