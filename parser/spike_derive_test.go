// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/stretchr/testify/require"
)

// Qualified slash callee vs division: `a / b` is division; `com/example/ns/name(x)` is a call
// whose callee is a slash chain (FQN), not `(a/b)*c`.
func TestSlashCalleeVsDivisionSpike(t *testing.T) {
	ctx := t.Context()

	div := NewParserFromString(`namespace n; policy p { export decision of r; rule r = { yield a / b } }`, "t.sentra")
	prog, err := div.ParseProgram(ctx)
	require.NoError(t, err)
	pol := prog.Statements[1].(*ast.PolicyStatement)
	rule := pol.Statements[1].(*ast.RuleStatement)
	body := rule.Body.(*ast.BlockExpression)
	infix, ok := body.Yield.(*ast.InfixExpression)
	require.True(t, ok, "a / b should be infix division")
	require.Equal(t, "/", infix.Operator)

	call := NewParserFromString(`namespace n; policy p { export decision of r; rule r = { yield com/example/ns/name(x) } }`, "t2.sentra")
	prog2, err := call.ParseProgram(ctx)
	require.NoError(t, err)
	pol2 := prog2.Statements[1].(*ast.PolicyStatement)
	rule2 := pol2.Statements[1].(*ast.RuleStatement)
	body2 := rule2.Body.(*ast.BlockExpression)
	ce, ok := body2.Yield.(*ast.CallExpression)
	require.True(t, ok, "FQN(x) should be call")
	fqn := ast.SlashCalleeFQNS(ce.Callee)
	require.Equal(t, "com/example/ns/name", fqn)
}

func TestTypedLambdaParamSpike(t *testing.T) {
	ctx := t.Context()
	p := NewParserFromString(`namespace n; policy p { export decision of r; rule r = { yield (a: number, b?: string) => { yield 1 } } }`, "tl.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)
	pol := prog.Statements[1].(*ast.PolicyStatement)
	rule := pol.Statements[1].(*ast.RuleStatement)
	body := rule.Body.(*ast.BlockExpression)
	lam, ok := body.Yield.(*ast.LambdaExpression)
	require.True(t, ok)
	require.Equal(t, []string{"a", "b"}, lam.Params)
	require.Len(t, lam.ParamOpts, 2)
	require.False(t, lam.ParamOpts[0])
	require.True(t, lam.ParamOpts[1])
}
