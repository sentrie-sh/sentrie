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
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *RuntimeTestSuite) TestEvalDispatchByExpressionKind() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	exec := &executorImpl{}

	tests := []struct {
		name       string
		expr       ast.Expression
		setup      func(*ExecutionContext)
		wantAny    any
		wantErr    string
		checkValue func(*testing.T, box.Value)
	}{
		{
			name:    "preceding comment routes to wrapped literal",
			expr:    ast.NewPrecedingCommentExpression("comment", ast.NewIntegerLiteral(7, stubRange()), stubRange()),
			wantAny: 7.0,
		},
		{
			name:    "trailing comment routes to wrapped literal",
			expr:    ast.NewTrailingCommentExpression("comment", ast.NewStringLiteral("ok", stubRange()), stubRange()),
			wantAny: "ok",
		},
		{
			name:    "identifier dispatch",
			expr:    ast.NewIdentifier("x", stubRange()),
			setup:   func(ec *ExecutionContext) { s.Require().NoError(ec.InjectFact(ctx, "x", box.Number(11), false, nil)) },
			wantAny: 11.0,
		},
		{
			name: "call dispatch for builtin",
			expr: ast.NewCallExpression(
				ast.NewIdentifier("as_list", stubRange()),
				[]ast.Expression{ast.NewIntegerLiteral(5, stubRange())},
				false,
				nil,
				stubRange(),
			),
			wantAny: []any{5.0},
		},
		{
			name: "cast dispatch",
			expr: ast.NewCastExpression(
				ast.NewStringLiteral("3.5", stubRange()),
				ast.NewNumberTypeRef(stubRange()),
				stubRange(),
			),
			wantAny: 3.5,
		},
		{
			name: "infix dispatch",
			expr: ast.NewInfixExpression(
				ast.NewIntegerLiteral(2, stubRange()),
				ast.NewIntegerLiteral(3, stubRange()),
				"+",
				stubRange(),
			),
			wantAny: 5.0,
		},
		{
			name: "unary dispatch",
			expr: ast.NewUnaryExpression(
				"!",
				ast.NewTrinaryLiteral(trinary.False, stubRange()),
				stubRange(),
			),
			wantAny: trinary.True,
		},
		{
			name: "ternary dispatch",
			expr: ast.NewTernaryExpression(
				ast.NewTrinaryLiteral(trinary.True, stubRange()),
				ast.NewIntegerLiteral(10, stubRange()),
				ast.NewIntegerLiteral(20, stubRange()),
				stubRange(),
			),
			wantAny: 10.0,
		},
		{
			name: "field access dispatch",
			expr: ast.NewFieldAccessExpression(
				ast.NewMapLiteral([]ast.MapEntry{
					{
						Key:   ast.NewStringLiteral("key", stubRange()),
						Value: ast.NewIntegerLiteral(9, stubRange()),
					},
				}, stubRange()),
				"key",
				stubRange(),
			),
			wantAny: 9.0,
		},
		{
			name: "index access dispatch",
			expr: ast.NewIndexAccessExpression(
				ast.NewListLiteral([]ast.Expression{
					ast.NewStringLiteral("a", stubRange()),
					ast.NewStringLiteral("b", stubRange()),
				}, stubRange()),
				ast.NewIntegerLiteral(1, stubRange()),
				stubRange(),
			),
			wantAny: "b",
		},
		{
			name: "block dispatch",
			expr: ast.NewBlockExpression(
				[]ast.Statement{
					ast.NewVarDeclaration("y", nil, ast.NewIntegerLiteral(42, stubRange()), stubRange()),
				},
				ast.NewIdentifier("y", stubRange()),
				stubRange(),
			),
			wantAny: 42.0,
		},
		{
			name: "any builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("any", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(1, stubRange()),
				}, stubRange()),
				stubLambda([]string{"v"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIntegerLiteral(1, stubRange()), "==", stubRange())),
			}, false, nil, stubRange()),
			wantAny: true,
		},
		{
			name: "all builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("all", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(2, stubRange()),
					ast.NewIntegerLiteral(3, stubRange()),
				}, stubRange()),
				stubLambda([]string{"v"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIntegerLiteral(1, stubRange()), ">", stubRange())),
			}, false, nil, stubRange()),
			wantAny: true,
		},
		{
			name: "first builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("first", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(1, stubRange()),
					ast.NewIntegerLiteral(2, stubRange()),
				}, stubRange()),
				stubLambda([]string{"v"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIntegerLiteral(1, stubRange()), ">", stubRange())),
			}, false, nil, stubRange()),
			wantAny: 2.0,
		},
		{
			name: "filter builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("filter", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(1, stubRange()),
					ast.NewIntegerLiteral(2, stubRange()),
				}, stubRange()),
				stubLambda([]string{"v"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIntegerLiteral(1, stubRange()), ">", stubRange())),
			}, false, nil, stubRange()),
			wantAny: []any{2.0},
		},
		{
			name: "reduce builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("reduce", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(1, stubRange()),
					ast.NewIntegerLiteral(2, stubRange()),
					ast.NewIntegerLiteral(3, stubRange()),
				}, stubRange()),
				ast.NewIntegerLiteral(0, stubRange()),
				stubLambda([]string{"acc", "v"}, ast.NewInfixExpression(ast.NewIdentifier("acc", stubRange()), ast.NewIdentifier("v", stubRange()), "+", stubRange())),
			}, false, nil, stubRange()),
			wantAny: 6.0,
		},
		{
			name: "map builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("map", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(3, stubRange()),
				}, stubRange()),
				stubLambda([]string{"v", "i"}, ast.NewInfixExpression(ast.NewIdentifier("v", stubRange()), ast.NewIdentifier("i", stubRange()), "+", stubRange())),
			}, false, nil, stubRange()),
			wantAny: []any{3.0},
		},
		{
			name: "distinct builtin dispatch",
			expr: ast.NewCallExpression(ast.NewIdentifier("distinct", stubRange()), []ast.Expression{
				ast.NewListLiteral([]ast.Expression{
					ast.NewIntegerLiteral(1, stubRange()),
					ast.NewIntegerLiteral(1, stubRange()),
					ast.NewIntegerLiteral(2, stubRange()),
				}, stubRange()),
			}, false, nil, stubRange()),
			wantAny: []any{1.0, 2.0},
		},
		{
			name: "transform dispatch reaches not implemented boundary",
			expr: ast.NewTransformExpression(
				ast.NewIntegerLiteral(1, stubRange()),
				"noop",
				stubRange(),
			),
			wantErr: xerr.ErrNotImplemented.Error(),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ec := NewExecutionContext(p, exec)
			if tt.setup != nil {
				tt.setup(ec)
			}

			got, node, err := eval(ctx, ec, exec, p, tt.expr)
			s.Require().NotNil(node)

			if tt.wantErr != "" {
				s.Require().ErrorContains(err, tt.wantErr)
			} else {
				s.Require().NoError(err)
			}

			if tt.checkValue != nil {
				tt.checkValue(s.T(), got)
				return
			}
			s.Require().Equal(tt.wantAny, got.Any())
		})
	}
}

func (s *RuntimeTestSuite) TestEvalCallJSBoundaryContracts() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	exec := &executorImpl{}

	s.Run("missing module alias reports module invocation error", func() {
		ec := NewExecutionContext(p, exec)
		expr := ast.NewCallExpression(
			ast.NewIdentifier("missing.fn", stubRange()),
			[]ast.Expression{ast.NewIntegerLiteral(1, stubRange())},
			false,
			nil,
			stubRange(),
		)

		got, _, err := eval(ctx, ec, exec, p, expr)
		s.Require().Error(err)
		s.Require().NotNil(got)
		s.Require().ErrorContains(err, "invoke module function failed")
	})

	s.Run("module call propagates runtime-side boundary error", func() {
		ec := NewExecutionContext(p, exec)
		ec.BindModule("mod", &ModuleBinding{Alias: "mod"})

		expr := ast.NewCallExpression(
			ast.NewIdentifier("mod.fn", stubRange()),
			[]ast.Expression{
				ast.NewMapLiteral([]ast.MapEntry{
					{
						Key:   ast.NewStringLiteral("k", stubRange()),
						Value: ast.NewIntegerLiteral(1, stubRange()),
					},
				}, stubRange()),
			},
			false,
			nil,
			stubRange(),
		)

		got, _, err := eval(ctx, ec, exec, p, expr)
		s.Require().Error(err)
		s.Require().NotNil(got)
		s.Require().ErrorContains(err, "failed to call function 'mod.fn'")
		s.Require().ErrorContains(err, "module has no JS binding")
	})
}

func (s *RuntimeTestSuite) TestEvalImportDispatchBoundaryFailure() {
	ctx := context.Background()
	p := &index.Policy{}
	exec := &executorImpl{}
	ec := NewExecutionContext(newEvalTestPolicy(), exec)

	imp := ast.NewImportClause(
		"allow",
		ast.NewFQN([]string{"policy_only"}, stubRange()).Ptr(),
		nil,
		stubRange(),
	)

	got, node, err := eval(ctx, ec, exec, p, imp)
	s.Require().NotNil(node)
	s.Require().Error(err)
	s.Require().True(got.IsNull())
	s.Require().ErrorContains(err, "import from must specify namespace/policy")
}
