// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/stretchr/testify/require"
)

func TestParseNamespaceDeriveAndExportDerive(t *testing.T) {
	ctx := t.Context()
	src := `namespace com/ex

derive dup = (n: number): number => { yield n + n }

export derive dup

policy pol {
  export decision of allow
  rule allow = { yield dup(2) == 4 }
}
`
	p := NewParserFromString(src, "d.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(prog.Statements), 4)
	_, ok := prog.Statements[1].(*ast.DeriveStatement)
	require.True(t, ok, "statement 1 should be derive")
	_, ok = prog.Statements[2].(*ast.ExportDeriveStatement)
	require.True(t, ok, "statement 2 should be export derive")
}

func TestParseDeriveRequiresLambda(t *testing.T) {
	ctx := t.Context()
	p := NewParserFromString(`namespace com/x
derive bad = 1
policy p { export decision of r; rule r = { yield true } }
`, "bad.sentra")
	_, err := p.ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseDeriveMissingAssignErrors(t *testing.T) {
	ctx := t.Context()
	_, err := NewParserFromString(`namespace com/x
derive bad (n: number): number => { yield n }
policy p { export decision of r; rule r = { yield true } }
`, "noassign.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseRuleElvisExpression(t *testing.T) {
	ctx := t.Context()
	_, err := NewParserFromString(`namespace com/x
policy p {
  let _s = 0
  export decision of r
  rule r = { yield _s ?: 1 == 1 }
}
`, "elvis.sentra").ParseProgram(ctx)
	require.NoError(t, err)
}

func TestParseRuleFullTernaryExpression(t *testing.T) {
	ctx := t.Context()
	_, err := NewParserFromString(`namespace com/x
policy p {
  let _s = 0
  export decision of r
  rule r = { yield true ? 1 : 0 == 0 }
}
`, "ternary.sentra").ParseProgram(ctx)
	require.NoError(t, err)
}

func TestParseTypedLambdaRequiredAfterOptionalIsRejected(t *testing.T) {
	ctx := t.Context()
	p := NewParserFromString(`namespace com/x
policy p {
  export decision of r
  rule r = { yield (a?: number, b: number) => { yield 1 } }
}
`, "opt.sentra")
	_, err := p.ParseProgram(ctx)
	require.Error(t, err)
}
