// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestLookupDeriveByIdentifierEvalDeriveDefineShortHitAndMiss() {
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	helper := attachNamespaceDerive(p, "helper", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	caller.DefineShort = map[string]*index.Derive{"helper": helper}
	ec.evalDerive = caller

	s.Same(helper, lookupDeriveByIdentifier(ec, p, "helper"))
	s.Nil(lookupDeriveByIdentifier(ec, p, "absent"))
}

func (s *RuntimeTestSuite) TestLookupDeriveByIdentifierPolicyScopedDerive() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	secret := attachPolicyDerive(p, "secret", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	s.Same(secret, lookupDeriveByIdentifier(ec, p, "secret"))
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQDefineFQNAndUnknown() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "slash_lookup.sentrie",
		src: `namespace com/ex
derive helper = () => { yield 1 }
derive caller = () => { yield com/ex/helper() }
policy pol {
  let _s = 0
  rule ok = { yield caller() == 1 }
  export decision of ok
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	caller := idx.DerivesByFQN["com/ex/caller"]
	s.Require().NotNil(caller)
	ec := NewExecutionContext(pol, exec)
	ec.evalDerive = caller

	got, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/ex/helper")
	s.Require().NoError(err)
	s.Same(idx.DerivesByFQN["com/ex/helper"], got)

	_, err = lookupDeriveBySlashFQ(ec, exec, pol, "com/ex/missing")
	s.Require().Error(err)
	s.Contains(err.Error(), "unknown derive")
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQResolvePathOutsideEvalDerive() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "slash_resolve.sentrie",
		src: `namespace com/ex
derive helper = () => { yield 1 }
policy pol {
  let _s = 0
  rule ok = { yield com/ex/helper() == 1 }
  export decision of ok
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)

	got, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/ex/helper")
	s.Require().NoError(err)
	s.Same(idx.DerivesByFQN["com/ex/helper"], got)
}

func (s *RuntimeTestSuite) TestGetTargetDeriveIdentifierAndSlashCallee() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "get_target.sentrie",
		src: `namespace com/ex
derive bump = (n: number): number => { yield n + 1 }
policy pol {
  let _s = 0
  rule by_name = { yield bump(1) == 2 }
  rule by_slash = { yield com/ex/bump(1) == 2 }
  export decision of by_name
  export decision of by_slash
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)

	nameCall := ast.NewCallExpression(
		ast.NewIdentifier("bump", stubRange()),
		[]ast.Expression{ast.NewIntegerLiteral(1, stubRange())},
		false, nil, stubRange(),
	)
	target, err := getTarget(ctx, ec, exec, pol, nameCall)
	s.Require().NoError(err)
	out, err := target(ctx, box.Number(1))
	s.Require().NoError(err)
	s.Equal(2.0, out.Any())

	rng := stubRange()
	com := ast.NewIdentifier("com", rng)
	ex := ast.NewIdentifier("ex", rng)
	bump := ast.NewIdentifier("bump", rng)
	slash1 := ast.NewInfixExpression(com, ex, "/", rng)
	callee := ast.NewInfixExpression(slash1, bump, "/", rng)
	slashCall := ast.NewCallExpression(callee, []ast.Expression{ast.NewIntegerLiteral(1, rng)}, false, nil, rng)
	target, err = getTarget(ctx, ec, exec, pol, slashCall)
	s.Require().NoError(err)
	out, err = target(ctx, box.Number(1))
	s.Require().NoError(err)
	s.Equal(2.0, out.Any())
}

func (s *RuntimeTestSuite) TestGetTargetRejectsTypeScriptModuleInsideDerive() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	ec.evalDerive = caller

	fieldCall := ast.NewCallExpression(
		ast.NewFieldAccessExpression(ast.NewIdentifier("mod", stubRange()), "fn", stubRange()),
		nil, false, nil, stubRange(),
	)
	_, err := getTarget(ctx, ec, exec, p, fieldCall)
	s.Require().Error(err)
	s.Contains(err.Error(), "TypeScript module calls")
}

func (s *RuntimeTestSuite) TestExecRuleExportedCrossNamespaceDerive() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{
			name: "alpha_export.sentrie",
			src: `namespace com/alpha
derive published = () => { yield 1 }
export derive published
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}`,
		},
		deriveTestProgram{
			name: "beta_use.sentrie",
			src: `namespace com/beta
policy pol {
  let _s = 0
  rule gate = { yield com/alpha/published() == 1 }
  export decision of gate
}`,
		},
	)
	out, err := exec.ExecRule(ctx, "com/beta", "pol", "gate", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestCallerPolicyForDeriveScopeUsesEvalDerivePolicy() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	policyDerive := attachPolicyDerive(p, "inner", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	ec.evalDerive = policyDerive

	s.Same(p, callerPolicyForDeriveScope(ec, p))
	s.Equal(p.Namespace.FQN.String(), callerNamespaceFQNForDeriveExport(ec, p))
}

func (s *RuntimeTestSuite) TestLookupDeriveByIdentifierNilPolicyReturnsNil() {
	p := newEvalTestPolicy()
	attachNamespaceDerive(p, "open", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))
	s.Nil(lookupDeriveByIdentifier(NewExecutionContext(p, &executorImpl{}), nil, "open"))
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQTooFewSegmentsReturnsNil() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "two_seg.sentrie",
		src: `namespace com/ex
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)
	got, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/ex")
	s.Require().NoError(err)
	s.Nil(got)
}

func (s *RuntimeTestSuite) TestExecRuleDeriveUsesNowBuiltin() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive t = (): trinary => { yield now() >= 0 }
policy pol {
  let _seed = 0
  rule ok = { yield t() }
  export decision of ok
}`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "now_derive.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestEvalIdentRuleDispatchPropagatesExecError() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "rule_err.sentrie",
		src: `namespace com/ex
policy pol {
  let _s = 0
  rule bad = { yield 1 / 0 }
  rule caller = { yield bad }
  export decision of caller
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)
	_, _, err := evalIdent(ctx, ec, exec, pol, ast.NewIdentifier("bad", stubRange()))
	s.Require().Error(err)
}

func (s *RuntimeTestSuite) TestEnforceDeriveExportEdgeCases() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	d := attachNamespaceDerive(p, "pub", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))

	s.NoError(enforceDeriveExportForCaller(ec, p, nil))
	s.NoError(enforceDeriveExportForCaller(ec, &index.Policy{}, d))

	ec.evalDerive = d
	s.NoError(enforceDeriveExportForCaller(ec, p, d))
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQEnforcesExportFromEvalDerive() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{
			name: "alpha.secret.sentrie",
			src: `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}`,
		},
		deriveTestProgram{
			name: "beta.caller.sentrie",
			src: `namespace com/beta
derive caller = () => { yield com/alpha/secret() }
policy pol {
  let _s = 0
  rule r = { yield caller() == 1 }
  export decision of r
}`,
		},
	)
	caller := idx.DerivesByFQN["com/beta/caller"]
	pol := idx.Namespaces["com/beta"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)
	ec.evalDerive = caller

	_, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/alpha/secret")
	s.Require().Error(err)
}

func (s *RuntimeTestSuite) TestBuiltinNowRejectsExtraArgs() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	site := &CallSite{EC: ec, Exec: &executorImpl{}, Policy: p}
	_, err := BuiltinNow(ctx, site, box.Number(1))
	s.Require().Error(err)
	s.Contains(err.Error(), "now requires 0 arguments")
}
