// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestLambdaCallableArityUsesRequiredParamCount() {
	lam := ast.NewLambdaExpressionFull(
		[]string{"a", "b"},
		[]ast.TypeRef{ast.NewNumberTypeRef(stubRange()), ast.NewNumberTypeRef(stubRange())},
		[]bool{false, true},
		ast.NewTrinaryTypeRef(stubRange()),
		ast.NewBlockExpression(nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()),
		stubRange(),
	)
	c := newLambdaCallable(lam, NewExecutionContext(newEvalTestPolicy(), &executorImpl{}))
	s.Equal(1, c.Arity())
}

func (s *RuntimeTestSuite) TestExecRuleTypedLambdaCallbackRejectsWrongArgType() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  let _seed = 0
  rule bad = { yield any([1], (x: string): trinary => { yield true }) }
  export decision of bad
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda_type.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "com/ex", "pol", "bad", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), `lambda argument "x"`)
}

func (s *RuntimeTestSuite) TestExecRuleTypedLambdaOptionalParamOmittedPasses() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  let _seed = 0
  rule ok = { yield any([2], (x: number, min?: number): trinary => { yield true }) }
  export decision of ok
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda_opt.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestExecRuleTypedLambdaReturnTypeMismatch() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  let _seed = 0
  rule bad = { yield count(collect([1], (x: number): string => { yield x })) == 1 }
  export decision of bad
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda_ret.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "com/ex", "pol", "bad", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "lambda return")
}

func (s *RuntimeTestSuite) TestExecRuleDeriveOptionalTypedParamOmittedPasses() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive withMin = (x: number, min?: number): trinary => { yield true }
policy pol {
  let _seed = 0
  rule ok = { yield withMin(2) }
  export decision of ok
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "derive_opt.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestLambdaCallableDirectInvokeValidatesTypedArgs() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}
	site := &CallSite{EC: ec, Exec: exec, Policy: p}

	lam := ast.NewLambdaExpressionFull(
		[]string{"x"},
		[]ast.TypeRef{ast.NewStringTypeRef(stubRange())},
		nil,
		ast.NewTrinaryTypeRef(stubRange()),
		ast.NewBlockExpression(nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()),
		stubRange(),
	)
	c := newLambdaCallable(lam, ec)

	_, err := c.Invoke(ctx, site, []box.Value{box.Number(1)})
	s.Require().Error(err)
	s.Contains(err.Error(), `lambda argument "x"`)
}
