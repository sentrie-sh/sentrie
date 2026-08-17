// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestEvalCastValidatesConstrainedStringTarget() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}

	typeRef := ast.NewStringTypeRef(stubRange())
	constraint := ast.NewTypeRefConstraint(
		"maxlength",
		[]ast.Expression{ast.NewIntegerLiteral(5, stubRange())},
		stubRange(),
	)
	s.Require().NoError(typeRef.AddConstraint(constraint))

	castExpr := ast.NewCastExpression(
		ast.NewStringLiteral("hello world", stubRange()),
		typeRef,
		stubRange(),
	)
	ec := NewExecutionContext(p, exec)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "constraint failed")
	s.Require().False(result.IsValid())
	s.Require().NotNil(node)

	okCast := ast.NewCastExpression(
		ast.NewStringLiteral("hi", stubRange()),
		typeRef,
		stubRange(),
	)
	result, node, err = evalCast(ctx, ec, exec, p, okCast)
	s.Require().NoError(err)
	s.Require().Equal("hi", result.Any())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNumberConversionSuccess() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewStringLiteral("42", stubRange()),
		ast.NewNumberTypeRef(stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(42.0, result.Any())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastListAssertionFailureReturnsError() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewListTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()),
		stubRange(),
	)
	_, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot cast")
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastListAssertionSuccess() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, stubRange())}, stubRange()),
		ast.NewListTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(box.ValueList, result.Kind())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastTrinaryTarget() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewIntegerLiteral(0, stubRange()),
		ast.NewTrinaryTypeRef(stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(trinary.False, box.TrinaryFrom(result))
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastOperandEvalError() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewIdentifier("missing_operand", stubRange()),
		ast.NewStringTypeRef(stubRange()),
		stubRange(),
	)
	_, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastDictAssertionSuccess() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	s.Require().NoError(ec.InjectFact(ctx, "d", box.Dict(map[string]box.Value{"k": box.Number(1)}), false, nil))

	castExpr := ast.NewCastExpression(
		ast.NewIdentifier("d", stubRange()),
		ast.NewDictTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(box.ValueDict, result.Kind())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastDictAssertionFailure() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewDictTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()),
		stubRange(),
	)
	_, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot cast")
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNumberFromBoolValue() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	s.Require().NoError(ec.InjectFact(ctx, "b", box.FromAny(true), false, nil))

	castExpr := ast.NewCastExpression(
		ast.NewIdentifier("b", stubRange()),
		ast.NewNumberTypeRef(stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(1.0, result.Any())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNullableTargetPassesThrough() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewStringLiteral("ok", stubRange()),
		ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal("ok", result.Any())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNumberParseFailure() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewStringLiteral("not-a-number", stubRange()),
		ast.NewNumberTypeRef(stubRange()),
		stubRange(),
	)
	_, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNumberFromString() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewStringLiteral("3.5", stubRange()),
		ast.NewNumberTypeRef(stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal(3.5, result.Any())
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastNumberFromBool() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewTrinaryLiteral(trinary.False, stubRange()),
		ast.NewNumberTypeRef(stubRange()),
		stubRange(),
	)
	_, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot cast")
	s.Require().NotNil(node)
}

func (s *RuntimeTestSuite) TestEvalCastTrinaryToString() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)

	castExpr := ast.NewCastExpression(
		ast.NewTrinaryLiteral(trinary.True, stubRange()),
		ast.NewStringTypeRef(stubRange()),
		stubRange(),
	)
	result, node, err := evalCast(ctx, ec, exec, p, castExpr)
	s.Require().NoError(err)
	s.Require().Equal("true", result.Any())
	s.Require().NotNil(node)
}
