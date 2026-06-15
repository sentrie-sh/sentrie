// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"math"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func stubRange() tokens.Range {
	return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}
}

func (s *RuntimeTestSuite) TestCalculateHashKeyDistinguishesUndefinedAndNull() {
	node := &ast.CallExpression{}
	undefinedHash := calculateHashKey(node, []box.Value{box.Undefined()})
	nullHash := calculateHashKey(node, []box.Value{box.Null()})

	s.Require().NotEmpty(undefinedHash)
	s.Require().NotEmpty(nullHash)
	s.Require().NotEqual(undefinedHash, nullHash)
}

func (s *RuntimeTestSuite) TestGetTargetBuiltinPreservesUndefined() {
	ec := NewExecutionContext(&index.Policy{}, &executorImpl{})
	call := ast.NewCallExpression(
		ast.NewIdentifier("as_list", stubRange()),
		[]ast.Expression{},
		false,
		nil,
		stubRange(),
	)

	target, err := getTarget(s.T().Context(), ec, &executorImpl{}, &index.Policy{}, call)
	s.Require().NoError(err)

	out, err := target(s.T().Context(), box.Undefined())
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestGetTargetBuiltinPreservesNestedUndefined() {
	ec := NewExecutionContext(&index.Policy{}, &executorImpl{})
	call := ast.NewCallExpression(
		ast.NewIdentifier("flatten_deep", stubRange()),
		[]ast.Expression{},
		false,
		nil,
		stubRange(),
	)

	target, err := getTarget(s.T().Context(), ec, &executorImpl{}, &index.Policy{}, call)
	s.Require().NoError(err)

	arg := box.List([]box.Value{
		box.List([]box.Value{
			box.Number(1),
			box.Undefined(),
		}),
	})
	out, err := target(s.T().Context(), arg)
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestCalculateHashKeyMapKeyOrderStable() {
	node := &ast.CallExpression{}
	arg1 := box.Dict(map[string]box.Value{"a": box.Number(1), "b": box.Number(2)})
	arg2 := box.Dict(map[string]box.Value{"b": box.Number(2), "a": box.Number(1)})
	hash1 := calculateHashKey(node, []box.Value{arg1})
	hash2 := calculateHashKey(node, []box.Value{arg2})
	s.Require().Equal(hash1, hash2)
}

func (s *RuntimeTestSuite) TestCalculateHashKeyNestedStructureStable() {
	node := &ast.CallExpression{}
	arg := box.List([]box.Value{
		box.Dict(map[string]box.Value{"k": box.List([]box.Value{box.Number(1), box.String("x")})}),
	})
	hash := calculateHashKey(node, []box.Value{arg})
	s.Require().NotEmpty(hash)
}

func (s *RuntimeTestSuite) TestCalculateHashKeyNumericEdges() {
	node := &ast.CallExpression{}
	hashNegZero := calculateHashKey(node, []box.Value{box.Number(math.Copysign(0, -1))})
	hashPosZero := calculateHashKey(node, []box.Value{box.Number(0)})
	hashNaN := calculateHashKey(node, []box.Value{box.Number(math.NaN())})
	hashInf := calculateHashKey(node, []box.Value{box.Number(math.Inf(1))})

	s.Require().NotEmpty(hashNaN)
	s.Require().NotEmpty(hashInf)
	s.Require().NotEqual(hashNegZero, hashPosZero)
}

func (s *RuntimeTestSuite) TestGetTargetDoesNotResolveImportedFunctionAsBareIdentifier() {
	p := newEvalTestPolicy()
	p.Uses = map[string]*ast.UseStatement{
		"string": ast.NewUseStatement(
			[]string{"trim"},
			"",
			[]string{"sentrie", "string"},
			"string",
			stubRange(),
		),
	}
	ec := NewExecutionContext(p, &executorImpl{})
	call := ast.NewCallExpression(
		ast.NewIdentifier("trim", stubRange()),
		[]ast.Expression{},
		false,
		nil,
		stubRange(),
	)

	_, err := getTarget(s.T().Context(), ec, &executorImpl{}, p, call)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unable to resolve import")
}

func (s *RuntimeTestSuite) TestPipelineHoleOutsidePipelineErrors() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	hole := ast.NewPipelineHoleExpression(stubRange())

	_, _, err := eval(s.T().Context(), ec, &executorImpl{}, p, hole)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "pipeline placeholder '#'")
}

func (s *RuntimeTestSuite) TestEvalCallMemoizedRejectsCallableArgument() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	lam := ast.NewLambdaExpression(
		[]string{"x"},
		ast.NewBlockExpression(nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()),
		stubRange(),
	)
	call := ast.NewCallExpression(
		ast.NewIdentifier("as_list", stubRange()),
		[]ast.Expression{lam},
		true,
		nil,
		stubRange(),
	)

	_, _, err := evalCall(ctx, ec, exec, p, call)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "memoized call cannot take callable arguments")
}

func (s *RuntimeTestSuite) TestGetTargetRejectsImpureBuiltinInsideDerive() {
	ctx := s.T().Context()
	const name = "test_impure_builtin"
	original, hadOriginal := builtins.Table[name]
	defer func() {
		if hadOriginal {
			builtins.Table[name] = original
			return
		}
		delete(builtins.Table, name)
	}()

	builtins.Table[name] = &builtins.Decl{
		Name:        name,
		Description: "test impure builtin",
		DeriveSafe:  false,
		Sig: builtins.Sig{
			TooFewError:  name + " requires 0 arguments",
			TooManyError: name + " requires 0 arguments",
		},
		Impl: func(_ context.Context, _ builtins.Env, _ ...box.Value) (box.Value, error) {
			return box.Undefined(), nil
		},
	}

	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	ec.evalDerive = attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(
		[]string{},
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
	))

	call := ast.NewCallExpression(ast.NewIdentifier(name, stubRange()), nil, false, nil, stubRange())
	_, err := getTarget(ctx, ec, exec, p, call)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "not permitted inside a derive")
}

func (s *RuntimeTestSuite) TestGetTargetModuleCallRejectsCallableArgument() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	ec.BindModule("mod", &ModuleBinding{Alias: "mod"})

	call := ast.NewCallExpression(ast.NewIdentifier("mod.fn", stubRange()), nil, false, nil, stubRange())
	target, err := getTarget(ctx, ec, exec, p, call)
	s.Require().NoError(err)

	_, err = target(ctx, box.Callable(callableStub{
		arity: 1,
		fn: func(_ context.Context, _ []box.Value) (box.Value, error) {
			return box.Undefined(), nil
		},
	}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "cannot pass callable value to module function")

	_, err = target(ctx, box.List([]box.Value{box.Callable(callableStub{
		arity: 1,
		fn: func(_ context.Context, _ []box.Value) (box.Value, error) {
			return box.Undefined(), nil
		},
	})}))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "module call mod.fn")
}

func (s *RuntimeTestSuite) TestCalculateHashKeyCallableArgumentReturnsEmpty() {
	node := &ast.CallExpression{}
	hash := calculateHashKey(node, []box.Value{box.Callable(callableStub{
		arity: 1,
		fn: func(_ context.Context, _ []box.Value) (box.Value, error) {
			return box.Undefined(), nil
		},
	})})
	s.Require().Empty(hash)
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQResolveErrorWithoutEvalDerive() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "slash_missing_no_eval_derive.sentrie",
		src: `namespace com/ex
derive helper = () => { yield 1 }
policy pol {
  let _s = 0
  rule ok = { yield true }
  export decision of ok
}`,
	})
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)

	_, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/ex/missing")
	s.Require().Error(err)
	s.Require().ErrorContains(err, "not found")
}

func (s *RuntimeTestSuite) TestLookupDeriveBySlashFQDefineFQNExportRejected() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{
			name: "alpha.secret.define_fqn.sentrie",
			src: `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}`,
		},
		deriveTestProgram{
			name: "beta.caller.define_fqn.sentrie",
			src: `namespace com/beta
derive caller = () => { yield 1 }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}`,
		},
	)
	secret := idx.DerivesByFQN["com/alpha/secret"]
	caller := idx.DerivesByFQN["com/beta/caller"]
	s.Require().NotNil(secret)
	s.Require().NotNil(caller)
	caller.DefineFQN = map[string]*index.Derive{"com/alpha/secret": secret}
	pol := idx.Namespaces["com/beta"].Policies["pol"]
	ec := NewExecutionContext(pol, exec)
	ec.evalDerive = caller

	_, err := lookupDeriveBySlashFQ(ec, exec, pol, "com/alpha/secret")
	s.Require().Error(err)
}
