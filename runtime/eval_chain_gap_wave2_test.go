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
	"errors"

	"github.com/binaek/perch"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *RuntimeTestSuite) TestEvalLiteralBranchesAndMapKeyTypeError() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	nullExpr := ast.NewNullLiteral(stubRange())
	nullValue, _, err := eval(ctx, ec, &executorImpl{}, p, nullExpr)
	s.Require().NoError(err)
	s.Require().True(nullValue.IsNull())

	floatExpr := ast.NewFloatLiteral(3.25, stubRange())
	floatValue, _, err := eval(ctx, ec, &executorImpl{}, p, floatExpr)
	s.Require().NoError(err)
	s.Require().Equal(3.25, floatValue.Any())

	badMap := ast.NewMapLiteral([]ast.MapEntry{
		{
			Key:   ast.NewIntegerLiteral(1, stubRange()),
			Value: ast.NewStringLiteral("x", stubRange()),
		},
	}, stubRange())
	_, _, err = eval(ctx, ec, &executorImpl{}, p, badMap)
	s.Require().ErrorContains(err, "map key is not a string")
}

func (s *RuntimeTestSuite) TestEvalIdentLetLocalAndMissingPaths() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	s.Require().NoError(ec.InjectLet("from_let", ast.NewVarDeclaration(
		"from_let",
		nil,
		ast.NewIntegerLiteral(7, stubRange()),
		stubRange(),
	)))
	letValue, _, err := evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("from_let", stubRange()))
	s.Require().NoError(err)
	s.Require().Equal(7.0, letValue.Any())
	cached, ok := ec.GetLocal("from_let")
	s.Require().True(ok)
	s.Require().Equal(7.0, cached.Any())

	s.Require().NoError(ec.InjectFact(ctx, "x", box.Number(1), false, nil))
	s.Require().NoError(ec.InjectLet("x", ast.NewVarDeclaration(
		"x",
		nil,
		ast.NewIntegerLiteral(2, stubRange()),
		stubRange(),
	)))
	ec.SetLocal("x", box.Number(99), true)
	localFirst, _, err := evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("x", stubRange()))
	s.Require().NoError(err)
	s.Require().Equal(99.0, localFirst.Any())

	_, _, err = evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("missing_symbol", stubRange()))
	s.Require().ErrorContains(err, "identifier not found: missing_symbol")
}

func (s *RuntimeTestSuite) TestEvalIdentTypedLetErrorWrapsUnderlyingValidationError() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	typedLet := ast.NewVarDeclaration(
		"typed_let",
		ast.NewStringTypeRef(stubRange()),
		ast.NewIntegerLiteral(123, stubRange()),
		stubRange(),
	)
	s.Require().NoError(ec.InjectLet("typed_let", typedLet))

	_, _, err := evalIdent(ctx, ec, &executorImpl{}, p, ast.NewIdentifier("typed_let", stubRange()))
	s.Require().Error(err)
	s.Require().ErrorContains(err, "invalid value for let declaration typed_let")
	s.Require().ErrorContains(err, "value 123 is not a string")

	underlying := errors.Unwrap(err)
	s.Require().Error(underlying)
	s.Require().ErrorContains(underlying, "value 123 is not a string")
}

func (s *RuntimeTestSuite) TestEvalCallMemoizedHitAndMiss() {
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
	Builtins[builtinName] = func(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
		callCount++
		if len(args) > 0 {
			return args[0], nil
		}
		return box.Undefined(), nil
	}

	ec := NewExecutionContext(p, exec)
	s.Require().NoError(ec.InjectFact(ctx, "memo_arg", box.Number(1), false, nil))
	memoizedCall := ast.NewCallExpression(
		ast.NewIdentifier(builtinName, stubRange()),
		[]ast.Expression{ast.NewIdentifier("memo_arg", stubRange())},
		true,
		nil,
		stubRange(),
	)

	first, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	s.Require().NoError(err)
	s.Require().Equal(1.0, first.Any())
	s.Require().Equal(1, callCount)

	second, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	s.Require().NoError(err)
	s.Require().Equal(1.0, second.Any())
	s.Require().Equal(1, callCount)

	s.Require().NoError(ec.InjectFact(ctx, "memo_arg", box.Number(2), false, nil))
	third, _, err := evalCall(ctx, ec, exec, p, memoizedCall)
	s.Require().NoError(err)
	s.Require().Equal(2.0, third.Any())
	s.Require().Equal(2, callCount)
}

func (s *RuntimeTestSuite) TestEvalCallInjectedErrorPassthrough() {
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
	s.Require().Error(err)
	s.Require().ErrorIs(err, xerr.InjectedError{})
	s.Require().ErrorContains(err, "boom")
	s.Require().NotContains(err.Error(), "failed to call function")
}

func (s *RuntimeTestSuite) TestEvalInfixAndUnaryUnsupportedAndInvalidOperands() {
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
	s.Require().ErrorContains(err, "unsupported infix op: ???")

	leftMismatch := ast.NewInfixExpression(
		ast.NewStringLiteral("x", stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
		"-",
		stubRange(),
	)
	_, _, err = evalInfix(ctx, ec, &executorImpl{}, p, leftMismatch)
	s.Require().ErrorContains(err, "left operand is not a number")

	rightMismatch := ast.NewInfixExpression(
		ast.NewIntegerLiteral(2, stubRange()),
		ast.NewStringLiteral("x", stubRange()),
		"-",
		stubRange(),
	)
	_, _, err = evalInfix(ctx, ec, &executorImpl{}, p, rightMismatch)
	s.Require().ErrorContains(err, "right operand is not a number")

	unsupportedUnary := ast.NewUnaryExpression("~", ast.NewIntegerLiteral(1, stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, &executorImpl{}, p, unsupportedUnary)
	s.Require().ErrorContains(err, "unsupported unary op: ~")

	invalidUnaryOperand := ast.NewUnaryExpression("-", ast.NewStringLiteral("x", stubRange()), stubRange())
	_, _, err = evalUnary(ctx, ec, &executorImpl{}, p, invalidUnaryOperand)
	s.Require().ErrorContains(err, "unary - requires number")
}

func (s *RuntimeTestSuite) TestImportDecisionSuccessWithWithInjection() {
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
	s.Require().NoError(err)
	s.Require().NotNil(node)
	outMap, ok := out.MapValue()
	s.Require().True(ok)
	s.Require().Equal(trinary.True, outMap["state"].Any())
	s.Require().Equal(9.0, outMap["value"].Any())
}
