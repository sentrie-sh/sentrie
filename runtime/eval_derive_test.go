// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *RuntimeTestSuite) TestIsBuiltinAllowedInDerive() {
	s.True(isBuiltinAllowedInDerive("now"))
	s.False(isBuiltinAllowedInDerive("not_a_builtin"))
}

func (s *RuntimeTestSuite) TestExecRuleInvokesNamespaceDerive() {
	ctx := s.T().Context()
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
	p := parser.NewParserFromString(src, "drv.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestExecRuleDeriveTooManyArgsErrors() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive id = (n: number): number => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield id(1, 2) == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "args.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Error(err)
	s.Contains(err.Error(), "too many arguments")
}

func (s *RuntimeTestSuite) TestExecRuleDeriveReturnTypeMismatchErrors() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive bad = (n: number): string => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield bad(1) == "1" }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "ret.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Error(err)
	s.Contains(err.Error(), "derive return")
}

func (s *RuntimeTestSuite) TestExecRuleUnknownSlashDeriveErrors() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive x = () => { yield com/ex/missing() }
policy pol {
  let _seed = 0
  rule allow = { yield x() == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "unk.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	s.Error(err)
	s.Contains(err.Error(), "unknown derive")
}

func (s *RuntimeTestSuite) TestExecRuleDeriveTooFewArgsErrors() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive sum2 = (a: number, b: number): number => { yield a + b }
policy pol {
  let _seed = 0
  rule allow = { yield sum2(1) == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "few.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Error(err)
	s.Contains(err.Error(), "not enough arguments")
}

func (s *RuntimeTestSuite) TestExecRuleDeriveArgTypeMismatchErrors() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive needStr = (a: string): string => { yield a }
policy pol {
  let _seed = 0
  rule allow = { yield needStr(1) == "1" }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "atype.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Error(err)
	s.Contains(err.Error(), "derive argument")
}

func (s *RuntimeTestSuite) TestExecRuleCallsNamespaceDeriveViaSlashFQN() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive ns = () => { yield 1 }
policy pol {
  let _seed = 0
  rule gate = { yield com/ex/ns() == 1 }
  export decision of gate
}
`
	p := parser.NewParserFromString(src, "slash.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "gate", nil)
	s.NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestGetTargetSlashCrossNamespaceDeriveRequiresExport() {
	ctx := s.T().Context()
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
	p1 := parser.NewParserFromString(srcAlpha, "alpha.sentrie")
	prog1, err := p1.ParseProgram(ctx)
	s.Require().NoError(err)
	p2 := parser.NewParserFromString(srcEx, "ex.sentrie")
	prog2, err := p2.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog1))
	s.Require().NoError(idx.AddProgram(ctx, prog2))
	s.Require().NoError(idx.Validate(ctx))

	pol := idx.Namespaces["com/ex"].Policies["pol"]
	exec := &executorImpl{index: idx}
	ec := NewExecutionContext(pol, exec)
	bridge := idx.DerivesByFQN["com/ex/bridge"]
	s.Require().NotNil(bridge)
	ec.evalDerive = bridge

	rng := stubRange()
	com := ast.NewIdentifier("com", rng)
	al := ast.NewIdentifier("alpha", rng)
	seg := ast.NewIdentifier("secret", rng)
	slash1 := ast.NewInfixExpression(com, al, "/", rng)
	callee := ast.NewInfixExpression(slash1, seg, "/", rng)
	call := ast.NewCallExpression(callee, nil, false, nil, rng)

	_, err = getTarget(ctx, ec, exec, pol, call)
	s.Error(err)
	var ne xerr.NotExportedError
	s.True(errors.As(err, &ne), "expected not-exported error, got %v", err)
}

func (s *RuntimeTestSuite) TestExecRuleSlashCrossNamespaceDeriveRequiresExport() {
	ctx := s.T().Context()
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
	p1 := parser.NewParserFromString(srcAlpha, "alpha.sentrie")
	prog1, err := p1.ParseProgram(ctx)
	s.Require().NoError(err)
	p2 := parser.NewParserFromString(srcEx, "ex.sentrie")
	prog2, err := p2.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog1))
	s.Require().NoError(idx.AddProgram(ctx, prog2))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "gate", nil)
	s.Error(err)
	var ne xerr.NotExportedError
	s.True(errors.As(err, &ne), "expected not-exported error, got %v", err)
}

func (s *RuntimeTestSuite) TestGetTargetFieldAccessCallBlockedInDerive() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive d = () => { yield 1 }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	p := parser.NewParserFromString(src, "d.sentrie")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	pol := idx.Namespaces["com/ex"].Policies["pol"]
	exec := &executorImpl{index: idx}
	ec := NewExecutionContext(pol, exec)
	d := idx.DerivesByFQN["com/ex/d"]
	s.Require().NotNil(d)
	ec.evalDerive = d

	rng := stubRange()
	callee := ast.NewFieldAccessExpression(ast.NewIdentifier("mod", rng), "fn", rng)
	call := ast.NewCallExpression(callee, nil, false, nil, rng)

	_, err = getTarget(ctx, ec, exec, pol, call)
	s.Error(err)
	s.Contains(err.Error(), "TypeScript module calls")
}
