// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/parser"
	"github.com/stretchr/testify/require"
)

func TestDeriveAddProgramResolveAndSpan(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex

derive helper = () => { yield 1 }

policy pol {
  let _seed = 0
  derive inner = () => { yield helper() }
  rule allow = { yield inner() == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "one.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	d, err := idx.ResolveDerive("com/ex/pol/inner")
	require.NoError(t, err)
	require.NotEmpty(t, d.String())
	_ = d.Span()
}

func TestDeriveCycleViaSlashFQNDetected(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex

derive a = () => { yield com/ex/b() }
derive b = () => { yield com/ex/a() }

policy pol {
  let _seed = 0
  rule allow = { yield true }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "cyc.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "cyclic derive") || strings.Contains(err.Error(), "derive dependency"))
}

func TestDeriveDuplicateFQNRejected(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src1 := `namespace com/ex
derive dup = () => { yield 1 }
policy p1 {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	src2 := `namespace com/ex
derive dup = () => { yield 2 }
policy p2 {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog1, err := parser.NewParserFromString(src1, "a.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog1))
	prog2, err := parser.NewParserFromString(src2, "b.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	err = idx.AddProgram(ctx, prog2)
	require.Error(t, err)
}

func TestExportDeriveUnknownNameRejected(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
export derive missing
derive x = () => { yield 1 }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "e.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	err = idx.AddProgram(ctx, prog)
	require.Error(t, err)
}
