// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

type deriveTestProgram struct {
	name string
	src  string
}

func (s *RuntimeTestSuite) mustBuildDeriveIndex(ctx context.Context, programs ...deriveTestProgram) *index.Index {
	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))

	for _, program := range programs {
		prog, err := parser.NewParserFromString(program.src, program.name).ParseProgram(ctx)
		s.Require().NoError(err)
		s.Require().NoError(idx.AddProgram(ctx, prog))
	}

	return idx
}

func (s *RuntimeTestSuite) mustBuildDeriveExecutor(ctx context.Context, programs ...deriveTestProgram) (*index.Index, *executorImpl) {
	idx := s.mustBuildDeriveIndex(ctx, programs...)
	s.Require().NoError(idx.Validate(ctx))

	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	exec, ok := rawExec.(*executorImpl)
	s.Require().True(ok)
	return idx, exec
}

func (s *RuntimeTestSuite) TestIsBuiltinAllowedInDerive() {
	s.True(isBuiltinAllowedInDerive("now"))
	s.False(isBuiltinAllowedInDerive("not_a_builtin"))
}

func (s *RuntimeTestSuite) TestExecRuleDeriveIntegrationCases() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive bump = (n: number): number => { yield n + 1 }
derive id = (n: number): number => { yield n }
derive bad = (n: number): string => { yield n }
derive sum2 = (a: number, b: number): number => { yield a + b }
derive needStr = (a: string): string => { yield a }
derive ns = () => { yield 1 }
derive isOne = (a: number): trinary => { yield a == 1 }

policy pol {
  let _seed = 0
  rule bump_ok = {
    yield bump(1) == 2
  }
  rule too_many = { yield id(1, 2) == 1 }
  rule return_mismatch = { yield bad(1) == "1" }
  rule too_few = { yield sum2(1) == 1 }
  rule arg_type = { yield needStr(1) == "1" }
  rule slash_ok = { yield com/ex/ns() == 1 }
  rule derive_callback_any = { yield any([1, 2], isOne) }
  rule derive_callback_filter = { yield count(filter([1, 2], isOne)) == 1 }
  rule elvis_short = { yield _seed ?: (1 / 0) == 0 }
  export decision of bump_ok
  export decision of too_many
  export decision of return_mismatch
  export decision of too_few
  export decision of arg_type
  export decision of slash_ok
  export decision of derive_callback_any
  export decision of derive_callback_filter
  export decision of elvis_short
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "derive_exec.sentrie", src: src})

	successRules := []string{"bump_ok", "slash_ok", "derive_callback_any", "derive_callback_filter", "elvis_short"}
	for _, ruleName := range successRules {
		s.Run(ruleName, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", ruleName, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	errorCases := []struct {
		rule string
		msg  string
	}{
		{rule: "too_many", msg: "too many arguments"},
		{rule: "return_mismatch", msg: "derive return"},
		{rule: "too_few", msg: "not enough arguments"},
		{rule: "arg_type", msg: `derive argument "a"`},
	}
	for _, tc := range errorCases {
		tc := tc
		s.Run(tc.rule, func() {
			_, err := exec.ExecRule(ctx, "com/ex", "pol", tc.rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
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
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "unk.sentrie", src: src})
	err := idx.Validate(ctx)
	s.Error(err)
	s.Contains(err.Error(), "unknown derive")
}

func (s *RuntimeTestSuite) TestDeriveExportAndTargetRestrictions() {
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
derive d = () => { yield 1 }
policy pol {
  let _seed = 0
  rule gate = { yield com/alpha/secret() == 1 }
  export decision of gate
}
`
	idx, exec := s.mustBuildDeriveExecutor(
		ctx,
		deriveTestProgram{name: "alpha.sentrie", src: srcAlpha},
		deriveTestProgram{name: "ex.sentrie", src: srcEx},
	)

	pol := idx.Namespaces["com/ex"].Policies["pol"]
	bridge := idx.DerivesByFQN["com/ex/bridge"]
	s.Require().NotNil(bridge)
	ec := NewExecutionContext(pol, exec)
	ec.evalDerive = bridge

	rng := stubRange()
	com := ast.NewIdentifier("com", rng)
	al := ast.NewIdentifier("alpha", rng)
	seg := ast.NewIdentifier("secret", rng)
	slash1 := ast.NewInfixExpression(com, al, "/", rng)
	callee := ast.NewInfixExpression(slash1, seg, "/", rng)
	call := ast.NewCallExpression(callee, nil, false, nil, rng)

	var err error
	_, err = getTarget(ctx, ec, exec, pol, call)
	s.Error(err)
	var ne xerr.NotExportedError
	s.True(errors.As(err, &ne), "expected not-exported error, got %v", err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "gate", nil)
	s.Error(err)
	ne = xerr.NotExportedError{}
	s.True(errors.As(err, &ne), "expected not-exported error, got %v", err)

	d := idx.DerivesByFQN["com/ex/d"]
	s.Require().NotNil(d)
	ec = NewExecutionContext(pol, exec)
	ec.evalDerive = d

	fieldCallee := ast.NewFieldAccessExpression(ast.NewIdentifier("mod", rng), "fn", rng)
	fieldCall := ast.NewCallExpression(fieldCallee, nil, false, nil, rng)

	_, err = getTarget(ctx, ec, exec, pol, fieldCall)
	s.Error(err)
	s.Contains(err.Error(), "TypeScript module calls")
}

func (s *RuntimeTestSuite) TestExecRuleRejectsCrossPolicyDeriveSlashFQN() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy polA {
  let _s = 0
  derive secret = () => { yield 1 }
  rule x = { yield true }
  export decision of x
}
policy polB {
  let _s = 0
  rule gate = { yield com/ex/polA/secret() == 1 }
  export decision of gate
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "crosspol.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "com/ex", "polB", "gate", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "not visible from policy")
}

func (s *RuntimeTestSuite) TestExecRuleAllowsSamePolicyDeriveSlashFQN() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy polA {
  let _s = 0
  derive secret = () => { yield 1 }
  rule ok_gate = { yield com/ex/polA/secret() == 1 }
  export decision of ok_gate
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "samepol.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "polA", "ok_gate", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}
