// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/require"
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

func TestExecutionContextBranchCoverage(t *testing.T) {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	child := ec.AttachedChildContext()

	require.ErrorIs(t, child.InjectFact(context.Background(), "f", box.Number(1), false, nil), ErrIllegalFactInjection)

	require.NoError(t, ec.InjectFact(context.Background(), "factA", box.Number(10), false, nil))
	require.NoError(t, ec.InjectLet("letA", ast.NewVarDeclaration("letA", nil, ast.NewIntegerLiteral(1, stubRange()), stubRange())))
	require.Error(t, ec.InjectLet("letA", ast.NewVarDeclaration("letA", nil, ast.NewIntegerLiteral(2, stubRange()), stubRange())))

	child.SetLocal("factA", box.Number(99), false)
	local, ok := child.GetLocal("factA")
	require.True(t, ok)
	require.Equal(t, 99.0, local.Any())

	child.SetLocal("forced", box.String("x"), true)
	forced, ok := child.GetLocal("forced")
	require.True(t, ok)
	require.Equal(t, "x", forced.Any())

	require.NoError(t, ec.PushRefStack("a"))
	require.Error(t, ec.PushRefStack("a"))
	ec.PopRefStack()
	require.Empty(t, ec.GetRefStack())
}

func TestEvalQuantifiersAndMapBranches(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	undefinedCollection := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())
	anyExpr := ast.NewAnyExpression(undefinedCollection, "v", "", ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	anyResult, _, err := evalAny(ctx, ec, &executorImpl{}, p, anyExpr)
	require.NoError(t, err)
	require.Equal(t, false, anyResult.Any())

	allExpr := ast.NewAllExpression(ast.NewStringLiteral("bad", stubRange()), "v", "", ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	_, _, err = evalAll(ctx, ec, &executorImpl{}, p, allExpr)
	require.ErrorContains(t, err, "all expects list source")

	firstExpr := ast.NewFirstExpression(
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, stubRange())}, stubRange()),
		"v",
		"",
		ast.NewTrinaryLiteral(trinary.False, stubRange()),
		stubRange(),
	)
	firstResult, _, err := evalFirst(ctx, ec, &executorImpl{}, p, firstExpr)
	require.NoError(t, err)
	require.True(t, firstResult.IsUndefined())

	filterExpr := ast.NewFilterExpression(
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, stubRange()), ast.NewIntegerLiteral(2, stubRange())}, stubRange()),
		"v",
		"idx",
		ast.NewInfixExpression(ast.NewIdentifier("idx", stubRange()), ast.NewIntegerLiteral(0, stubRange()), ">", stubRange()),
		stubRange(),
	)
	filterResult, _, err := evalFilter(ctx, ec, &executorImpl{}, p, filterExpr)
	require.NoError(t, err)
	filtered, ok := filterResult.ListValue()
	require.True(t, ok)
	require.Len(t, filtered, 1)
	require.Equal(t, 2.0, filtered[0].Any())

	mapExpr := ast.NewMapExpression(
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(3, stubRange())}, stubRange()),
		"v",
		"idx",
		ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIdentifier("idx", stubRange()), "+", stubRange()),
		stubRange(),
	)
	mapResult, _, err := evalMap(ctx, ec, &executorImpl{}, p, mapExpr)
	require.NoError(t, err)
	mapped, ok := mapResult.ListValue()
	require.True(t, ok)
	require.Equal(t, 3.0, mapped[0].Any())
}

func TestEvalReduceTransformTernaryUnaryBlockCastBranches(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	undefinedCollection := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())
	reduceUndefinedExpr := ast.NewReduceExpression(undefinedCollection, ast.NewIntegerLiteral(0, stubRange()), "acc", "v", "i", ast.NewIdentifier("acc", stubRange()), stubRange())
	reduceUndefined, _, err := evalReduce(ctx, ec, &executorImpl{}, p, reduceUndefinedExpr)
	require.NoError(t, err)
	require.True(t, reduceUndefined.IsUndefined())

	reduceErrExpr := ast.NewReduceExpression(ast.NewStringLiteral("bad", stubRange()), ast.NewIntegerLiteral(0, stubRange()), "acc", "v", "", ast.NewIdentifier("acc", stubRange()), stubRange())
	_, _, err = evalReduce(ctx, ec, &executorImpl{}, p, reduceErrExpr)
	require.ErrorContains(t, err, "filter expects list source")

	transformExpr := ast.NewTransformExpression(ast.NewIntegerLiteral(1, stubRange()), "noop", stubRange())
	_, _, err = evalTransform(ctx, ec, &executorImpl{}, p, transformExpr)
	require.ErrorIs(t, err, xerr.ErrNotImplemented)

	thenExpr := ast.NewTernaryExpression(ast.NewTrinaryLiteral(trinary.True, stubRange()), ast.NewIntegerLiteral(10, stubRange()), ast.NewIntegerLiteral(20, stubRange()), stubRange())
	thenResult, _, err := evalTernary(ctx, ec, &executorImpl{}, p, thenExpr)
	require.NoError(t, err)
	require.Equal(t, 10.0, thenResult.Any())

	elseExpr := ast.NewTernaryExpression(ast.NewTrinaryLiteral(trinary.False, stubRange()), ast.NewIntegerLiteral(10, stubRange()), ast.NewIntegerLiteral(20, stubRange()), stubRange())
	elseResult, _, err := evalTernary(ctx, ec, &executorImpl{}, p, elseExpr)
	require.NoError(t, err)
	require.Equal(t, 20.0, elseResult.Any())

	unaryNot, _, err := evalUnary(ctx, ec, &executorImpl{}, p, ast.NewUnaryExpression("not", ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange()))
	require.NoError(t, err)
	require.Equal(t, trinary.False, unaryNot.Any())

	unaryErrExpr := ast.NewUnaryExpression("+", ast.NewStringLiteral("x", stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, &executorImpl{}, p, unaryErrExpr)
	require.ErrorContains(t, err, "unary + requires number")

	blockExpr := ast.NewBlockExpression(
		[]ast.Statement{
			ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(1, stubRange()), stubRange()),
			ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(2, stubRange()), stubRange()),
		},
		ast.NewIdentifier("x", stubRange()),
		stubRange(),
	)
	_, _, err = evalBlock(ctx, ec, &executorImpl{}, p, blockExpr)
	require.Error(t, err)

	castBoolExpr := ast.NewCastExpression(ast.NewUnaryExpression("!", ast.NewTrinaryLiteral(trinary.False, stubRange()), stubRange()), ast.NewNumberTypeRef(stubRange()), stubRange())
	castBool, _, err := evalCast(ctx, ec, &executorImpl{}, p, castBoolExpr)
	require.NoError(t, err)
	require.Equal(t, 1.0, castBool.Any())

	castParseErr := ast.NewCastExpression(ast.NewStringLiteral("abc", stubRange()), ast.NewNumberTypeRef(stubRange()), stubRange())
	_, _, err = evalCast(ctx, ec, &executorImpl{}, p, castParseErr)
	require.Error(t, err)

	castListErr := ast.NewCastExpression(ast.NewIntegerLiteral(1, stubRange()), ast.NewListTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()), stubRange())
	_, _, err = evalCast(ctx, ec, &executorImpl{}, p, castListErr)
	require.ErrorContains(t, err, "cannot cast number to list")
}
