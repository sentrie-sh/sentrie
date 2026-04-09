// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func newEvalTestPolicy() *index.Policy {
	ns := ast.NewFQN([]string{"test", "ns"}, stubRange())
	return &index.Policy{
		Name: "pol",
		FQN:  ast.CreateFQN(ns, "pol"),
		Namespace: &index.Namespace{
			FQN: ns,
		},
		Rules:       map[string]*index.Rule{},
		Facts:       map[string]*ast.FactStatement{},
		Lets:        map[string]*ast.VarDeclaration{},
		RuleExports: map[string]*index.ExportedRule{},
	}
}

func stubLambda(params []string, yield ast.Expression) *ast.LambdaExpression {
	return ast.NewLambdaExpression(params, ast.NewBlockExpression(nil, yield, stubRange()), stubRange())
}

func (s *RuntimeTestSuite) TestExecutionContextBranchCoverage() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	child := ec.AttachedChildContext()

	s.Require().ErrorIs(child.InjectFact(context.Background(), "f", box.Number(1), false, nil), ErrIllegalFactInjection)

	s.Require().NoError(ec.InjectFact(context.Background(), "factA", box.Number(10), false, nil))
	s.Require().NoError(ec.InjectLet("letA", ast.NewVarDeclaration("letA", nil, ast.NewIntegerLiteral(1, stubRange()), stubRange())))
	s.Require().Error(ec.InjectLet("letA", ast.NewVarDeclaration("letA", nil, ast.NewIntegerLiteral(2, stubRange()), stubRange())))

	child.SetLocal("factA", box.Number(99), false)
	local, ok := child.GetLocal("factA")
	s.Require().True(ok)
	s.Require().Equal(99.0, local.Any())

	child.SetLocal("forced", box.String("x"), true)
	forced, ok := child.GetLocal("forced")
	s.Require().True(ok)
	s.Require().Equal("x", forced.Any())

	s.Require().NoError(ec.PushRefStack("a"))
	s.Require().Error(ec.PushRefStack("a"))
	ec.PopRefStack()
	s.Require().Empty(ec.GetRefStack())
}

func (s *RuntimeTestSuite) TestEvalQuantifiersAndMapBranches() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	undefinedCollection := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())
	anyCall := ast.NewCallExpression(ast.NewIdentifier("any", stubRange()), []ast.Expression{
		undefinedCollection,
		stubLambda([]string{"v"}, ast.NewTrinaryLiteral(trinary.True, stubRange())),
	}, false, nil, stubRange())
	anyResult, _, err := eval(ctx, ec, exec, p, anyCall)
	s.Require().NoError(err)
	s.Require().Equal(false, anyResult.Any())

	allCall := ast.NewCallExpression(ast.NewIdentifier("all", stubRange()), []ast.Expression{
		ast.NewStringLiteral("bad", stubRange()),
		stubLambda([]string{"v"}, ast.NewTrinaryLiteral(trinary.True, stubRange())),
	}, false, nil, stubRange())
	_, _, err = eval(ctx, ec, exec, p, allCall)
	s.Require().ErrorContains(err, "all: first argument must be a list")

	firstCall := ast.NewCallExpression(ast.NewIdentifier("first", stubRange()), []ast.Expression{
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, stubRange())}, stubRange()),
		stubLambda([]string{"v"}, ast.NewTrinaryLiteral(trinary.False, stubRange())),
	}, false, nil, stubRange())
	firstResult, _, err := eval(ctx, ec, exec, p, firstCall)
	s.Require().NoError(err)
	s.Require().True(firstResult.IsUndefined())

	filterCall := ast.NewCallExpression(ast.NewIdentifier("filter", stubRange()), []ast.Expression{
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, stubRange()), ast.NewIntegerLiteral(2, stubRange())}, stubRange()),
		stubLambda([]string{"v", "idx"}, ast.NewInfixExpression(ast.NewIdentifier("idx", stubRange()), ast.NewIntegerLiteral(0, stubRange()), ">", stubRange())),
	}, false, nil, stubRange())
	filterResult, _, err := eval(ctx, ec, exec, p, filterCall)
	s.Require().NoError(err)
	filtered, ok := filterResult.ListValue()
	s.Require().True(ok)
	s.Require().Len(filtered, 1)
	s.Require().Equal(2.0, filtered[0].Any())

	mapCall := ast.NewCallExpression(ast.NewIdentifier("map", stubRange()), []ast.Expression{
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(3, stubRange())}, stubRange()),
		stubLambda([]string{"v", "idx"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIdentifier("idx", stubRange()), "+", stubRange())),
	}, false, nil, stubRange())
	mapResult, _, err := eval(ctx, ec, exec, p, mapCall)
	s.Require().NoError(err)
	mapped, ok := mapResult.ListValue()
	s.Require().True(ok)
	s.Require().Equal(3.0, mapped[0].Any())
}

func (s *RuntimeTestSuite) TestEvalReduceTransformTernaryUnaryBlockCastBranches() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	undefinedCollection := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())
	reduceUndefinedExpr := ast.NewCallExpression(ast.NewIdentifier("reduce", stubRange()), []ast.Expression{
		undefinedCollection,
		ast.NewIntegerLiteral(0, stubRange()),
		stubLambda([]string{"acc", "v", "i"}, ast.NewIdentifier("acc", stubRange())),
	}, false, nil, stubRange())
	reduceUndefined, _, err := eval(ctx, ec, exec, p, reduceUndefinedExpr)
	s.Require().NoError(err)
	s.Require().True(reduceUndefined.IsUndefined())

	reduceErrExpr := ast.NewCallExpression(ast.NewIdentifier("reduce", stubRange()), []ast.Expression{
		ast.NewStringLiteral("bad", stubRange()),
		ast.NewIntegerLiteral(0, stubRange()),
		stubLambda([]string{"acc", "v"}, ast.NewIdentifier("acc", stubRange())),
	}, false, nil, stubRange())
	_, _, err = eval(ctx, ec, exec, p, reduceErrExpr)
	s.Require().ErrorContains(err, "reduce: first argument must be a list")

	transformExpr := ast.NewTransformExpression(ast.NewIntegerLiteral(1, stubRange()), "noop", stubRange())
	_, _, err = evalTransform(ctx, ec, exec, p, transformExpr)
	s.Require().ErrorIs(err, xerr.ErrNotImplemented)

	thenExpr := ast.NewTernaryExpression(ast.NewTrinaryLiteral(trinary.True, stubRange()), ast.NewIntegerLiteral(10, stubRange()), ast.NewIntegerLiteral(20, stubRange()), stubRange())
	thenResult, _, err := evalTernary(ctx, ec, exec, p, thenExpr)
	s.Require().NoError(err)
	s.Require().Equal(10.0, thenResult.Any())

	elseExpr := ast.NewTernaryExpression(ast.NewTrinaryLiteral(trinary.False, stubRange()), ast.NewIntegerLiteral(10, stubRange()), ast.NewIntegerLiteral(20, stubRange()), stubRange())
	elseResult, _, err := evalTernary(ctx, ec, exec, p, elseExpr)
	s.Require().NoError(err)
	s.Require().Equal(20.0, elseResult.Any())

	unaryNot, _, err := evalUnary(ctx, ec, exec, p, ast.NewUnaryExpression("not", ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()))
	s.Require().NoError(err)
	s.Require().Equal(trinary.False, unaryNot.Any())

	unaryErrExpr := ast.NewUnaryExpression("+", ast.NewStringLiteral("x", stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, exec, p, unaryErrExpr)
	s.Require().ErrorContains(err, "unary + requires number")

	blockExpr := ast.NewBlockExpression(
		[]ast.Statement{
			ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(1, stubRange()), stubRange()),
			ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(2, stubRange()), stubRange()),
		},
		ast.NewIdentifier("x", stubRange()),
		stubRange(),
	)
	_, _, err = evalBlock(ctx, ec, exec, p, blockExpr)
	s.Require().Error(err)

	castBoolExpr := ast.NewCastExpression(ast.NewUnaryExpression("!", ast.NewTrinaryLiteral(trinary.False, stubRange()), stubRange()), ast.NewNumberTypeRef(stubRange()), stubRange())
	_, _, err = evalCast(ctx, ec, exec, p, castBoolExpr)
	s.Require().ErrorContains(err, "cannot cast trinary to number")

	castParseErr := ast.NewCastExpression(ast.NewStringLiteral("abc", stubRange()), ast.NewNumberTypeRef(stubRange()), stubRange())
	_, _, err = evalCast(ctx, ec, exec, p, castParseErr)
	s.Require().Error(err)

	castListErr := ast.NewCastExpression(ast.NewIntegerLiteral(1, stubRange()), ast.NewListTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()), stubRange())
	_, _, err = evalCast(ctx, ec, exec, p, castListErr)
	s.Require().ErrorContains(err, "cannot cast number to list")
}
