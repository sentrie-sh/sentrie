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

func TestDeriveFatBodyWalkPassesValidation(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive leaf = () => { yield 1 }
derive fat = () => {
  let v = leaf()
  yield v + (true ? 1 : 0) + (null ?: 2) + (-1) + [1, 2][0] + (1 as number) + (!false) + ({"a":1}["a"])
}
policy pol {
  let _s = 0
  rule allow = { yield fat() > 0 }
  export decision of allow
}
`
	prog, err := parser.NewParserFromString(src, "fat.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))
}

func TestResolveDeriveNotFound(t *testing.T) {
	idx := CreateIndex()
	_, err := idx.ResolveDerive("com/nope/never")
	require.Error(t, err)
}

func TestAddProgramExportDeriveAndVerifyExported(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive published = () => { yield 1 }
export derive published
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "exp.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	ns := idx.Namespaces["com/ex"]
	require.NoError(t, ns.VerifyDeriveExported("published"))
	require.Error(t, ns.VerifyDeriveExported("unpublished"))
}

func TestAlphaSecretNotExportedInIndex(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}
`
	prog, err := parser.NewParserFromString(src, "a.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))
	ns := idx.Namespaces["com/alpha"]
	require.Error(t, ns.VerifyDeriveExported("secret"))
}

func TestDerivePurityRejectsFieldAccessModuleCall(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield mod.fn(1) }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "mod.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TypeScript module calls are not permitted")
}

func TestDerivePurityRejectsFactIdentifier(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  fact user:string
  let _s = 0
  derive d = () => { yield user }
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "fact.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "facts are not available inside a derive")
}

func TestDerivePurityRejectsYieldLambda(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield (x: number): number => { yield x } }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "lam.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "derive cannot yield a lambda value")
}

func TestDerivePurityLetArithmeticPasses(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = (x: number): number => {
  let doubled = x * 2
  yield doubled + 1
}
policy pol {
  let _s = 0
  rule r = { yield d(1) == 3 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "let.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))
}

func TestDerivePurityAllowsLambdaInsidePureBuiltinCall(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield count(filter([1, 2], (a: number): trinary => { yield a == 1 })) }
policy pol {
  let _s = 0
  rule r = { yield d() == 1 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "filt.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))
}

func TestDeriveDefineSiteOrderHelperAfterCallerFails(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	srcCaller := `namespace com/ex
derive caller = () => { yield helper() }
policy pol {
  let _s = 0
  rule r = { yield caller() == 1 }
  export decision of r
}
`
	srcHelper := `namespace com/ex
derive helper = () => { yield 1 }
policy pol2 {
  let _s = 0
  rule r2 = { yield true }
  export decision of r2
}
`
	p1, err := parser.NewParserFromString(srcCaller, "caller.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, p1))
	p2, err := parser.NewParserFromString(srcHelper, "helper.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, p2))
	err = idx.Validate(ctx)
	require.Error(t, err)
}

func TestDeriveDefineSiteOrderHelperFirstPasses(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	srcHelper := `namespace com/ex
derive helper = () => { yield 1 }
policy pol2 {
  let _s = 0
  rule r2 = { yield true }
  export decision of r2
}
`
	srcCaller := `namespace com/ex
derive caller = () => { yield helper() }
policy pol {
  let _s = 0
  rule r = { yield caller() == 1 }
  export decision of r
}
`
	p1, err := parser.NewParserFromString(srcHelper, "h.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, p1))
	p2, err := parser.NewParserFromString(srcCaller, "c.sentra").ParseProgram(ctx)
	require.NoError(t, err)
	require.NoError(t, idx.AddProgram(ctx, p2))
	require.NoError(t, idx.Validate(ctx))
}
