// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/require"
)

// Derive integration tests are standalone *testing.T functions (not suite methods).
// testify/suite.Run uses one shared suite value; do not call s.T().Parallel() there.

func TestIsBuiltinAllowedInDerive(t *testing.T) {
	require.True(t, isBuiltinAllowedInDerive("now"))
	require.False(t, isBuiltinAllowedInDerive("not_a_builtin"))
}

func TestExecRuleInvokesNamespaceDerive(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex

derive bump = (n: number): number => { yield n + 1 }

policy pol {
  let _seed = 0
  rule allow = {
    yield bump(1) == 2
  }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "drv.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	require.NoError(t, err)
	require.Equal(t, trinary.True, out.Decision.State)
}

func TestExecRuleDeriveTooManyArgsErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex

derive id = (n: number): number => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield id(1, 2) == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "args.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too many arguments")
}

func TestExecRuleDeriveReturnTypeMismatchErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex

derive bad = (n: number): string => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield bad(1) == "1" }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "ret.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "derive return")
}

func TestExecRuleUnknownSlashDeriveErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex
derive x = () => { yield com/ex/missing() }
policy pol {
  let _seed = 0
  rule allow = { yield x() == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "unk.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown derive")
}

func TestExecRuleDeriveTooFewArgsErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex
derive sum2 = (a: number, b: number): number => { yield a + b }
policy pol {
  let _seed = 0
  rule allow = { yield sum2(1) == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "few.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not enough arguments")
}

func TestExecRuleDeriveArgTypeMismatchErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex
derive needStr = (a: string): string => { yield a }
policy pol {
  let _seed = 0
  rule allow = { yield needStr(1) == "1" }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "atype.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "derive argument")
}

func TestExecRuleCallsNamespaceDeriveViaSlashFQN(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex
derive ns = () => { yield 1 }
policy pol {
  let _seed = 0
  rule gate = { yield com/ex/ns() == 1 }
  export decision of gate
}
`
	p := parser.NewParserFromString(src, "slash.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "gate", nil)
	require.NoError(t, err)
	require.Equal(t, trinary.True, out.Decision.State)
}

func TestGetTargetSlashCrossNamespaceDeriveRequiresExport(t *testing.T) {
	ctx := context.Background()
	srcAlpha := `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}
`
	srcEx := `namespace com/ex
derive bridge = () => { yield com/alpha/secret() }
policy pol {
  let _seed = 0
  rule gate = { yield bridge() == 0 }
  export decision of gate
}
`
	p1 := parser.NewParserFromString(srcAlpha, "alpha.sentra")
	prog1, err := p1.ParseProgram(ctx)
	require.NoError(t, err)
	p2 := parser.NewParserFromString(srcEx, "ex.sentra")
	prog2, err := p2.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog1))
	require.NoError(t, idx.AddProgram(ctx, prog2))
	require.NoError(t, idx.Validate(ctx))

	pol := idx.Namespaces["com/ex"].Policies["pol"]
	exec := &executorImpl{index: idx}
	ec := NewExecutionContext(pol, exec)
	bridge := idx.DerivesByFQN["com/ex/bridge"]
	require.NotNil(t, bridge)
	ec.evalDerive = bridge

	rng := stubRange()
	com := ast.NewIdentifier("com", rng)
	al := ast.NewIdentifier("alpha", rng)
	seg := ast.NewIdentifier("secret", rng)
	slash1 := ast.NewInfixExpression(com, al, "/", rng)
	callee := ast.NewInfixExpression(slash1, seg, "/", rng)
	call := ast.NewCallExpression(callee, nil, false, nil, rng)

	_, err = getTarget(ctx, ec, exec, pol, call)
	require.Error(t, err)
	var ne xerr.NotExportedError
	require.True(t, errors.As(err, &ne), "expected not-exported error, got %v", err)
}

func TestExecRuleSlashCrossNamespaceDeriveRequiresExport(t *testing.T) {
	ctx := context.Background()
	srcAlpha := `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}
`
	srcEx := `namespace com/ex
policy pol {
  let _seed = 0
  rule gate = { yield com/alpha/secret() == 1 }
  export decision of gate
}
`
	p1 := parser.NewParserFromString(srcAlpha, "alpha.sentra")
	prog1, err := p1.ParseProgram(ctx)
	require.NoError(t, err)
	p2 := parser.NewParserFromString(srcEx, "ex.sentra")
	prog2, err := p2.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog1))
	require.NoError(t, idx.AddProgram(ctx, prog2))
	require.NoError(t, idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	require.NoError(t, err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "gate", nil)
	require.Error(t, err)
	var ne xerr.NotExportedError
	require.True(t, errors.As(err, &ne), "expected not-exported error, got %v", err)
}

func TestGetTargetFieldAccessCallBlockedInDerive(t *testing.T) {
	ctx := context.Background()
	src := `namespace com/ex
derive d = () => { yield 1 }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	p := parser.NewParserFromString(src, "d.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(t, err)

	idx := index.CreateIndex()
	require.NoError(t, idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	require.NoError(t, idx.AddProgram(ctx, prog))
	require.NoError(t, idx.Validate(ctx))

	pol := idx.Namespaces["com/ex"].Policies["pol"]
	exec := &executorImpl{index: idx}
	ec := NewExecutionContext(pol, exec)
	d := idx.DerivesByFQN["com/ex/d"]
	require.NotNil(t, d)
	ec.evalDerive = d

	rng := stubRange()
	callee := ast.NewFieldAccessExpression(ast.NewIdentifier("mod", rng), "fn", rng)
	call := ast.NewCallExpression(callee, nil, false, nil, rng)

	_, err = getTarget(ctx, ec, exec, pol, call)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TypeScript module calls")
}
