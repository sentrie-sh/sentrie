// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/stretchr/testify/require"
)

func TestParseElvisInRuleYield(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield null ?: 7 } }`
	prog, err := NewParserFromString(src, "elvis.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	pol := prog.Statements[1].(*ast.PolicyStatement)
	rule := findRuleStmt(t, pol.Statements, "r")
	body := rule.Body.(*ast.BlockExpression)
	tern, ok := body.Yield.(*ast.TernaryExpression)
	require.True(t, ok)
	require.True(t, tern.Elvis)
}

func TestParseClassicTernaryInRuleYield(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield true ? 1 : 2 } }`
	prog, err := NewParserFromString(src, "tern.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	pol := prog.Statements[1].(*ast.PolicyStatement)
	rule := findRuleStmt(t, pol.Statements, "r")
	body := rule.Body.(*ast.BlockExpression)
	tern, ok := body.Yield.(*ast.TernaryExpression)
	require.True(t, ok)
	require.False(t, tern.Elvis)
}

func findRuleStmt(t *testing.T, stmts []ast.Statement, name string) *ast.RuleStatement {
	t.Helper()
	for _, s := range stmts {
		if r, ok := s.(*ast.RuleStatement); ok && r.RuleName == name {
			return r
		}
	}
	t.Fatalf("rule %q not found", name)
	return nil
}
