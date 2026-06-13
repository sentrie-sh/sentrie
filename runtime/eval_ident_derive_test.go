// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

func stubDeriveLambda(params []string, opts []bool, ret ast.TypeRef, yield ast.Expression) *ast.LambdaExpression {
	types := make([]ast.TypeRef, len(params))
	for i := range types {
		types[i] = ast.NewNumberTypeRef(stubRange())
	}
	return ast.NewLambdaExpressionFull(params, types, opts, ret, ast.NewBlockExpression(nil, yield, stubRange()), stubRange())
}

func stubTrinaryDeriveLambda(params []string, opts []bool, yield ast.Expression) *ast.LambdaExpression {
	return stubDeriveLambda(params, opts, ast.NewTrinaryTypeRef(stubRange()), yield)
}

func stubNumberDeriveLambda(params []string, opts []bool, yield ast.Expression) *ast.LambdaExpression {
	return stubDeriveLambda(params, opts, ast.NewNumberTypeRef(stubRange()), yield)
}

func attachNamespaceDerive(p *index.Policy, name string, lam *ast.LambdaExpression) *index.Derive {
	if p.Namespace.Derives == nil {
		p.Namespace.Derives = map[string]*index.Derive{}
	}
	d := &index.Derive{
		Name:      name,
		FQN:       ast.CreateFQN(p.Namespace.FQN, name),
		Lambda:    lam,
		Namespace: p.Namespace,
	}
	p.Namespace.Derives[name] = d
	return d
}

func attachPolicyDerive(p *index.Policy, name string, lam *ast.LambdaExpression) *index.Derive {
	if p.Derives == nil {
		p.Derives = map[string]*index.Derive{}
	}
	d := &index.Derive{
		Name:      name,
		FQN:       ast.CreateFQN(p.FQN, name),
		Lambda:    lam,
		Namespace: p.Namespace,
		Policy:    p,
	}
	p.Derives[name] = d
	return d
}

func (s *RuntimeTestSuite) requireDeriveCallable(val box.Value, want *index.Derive, wantArity int) Callable {
	s.T().Helper()
	s.Require().True(val.IsCallable())
	c, err := callableFromValue(val)
	s.Require().NoError(err)
	dc, ok := c.(*deriveCallable)
	s.Require().True(ok, "expected *deriveCallable, got %T", c)
	s.Require().Same(want, dc.derive)
	s.Equal(wantArity, c.Arity())
	return c
}

func (s *RuntimeTestSuite) TestEvalIdentDeriveCallableResolution() {
	ctx := s.T().Context()
	exec := &executorImpl{}

	s.Run("namespace derive becomes callable", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		isOne := attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewInfixExpression(ast.NewIdentifier("n", stubRange()), ast.NewIntegerLiteral(1, stubRange()), "==", stubRange()),
		))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
		s.Require().NoError(err)
		s.requireDeriveCallable(val, isOne, 1)
	})

	s.Run("policy-scoped derive becomes callable", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		policyPred := attachPolicyDerive(p, "policyPred", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewInfixExpression(ast.NewIdentifier("n", stubRange()), ast.NewIntegerLiteral(2, stubRange()), "==", stubRange()),
		))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("policyPred", stubRange()))
		s.Require().NoError(err)
		s.requireDeriveCallable(val, policyPred, 1)
	})

	s.Run("arity counts required params only", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		optionalPred := attachNamespaceDerive(p, "optionalPred", stubTrinaryDeriveLambda(
			[]string{"a", "b"},
			[]bool{true, false},
			ast.NewTrinaryLiteral(trinary.True, stubRange()),
		))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("optionalPred", stubRange()))
		s.Require().NoError(err)
		s.requireDeriveCallable(val, optionalPred, 1)
	})

	s.Run("two required params expose arity two", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		twoParamPred := attachNamespaceDerive(p, "twoParamPred", stubTrinaryDeriveLambda(
			[]string{"item", "idx"},
			nil,
			ast.NewInfixExpression(
				ast.NewIdentifier("idx", stubRange()),
				ast.NewIntegerLiteral(1, stubRange()),
				"==",
				stubRange(),
			),
		))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("twoParamPred", stubRange()))
		s.Require().NoError(err)
		s.requireDeriveCallable(val, twoParamPred, 2)
	})

	s.Run("local binding shadows derive name", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewTrinaryLiteral(trinary.True, stubRange()),
		))
		ec.SetLocal("isOne", box.Number(42), true)
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
		s.Require().NoError(err)
		s.Equal(42.0, val.Any())
		s.False(val.IsCallable())
	})

	s.Run("let binding shadows derive name", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewTrinaryLiteral(trinary.True, stubRange()),
		))
		s.Require().NoError(ec.InjectLet("isOne", ast.NewVarDeclaration(
			"isOne",
			nil,
			ast.NewIntegerLiteral(7, stubRange()),
			stubRange(),
		)))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
		s.Require().NoError(err)
		s.Equal(7.0, val.Any())
		s.False(val.IsCallable())
	})

	s.Run("fact binding shadows derive name", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewTrinaryLiteral(trinary.True, stubRange()),
		))
		s.Require().NoError(ec.InjectFact(ctx, "isOne", box.String("fact-wins"), false, nil))
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
		s.Require().NoError(err)
		s.Equal("fact-wins", val.Any())
		s.False(val.IsCallable())
	})

	s.Run("unknown identifier errors", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		_, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("missingPred", stubRange()))
		s.Require().ErrorContains(err, "identifier not found: missingPred")
	})

	s.Run("rule name dispatches outside derive context", func() {
		idx, execWithIndex := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
			name: "rule_dispatch.sentrie",
			src: `namespace test/ns
policy pol {
  let _s = 0
  rule gate = { yield true }
  export decision of gate
}`,
		})
		pol := idx.Namespaces["test/ns"].Policies["pol"]
		ec := NewExecutionContext(pol, execWithIndex)
		val, _, err := evalIdent(ctx, ec, execWithIndex, pol, ast.NewIdentifier("gate", stubRange()))
		s.Require().NoError(err)
		s.False(val.IsCallable())
	})

	s.Run("rule name does not dispatch inside derive context", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
			[]string{},
			nil,
			ast.NewIntegerLiteral(1, stubRange()),
		))
		p.Rules["gate"] = &index.Rule{
			Name: "gate",
			Body: ast.NewBlockExpression(nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()),
		}
		ec.evalDerive = caller
		_, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("gate", stubRange()))
		s.Require().ErrorContains(err, "identifier not found: gate")
	})

	s.Run("fact identifier forbidden inside derive context", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
			[]string{},
			nil,
			ast.NewIntegerLiteral(1, stubRange()),
		))
		p.Facts["user"] = &ast.FactStatement{}
		ec.evalDerive = caller
		_, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("user", stubRange()))
		s.Require().ErrorContains(err, "facts are not available inside a derive")
	})

	s.Run("define-site snapshot resolves helper inside derive context", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		helper := attachNamespaceDerive(p, "helper", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewInfixExpression(ast.NewIdentifier("n", stubRange()), ast.NewIntegerLiteral(1, stubRange()), "==", stubRange()),
		))
		caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
			[]string{},
			nil,
			ast.NewIntegerLiteral(1, stubRange()),
		))
		caller.DefineShort = map[string]*index.Derive{"helper": helper}
		ec.evalDerive = caller
		val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("helper", stubRange()))
		s.Require().NoError(err)
		s.requireDeriveCallable(val, helper, 1)
	})

	s.Run("define-site snapshot hides namespace derive not in snapshot", func() {
		p := newEvalTestPolicy()
		ec := NewExecutionContext(p, exec)
		attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
			[]string{"n"},
			nil,
			ast.NewTrinaryLiteral(trinary.True, stubRange()),
		))
		caller := attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
			[]string{},
			nil,
			ast.NewIntegerLiteral(1, stubRange()),
		))
		caller.DefineShort = map[string]*index.Derive{}
		ec.evalDerive = caller
		_, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
		s.Require().ErrorContains(err, "identifier not found: isOne")
	})
}

func (s *RuntimeTestSuite) TestEvalIdentDeriveCallableInvokeViaHOF() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	isOne := attachNamespaceDerive(p, "isOne", stubTrinaryDeriveLambda(
		[]string{"n"},
		nil,
		ast.NewInfixExpression(ast.NewIdentifier("n", stubRange()), ast.NewIntegerLiteral(1, stubRange()), "==", stubRange()),
	))
	val, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("isOne", stubRange()))
	s.Require().NoError(err)
	c := s.requireDeriveCallable(val, isOne, 1)

	site := &CallSite{EC: ec, Exec: exec, Policy: p}
	out, err := invokeTestBuiltin(ctx, site, "filter", box.List([]box.Value{box.Number(1), box.Number(2)}), box.Callable(c))
	s.Require().NoError(err)
	list, ok := out.ListValue()
	s.Require().True(ok)
	s.Len(list, 1)
	s.Equal(1.0, list[0].Any())
}

func (s *RuntimeTestSuite) TestExecRuleDeriveIdentifierCallbacksExhaustive() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive isOne = (n: number): trinary => { yield n == 1 }
derive isIdxOne = (item: number, idx: number): trinary => { yield idx == 1 }
derive double = (n: number): number => { yield n * 2 }
derive alwaysTrue = (n: number): trinary => { yield true }
derive keyNum = (n: number): number => { yield n }
derive zeroArity = (): trinary => { yield true }

policy pol {
  let _seed = 0
  rule any_cb = { yield any([1, 2, 3], isOne) }
  rule all_cb = { yield all([1, 2], alwaysTrue) }
  rule first_cb = { yield first([2, 1, 3], isOne) == 1 }
  rule filter_cb = { yield count(filter([1, 2, 3], isOne)) == 1 }
  rule collect_cb = { yield collect([1, 2], double) == [2, 4] }
  rule distinct_cb = { yield count(distinct([1, 1, 2, 2], keyNum)) == 2 }
  rule idx_cb = { yield count(filter([9, 8, 7], isIdxOne)) == 1 }
  rule zero_arity_cb = { yield any([1], zeroArity) }
  export decision of any_cb
  export decision of all_cb
  export decision of first_cb
  export decision of filter_cb
  export decision of collect_cb
  export decision of distinct_cb
  export decision of idx_cb
  export decision of zero_arity_cb
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "derive_cb.sentrie", src: src})

	successRules := []string{
		"any_cb", "all_cb", "first_cb", "filter_cb", "collect_cb", "distinct_cb", "idx_cb",
	}
	for _, ruleName := range successRules {
		ruleName := ruleName
		s.Run(ruleName, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", ruleName, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	s.Run("zero_arity_cb", func() {
		_, err := exec.ExecRule(ctx, "com/ex", "pol", "zero_arity_cb", nil)
		s.Require().Error(err)
		s.Contains(err.Error(), "any: callable must have arity 1 or 2")
	})
}

func (s *RuntimeTestSuite) TestExecRuleDeriveCallbackArgTypeUsesParamName() {
	ctx := s.T().Context()
	src := `namespace com/ex
derive needStr = (a: string): string => { yield a }
policy pol {
  let _seed = 0
  rule arg_type = { yield needStr(1) == "1" }
  export decision of arg_type
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "derive_arg_name.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "com/ex", "pol", "arg_type", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), `derive argument "a"`)
}
