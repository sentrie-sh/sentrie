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
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestMatchesValueTypeErrorsAndRegexBranches() {
	_, err := box.MatchesValue(box.Number(1), box.String("a+"))
	s.Require().ErrorContains(err, "haystack must be a string")

	_, err = box.MatchesValue(box.String("abc"), box.Number(1))
	s.Require().ErrorContains(err, "pattern must be a string")

	ok, err := box.MatchesValue(box.String("abc123"), box.String("^[a-z]+\\d+$"))
	s.Require().NoError(err)
	s.Require().True(ok)

	ok, err = box.MatchesValue(box.String("abc"), box.String("^\\d+$"))
	s.Require().NoError(err)
	s.Require().False(ok)

	_, err = box.MatchesValue(box.String("abc"), box.String("["))
	s.Require().Error(err)
}

func (s *RuntimeTestSuite) TestContainsValueStringListAndMapBranches() {
	s.Require().True(box.ContainsValue(box.String("sentrie runtime"), box.String("runtime")))
	s.Require().False(box.ContainsValue(box.String("sentrie"), box.String("missing")))
	s.Require().False(box.ContainsValue(box.String("sentrie"), box.String("")))
	s.Require().False(box.ContainsValue(box.String("sentrie"), box.Number(1)))

	s.Require().True(box.ContainsValue(box.List([]box.Value{box.Number(1), box.String("x")}), box.String("x")))
	s.Require().False(box.ContainsValue(box.List([]box.Value{box.Number(1), box.String("x")}), box.String("y")))

	haystack := box.Dict(map[string]box.Value{
		"id":   box.Number(7),
		"name": box.String("alice"),
		"meta": box.Dict(map[string]box.Value{"active": box.Bool(true)}),
	})

	s.Require().True(box.ContainsValue(haystack, box.String("name")))
	s.Require().False(box.ContainsValue(haystack, box.String("missing")))

	s.Require().True(box.ContainsValue(haystack, box.Dict(map[string]box.Value{
		"id": box.Number(7),
	})))
	s.Require().False(box.ContainsValue(haystack, box.Dict(map[string]box.Value{
		"id": box.Number(8),
	})))
	s.Require().False(box.ContainsValue(haystack, box.Dict(map[string]box.Value{
		"id":      box.Number(7),
		"missing": box.Number(1),
	})))

	s.Require().False(box.ContainsValue(haystack, box.String("alice")))
	s.Require().False(box.ContainsValue(haystack, box.String("bob")))
	s.Require().False(box.ContainsValue(haystack, box.Number(7)))
	s.Require().False(box.ContainsValue(haystack, box.Number(99)))
	s.Require().False(box.ContainsValue(box.Number(1), box.Number(1)))
}

func (s *RuntimeTestSuite) TestEqualValuesDeepAndKindSensitiveBranches() {
	s.Require().True(box.EqualValues(box.Undefined(), box.Undefined()))
	s.Require().True(box.EqualValues(box.Null(), box.Null()))
	s.Require().False(box.EqualValues(box.Undefined(), box.Null()))

	s.Require().True(box.EqualValues(box.Bool(true), box.Bool(true)))
	s.Require().False(box.EqualValues(box.Bool(true), box.Bool(false)))
	s.Require().True(box.EqualValues(box.Number(1.5), box.Number(1.5)))
	s.Require().False(box.EqualValues(box.Number(1.5), box.Number(2)))
	s.Require().True(box.EqualValues(box.String("x"), box.String("x")))
	s.Require().False(box.EqualValues(box.String("x"), box.String("y")))
	s.Require().True(box.EqualValues(box.Trinary(trinary.Unknown), box.Trinary(trinary.Unknown)))
	s.Require().False(box.EqualValues(box.Trinary(trinary.True), box.Trinary(trinary.False)))

	s.Require().True(box.EqualValues(
		box.List([]box.Value{box.Number(1), box.Dict(map[string]box.Value{"k": box.String("v")})}),
		box.List([]box.Value{box.Number(1), box.Dict(map[string]box.Value{"k": box.String("v")})}),
	))
	s.Require().False(box.EqualValues(
		box.List([]box.Value{box.Number(1)}),
		box.List([]box.Value{box.Number(1), box.Number(2)}),
	))
	s.Require().False(box.EqualValues(
		box.List([]box.Value{box.Number(1), box.Number(2)}),
		box.List([]box.Value{box.Number(1), box.Number(3)}),
	))

	s.Require().True(box.EqualValues(
		box.Dict(map[string]box.Value{"a": box.Number(1), "b": box.String("x")}),
		box.Dict(map[string]box.Value{"b": box.String("x"), "a": box.Number(1)}),
	))
	s.Require().False(box.EqualValues(
		box.Dict(map[string]box.Value{"a": box.Number(1)}),
		box.Dict(map[string]box.Value{"a": box.Number(1), "b": box.Number(2)}),
	))
	s.Require().False(box.EqualValues(
		box.Dict(map[string]box.Value{"a": box.Number(1)}),
		box.Dict(map[string]box.Value{"a": box.Number(2)}),
	))
	s.Require().False(box.EqualValues(
		box.Dict(map[string]box.Value{"a": box.Number(1)}),
		box.Dict(map[string]box.Value{"b": box.Number(1)}),
	))

	shared := &struct{ Name string }{Name: "same"}
	s.Require().True(box.EqualValues(box.Object(shared), box.Object(shared)))
	s.Require().False(box.EqualValues(
		box.Object(&struct{ Name string }{Name: "same"}),
		box.Object(&struct{ Name string }{Name: "same"}),
	))

	s.Require().False(box.EqualValues(box.List([]box.Value{}), box.Dict(map[string]box.Value{})))
}

func (s *RuntimeTestSuite) TestEvalInfixArithmeticComparisonAndTrinaryMatrix() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	tests := []struct {
		name        string
		operator    string
		left        ast.Expression
		right       ast.Expression
		want        any
		wantErr     string
		wantIsUndef bool
	}{
		{
			name:     "plus concatenates when left is string",
			operator: "+",
			left:     ast.NewStringLiteral("x=", stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     "x=2",
		},
		{
			name:     "plus concatenates when right is string",
			operator: "+",
			left:     ast.NewIntegerLiteral(2, stubRange()),
			right:    ast.NewStringLiteral(" apples", stubRange()),
			want:     "2 apples",
		},
		{
			name:     "minus numbers",
			operator: "-",
			left:     ast.NewIntegerLiteral(8, stubRange()),
			right:    ast.NewIntegerLiteral(3, stubRange()),
			want:     5.0,
		},
		{
			name:     "multiply numbers",
			operator: "*",
			left:     ast.NewIntegerLiteral(3, stubRange()),
			right:    ast.NewIntegerLiteral(4, stubRange()),
			want:     12.0,
		},
		{
			name:     "divide numbers",
			operator: "/",
			left:     ast.NewIntegerLiteral(8, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     4.0,
		},
		{
			name:     "mod numbers",
			operator: "%",
			left:     ast.NewIntegerLiteral(8, stubRange()),
			right:    ast.NewIntegerLiteral(3, stubRange()),
			want:     2.0,
		},
		{
			name:     "less than",
			operator: "<",
			left:     ast.NewIntegerLiteral(1, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     true,
		},
		{
			name:     "less than or equal",
			operator: "<=",
			left:     ast.NewIntegerLiteral(2, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     true,
		},
		{
			name:     "greater than",
			operator: ">",
			left:     ast.NewIntegerLiteral(3, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     true,
		},
		{
			name:     "greater than or equal",
			operator: ">=",
			left:     ast.NewIntegerLiteral(2, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			want:     true,
		},
		{
			name:     "and trinary unknown and true",
			operator: "and",
			left:     ast.NewTrinaryLiteral(trinary.Unknown, stubRange()),
			right:    ast.NewTrinaryLiteral(trinary.True, stubRange()),
			want:     trinary.Unknown,
		},
		{
			name:     "or trinary false and unknown",
			operator: "or",
			left:     ast.NewTrinaryLiteral(trinary.False, stubRange()),
			right:    ast.NewTrinaryLiteral(trinary.Unknown, stubRange()),
			want:     trinary.Unknown,
		},
		{
			name:     "xor trinary true xor true",
			operator: "xor",
			left:     ast.NewTrinaryLiteral(trinary.True, stubRange()),
			right:    ast.NewTrinaryLiteral(trinary.True, stubRange()),
			want:     trinary.False,
		},
		{
			name:        "undefined short-circuits before operator logic",
			operator:    "+",
			left:        ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange()),
			right:       ast.NewIntegerLiteral(1, stubRange()),
			wantIsUndef: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			got, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			if tt.wantErr != "" {
				s.Require().ErrorContains(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			if tt.wantIsUndef {
				s.Require().True(got.IsUndefined())
				return
			}
			s.Require().Equal(tt.want, got.Any())
		})
	}
}

func (s *RuntimeTestSuite) TestEvalInfixOperatorSpecificErrorBranches() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	tests := []struct {
		name     string
		operator string
		left     ast.Expression
		right    ast.Expression
		wantErr  string
	}{
		{
			name:     "divide by zero errors",
			operator: "/",
			left:     ast.NewIntegerLiteral(8, stubRange()),
			right:    ast.NewIntegerLiteral(0, stubRange()),
			wantErr:  "divide by zero",
		},
		{
			name:     "mod by zero errors",
			operator: "%",
			left:     ast.NewIntegerLiteral(8, stubRange()),
			right:    ast.NewIntegerLiteral(0, stubRange()),
			wantErr:  "divide by zero",
		},
		{
			name:     "plus numeric path rejects non numeric left",
			operator: "+",
			left:     ast.NewTrinaryLiteral(trinary.True, stubRange()),
			right:    ast.NewIntegerLiteral(1, stubRange()),
			wantErr:  "left operand is not a number",
		},
		{
			name:     "comparison rejects non numeric right",
			operator: "<",
			left:     ast.NewIntegerLiteral(1, stubRange()),
			right:    ast.NewStringLiteral("x", stubRange()),
			wantErr:  "right operand is not a number",
		},
		{
			name:     "matches operator reports haystack mismatch",
			operator: "matches",
			left:     ast.NewIntegerLiteral(1, stubRange()),
			right:    ast.NewStringLiteral("^\\d+$", stubRange()),
			wantErr:  "haystack must be a string",
		},
		{
			name:     "unsupported operator branch",
			operator: "<>",
			left:     ast.NewIntegerLiteral(1, stubRange()),
			right:    ast.NewIntegerLiteral(2, stubRange()),
			wantErr:  "unsupported infix op",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			_, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			s.Require().ErrorContains(err, tt.wantErr)
		})
	}
}

func (s *RuntimeTestSuite) TestEvalInfixMembershipAndComparisonAliases() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	tests := []struct {
		name     string
		operator string
		left     ast.Expression
		right    ast.Expression
		want     any
	}{
		{
			name:     "is alias uses equality",
			operator: "is",
			left:     ast.NewIntegerLiteral(4, stubRange()),
			right:    ast.NewIntegerLiteral(4, stubRange()),
			want:     true,
		},
		{
			name:     "not equals branch",
			operator: "!=",
			left:     ast.NewIntegerLiteral(4, stubRange()),
			right:    ast.NewIntegerLiteral(5, stubRange()),
			want:     true,
		},
		{
			name:     "in operator over list haystack",
			operator: "in",
			left:     ast.NewIntegerLiteral(2, stubRange()),
			right: ast.NewListLiteral([]ast.Expression{
				ast.NewIntegerLiteral(1, stubRange()),
				ast.NewIntegerLiteral(2, stubRange()),
			}, stubRange()),
			want: true,
		},
		{
			name:     "contains over string haystack",
			operator: "contains",
			left:     ast.NewStringLiteral("sentrie", stubRange()),
			right:    ast.NewStringLiteral("trie", stubRange()),
			want:     true,
		},
		{
			name:     "matches success branch",
			operator: "matches",
			left:     ast.NewStringLiteral("abc123", stubRange()),
			right:    ast.NewStringLiteral("^[a-z]+\\d+$", stubRange()),
			want:     true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			got, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			s.Require().NoError(err)
			s.Require().Equal(tt.want, got.Any())
		})
	}
}
