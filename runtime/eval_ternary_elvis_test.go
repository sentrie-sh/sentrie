// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func elvisTestRange() tokens.Range {
	return tokens.Range{File: "elvis_test.sentra", From: tokens.Pos{}, To: tokens.Pos{}}
}

func TestEvalElvisFiveCases(t *testing.T) {
	ctx := context.Background()
	p := &index.Policy{}
	ec := NewExecutionContext(p, &executorImpl{})
	defer ec.Dispose()
	exec := &executorImpl{}
	ec.SetLocal("undefLocal", box.Undefined(), true)

	rng := elvisTestRange()
	t.Run("null", func(t *testing.T) {
		el := ast.NewTernaryElvis(ast.NewNullLiteral(rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(t, err, v, 99)
	})
	t.Run("undefined", func(t *testing.T) {
		el := ast.NewTernaryElvis(ast.NewIdentifier("undefLocal", rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(t, err, v, 99)
	})
	t.Run("zero", func(t *testing.T) {
		el := ast.NewTernaryElvis(ast.NewIntegerLiteral(0, rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(t, err, v, 0)
	})
	t.Run("empty_string", func(t *testing.T) {
		el := ast.NewTernaryElvis(ast.NewStringLiteral("", rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		require.NoError(t, err)
		s, ok := v.StringValue()
		require.True(t, ok)
		require.Equal(t, "", s)
	})
	t.Run("false", func(t *testing.T) {
		el := ast.NewTernaryElvis(ast.NewTrinaryLiteral(trinary.False, rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		require.NoError(t, err)
		require.True(t, box.EqualValues(v, box.Trinary(trinary.False)))
	})
}

func TestExecRuleElvisParsedShortCircuitsDivideByZero(t *testing.T) {
	ctx := context.Background()
	src := `namespace n
policy p {
  let _seed = 0
  rule r = { yield _seed ?: (1 / 0) == 0 }
  export decision of r
}
`
	p := parser.NewParserFromString(src, "elvis_exec.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	out, err := exec.ExecRule(ctx, "n", "p", "r", nil)
	require.NoError(t, err)
	require.Equal(t, trinary.True, out.Decision.State)
}

func requireNoErrorNum(t *testing.T, err error, v box.Value, want float64) {
	t.Helper()
	require.NoError(t, err)
	n, ok := v.NumberValue()
	require.True(t, ok)
	require.Equal(t, want, n)
}
