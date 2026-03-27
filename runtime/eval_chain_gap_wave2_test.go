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

	"github.com/binaek/perch"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/require"
)

func TestEvalLiteralBranchesAndMapKeyTypeError(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	nullExpr := ast.NewNullLiteral(stubRange())
	nullValue, _, err := eval(ctx, ec, &executorImpl{}, p, nullExpr)
	require.NoError(t, err)
	require.True(t, nullValue.IsNull())

	floatExpr := ast.NewFloatLiteral(3.25, stubRange())
	floatValue, _, err := eval(ctx, ec, &executorImpl{}, p, floatExpr)
	require.NoError(t, err)
	require.Equal(t, 3.25, floatValue.Any())

	badMap := ast.NewMapLiteral([]ast.MapEntry{
		{
			Key:   ast.NewIntegerLiteral(1, stubRange()),
			Value: ast.NewStringLiteral("x", stubRange()),
		},
	}, stubRange())
	_, _, err = eval(ctx, ec, &executorImpl{}, p, badMap)
	require.ErrorContains(t, err, "map key is not a string")
}

func TestEvalIdentLetLocalAndMissingPaths(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	require.NoError(t, ec.InjectLet("from_let", ast.NewVarDeclaration(
		"from_let",
		nil,
		ast.NewIntegerLiteral(7, stubRange()),
		stubRange(),
	)))
	letValue, _, err := evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("from_let", stubRange()))
	require.NoError(t, err)
	require.Equal(t, 7.0, letValue.Any())
	cached, ok := ec.GetLocal("from_let")
	require.True(t, ok)
	require.Equal(t, 7.0, cached.Any())

	require.NoError(t, ec.InjectFact(ctx, "x", Number(1), false, nil))
	require.NoError(t, ec.InjectLet("x", ast.NewVarDeclaration(
		"x",
		nil,
		ast.NewIntegerLiteral(2, stubRange()),
		stubRange(),
	)))
	ec.SetLocal("x", Number(99), true)
	localFirst, _, err := evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("x", stubRange()))
	require.NoError(t, err)
	require.Equal(t, 99.0, localFirst.Any())

	_, _, err = evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("missing_symbol", stubRange()))
	require.ErrorContains(t, err, "identifier not found: missing_symbol")
}

func TestEvalCallMemoizedHitAndMiss(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	exec := &executorImpl{
		callMemoizePerch: perch.New[any](1 << 20),
	}
	exec.callMemoizePerch.Reserve()

	const builtinName = "test_wave2_memoized_builtin"
	original, hadOriginal := Builtins[builtinName]
	defer func() {
		if hadOriginal {
			Builtins[builtinName] = original
			return
		}
		delete(Builtins, builtinName)
	}()

	callCount := 0
	Builtins[builtinName] = func(_ context.Context, args []any) (any, error) {
		callCount++
		return args[0], nil
	}

	ec := NewExecutionContext(p, exec)
	require.NoError(t, ec.InjectFact(ctx, "memo_arg", Number(1), false, nil))
	memoizedCall := ast.NewCallExpression(
		ast.NewIdentifier(builtinName, stubRange()),
		[]ast.Expression{ast.NewIdentifier("memo_arg", stubRange())},
		true,
		nil,
		stubRange(),
	)

	first, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	require.NoError(t, err)
	require.Equal(t, 1.0, first.Any())
	require.Equal(t, 1, callCount)

	second, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	require.NoError(t, err)
	require.Equal(t, 1.0, second.Any())
	require.Equal(t, 1, callCount)

	require.NoError(t, ec.InjectFact(ctx, "memo_arg", Number(2), false, nil))
	third, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	require.NoError(t, err)
	require.Equal(t, 2.0, third.Any())
	require.Equal(t, 2, callCount)
}

func TestEvalCallInjectedErrorPassthrough(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	exec := &executorImpl{
		callMemoizePerch: perch.New[any](1 << 20),
	}
	exec.callMemoizePerch.Reserve()
	ec := NewExecutionContext(p, exec)

	call := ast.NewCallExpression(
		ast.NewIdentifier("error", stubRange()),
		[]ast.Expression{ast.NewStringLiteral("boom", stubRange())},
		false,
		nil,
		stubRange(),
	)
	_, _, err := evalCall(ctx, ec, exec, p, call)
	require.Error(t, err)
	require.ErrorIs(t, err, xerr.InjectedError{})
	require.ErrorContains(t, err, "boom")
	require.NotContains(t, err.Error(), "failed to call function")
}

func TestEvalInfixAndUnaryUnsupportedAndInvalidOperands(t *testing.T) {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	unsupportedInfix := ast.NewInfixExpression(
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
		"???",
		stubRange(),
	)
	_, _, err := evalInfix(ctx, ec, &executorImpl{}, p, unsupportedInfix)
	require.ErrorContains(t, err, "unsupported infix op: ???")

	leftMismatch := ast.NewInfixExpression(
		ast.NewStringLiteral("x", stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
		"-",
		stubRange(),
	)
	_, _, err = evalInfix(ctx, ec, &executorImpl{}, p, leftMismatch)
	require.ErrorContains(t, err, "left operand is not a number")

	rightMismatch := ast.NewInfixExpression(
		ast.NewIntegerLiteral(2, stubRange()),
		ast.NewStringLiteral("x", stubRange()),
		"-",
		stubRange(),
	)
	_, _, err = evalInfix(ctx, ec, &executorImpl{}, p, rightMismatch)
	require.ErrorContains(t, err, "right operand is not a number")

	unsupportedUnary := ast.NewUnaryExpression("~", ast.NewIntegerLiteral(1, stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, &executorImpl{}, p, unsupportedUnary)
	require.ErrorContains(t, err, "unsupported unary op: ~")

	invalidUnaryOperand := ast.NewUnaryExpression("-", ast.NewStringLiteral("x", stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, &executorImpl{}, p, invalidUnaryOperand)
	require.ErrorContains(t, err, "unary - requires number")
}

func TestImportDecisionSuccessWithWithInjection(t *testing.T) {
	ctx := context.Background()
	idx := index.CreateIndex()
	exec := &executorImpl{
		index: idx,
	}

	nsFQN := ast.NewFQN([]string{"test", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	targetPolicy := &index.Policy{
		Name:      "target",
		FQN:       ast.CreateFQN(nsFQN, "target"),
		Namespace: ns,
		Lets:      map[string]*ast.VarDeclaration{},
		Facts:     map[string]*ast.FactStatement{},
		Rules:     map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{
			"allow": {RuleName: "allow"},
		},
		Uses:   map[string]*ast.UseStatement{},
		Shapes: map[string]*index.Shape{},
	}
	targetPolicy.Facts["f"] = ast.NewFactStatement("f", nil, "f", nil, false, stubRange())
	targetRuleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewIdentifier("f", stubRange()), stubRange())
	targetPolicy.Rules["allow"] = &index.Rule{
		Node:   targetRuleStmt,
		Policy: targetPolicy,
		Name:   "allow",
		FQN:    ast.CreateFQN(targetPolicy.FQN, "allow"),
		Body:   targetRuleStmt.Body,
	}
	ns.Policies[targetPolicy.Name] = targetPolicy

	callerPolicy := &index.Policy{
		Name:        "caller",
		FQN:         ast.CreateFQN(nsFQN, "caller"),
		Namespace:   ns,
		Lets:        map[string]*ast.VarDeclaration{},
		Facts:       map[string]*ast.FactStatement{},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	ec := NewExecutionContext(callerPolicy, exec)

	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"test", "ns", "target"}, stubRange()).Ptr(),
		[]*ast.WithClause{
			ast.NewWithClause("f", ast.NewIntegerLiteral(9, stubRange()), stubRange()),
			ast.NewWithClause("unused", ast.NewIntegerLiteral(100, stubRange()), stubRange()),
		},
		stubRange(),
	)

	out, node, err := ImportDecision(ctx, exec, ec, callerPolicy, imp)
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, 9.0, out.Any())
}
