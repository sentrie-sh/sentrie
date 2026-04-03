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
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestMatchesValueTypeErrorsAndRegexBranches(t *testing.T) {
	_, err := box.MatchesValue(box.Number(1), box.String("a+"))
	require.ErrorContains(t, err, "haystack must be a string")

	_, err = box.MatchesValue(box.String("abc"), box.Number(1))
	require.ErrorContains(t, err, "pattern must be a string")

	ok, err := box.MatchesValue(box.String("abc123"), box.String("^[a-z]+\\d+$"))
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = box.MatchesValue(box.String("abc"), box.String("^\\d+$"))
	require.NoError(t, err)
	require.False(t, ok)

	_, err = box.MatchesValue(box.String("abc"), box.String("["))
	require.Error(t, err)
}

func TestContainsValueStringListAndMapBranches(t *testing.T) {
	require.True(t, box.ContainsValue(box.String("sentrie runtime"), box.String("runtime")))
	require.False(t, box.ContainsValue(box.String("sentrie"), box.String("missing")))
	require.False(t, box.ContainsValue(box.String("sentrie"), box.String("")))
	require.False(t, box.ContainsValue(box.String("sentrie"), box.Number(1)))

	require.True(t, box.ContainsValue(box.List([]box.Value{box.Number(1), box.String("x")}), box.String("x")))
	require.False(t, box.ContainsValue(box.List([]box.Value{box.Number(1), box.String("x")}), box.String("y")))

	haystack := box.Map(map[string]box.Value{
		"id":   box.Number(7),
		"name": box.String("alice"),
		"meta": box.Map(map[string]box.Value{"active": box.Bool(true)}),
	})

	require.True(t, box.ContainsValue(haystack, box.String("name")))
	require.False(t, box.ContainsValue(haystack, box.String("missing")))

	require.True(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
		"id": box.Number(7),
	})))
	require.False(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
		"id": box.Number(8),
	})))
	require.False(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
		"id":      box.Number(7),
		"missing": box.Number(1),
	})))

	require.False(t, box.ContainsValue(haystack, box.String("alice")))
	require.False(t, box.ContainsValue(haystack, box.String("bob")))
	require.False(t, box.ContainsValue(haystack, box.Number(7)))
	require.False(t, box.ContainsValue(haystack, box.Number(99)))
	require.False(t, box.ContainsValue(box.Number(1), box.Number(1)))
}

func TestEqualValuesDeepAndKindSensitiveBranches(t *testing.T) {
	require.True(t, box.EqualValues(box.Undefined(), box.Undefined()))
	require.True(t, box.EqualValues(box.Null(), box.Null()))
	require.False(t, box.EqualValues(box.Undefined(), box.Null()))

	require.True(t, box.EqualValues(box.Bool(true), box.Bool(true)))
	require.False(t, box.EqualValues(box.Bool(true), box.Bool(false)))
	require.True(t, box.EqualValues(box.Number(1.5), box.Number(1.5)))
	require.False(t, box.EqualValues(box.Number(1.5), box.Number(2)))
	require.True(t, box.EqualValues(box.String("x"), box.String("x")))
	require.False(t, box.EqualValues(box.String("x"), box.String("y")))
	require.True(t, box.EqualValues(box.Trinary(trinary.Unknown), box.Trinary(trinary.Unknown)))
	require.False(t, box.EqualValues(box.Trinary(trinary.True), box.Trinary(trinary.False)))

	require.True(t, box.EqualValues(
		box.List([]box.Value{box.Number(1), box.Map(map[string]box.Value{"k": box.String("v")})}),
		box.List([]box.Value{box.Number(1), box.Map(map[string]box.Value{"k": box.String("v")})}),
	))
	require.False(t, box.EqualValues(
		box.List([]box.Value{box.Number(1)}),
		box.List([]box.Value{box.Number(1), box.Number(2)}),
	))
	require.False(t, box.EqualValues(
		box.List([]box.Value{box.Number(1), box.Number(2)}),
		box.List([]box.Value{box.Number(1), box.Number(3)}),
	))

	require.True(t, box.EqualValues(
		box.Map(map[string]box.Value{"a": box.Number(1), "b": box.String("x")}),
		box.Map(map[string]box.Value{"b": box.String("x"), "a": box.Number(1)}),
	))
	require.False(t, box.EqualValues(
		box.Map(map[string]box.Value{"a": box.Number(1)}),
		box.Map(map[string]box.Value{"a": box.Number(1), "b": box.Number(2)}),
	))
	require.False(t, box.EqualValues(
		box.Map(map[string]box.Value{"a": box.Number(1)}),
		box.Map(map[string]box.Value{"a": box.Number(2)}),
	))
	require.False(t, box.EqualValues(
		box.Map(map[string]box.Value{"a": box.Number(1)}),
		box.Map(map[string]box.Value{"b": box.Number(1)}),
	))

	shared := &struct{ Name string }{Name: "same"}
	require.True(t, box.EqualValues(box.Object(shared), box.Object(shared)))
	require.False(t, box.EqualValues(
		box.Object(&struct{ Name string }{Name: "same"}),
		box.Object(&struct{ Name string }{Name: "same"}),
	))

	require.False(t, box.EqualValues(box.List([]box.Value{}), box.Map(map[string]box.Value{})))
}

func TestEvalInfixArithmeticComparisonAndTrinaryMatrix(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			got, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantIsUndef {
				require.True(t, got.IsUndefined())
				return
			}
			require.Equal(t, tt.want, got.Any())
		})
	}
}

func TestEvalInfixOperatorSpecificErrorBranches(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			_, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestEvalInfixMembershipAndComparisonAliases(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			expr := ast.NewInfixExpression(tt.left, tt.right, tt.operator, stubRange())
			got, _, err := evalInfix(ctx, ec, &executorImpl{}, p, expr)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.Any())
		})
	}
}
