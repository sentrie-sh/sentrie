// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
)

// TestLambdaCapture_LateBoundLexicalFollowsParentContextByReference ensures closure
// capture uses the live parent execution context (v1), not a snapshot at creation time.
func (s *RuntimeTestSuite) TestLambdaCapture_LateBoundLexicalFollowsParentContextByReference() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	s.Require().NoError(ec.InjectLet("cell", ast.NewVarDeclaration(
		"cell",
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
		stubRange(),
	)))

	// filter(list, (item) => item == cell) - cell is read when the lambda runs, not copied at creation.
	list := ast.NewListLiteral([]ast.Expression{
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
	}, stubRange())
	lam := stubLambda([]string{"item"}, ast.NewInfixExpression(
		ast.NewIdentifier("item", stubRange()),
		ast.NewIdentifier("cell", stubRange()),
		"==",
		stubRange(),
	))
	call := ast.NewCallExpression(ast.NewIdentifier("filter", stubRange()), []ast.Expression{list, lam}, false, nil, stubRange())

	out1, _, err := eval(ctx, ec, exec, p, call)
	s.Require().NoError(err)
	lv1, ok := out1.ListValue()
	s.Require().True(ok)
	s.Require().Len(lv1, 1)

	// Mutate the lexical binding in the parent context; snapshot capture would still filter for 1.
	ec.SetLocal("cell", box.Number(2), true)

	out2, _, err := eval(ctx, ec, exec, p, call)
	s.Require().NoError(err)
	lv2, ok := out2.ListValue()
	s.Require().True(ok)
	s.Require().Len(lv2, 1)
	s.Require().Equal(2.0, lv2[0].Any())
}
