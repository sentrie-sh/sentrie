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
)

// TestLambdaCapture_LateBoundLexicalFollowsParentContextByReference ensures closure
// capture uses the live parent execution context (v1), not a snapshot at creation time.
func (s *RuntimeTestSuite) TestLambdaCapture_LateBoundLexicalFollowsParentContextByReference() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	s.Require().NoError(ec.InjectLet("cell", ast.NewVarDeclaration(
		"cell",
		nil,
		ast.NewIntegerLiteral(1, stubRange()),
		stubRange(),
	)))

	// filter(list, (item) => item == cell) — cell is read when the lambda runs, not copied at creation.
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
