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
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestEvalUnaryNotAndBangUseTrinaryNegation(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		in       trinary.Value
		want     trinary.Value
	}{
		{name: "bang true to false", operator: "!", in: trinary.True, want: trinary.False},
		{name: "bang false to true", operator: "!", in: trinary.False, want: trinary.True},
		{name: "bang unknown remains unknown", operator: "!", in: trinary.Unknown, want: trinary.Unknown},
		{name: "not true to false", operator: "not", in: trinary.True, want: trinary.False},
		{name: "not false to true", operator: "not", in: trinary.False, want: trinary.True},
		{name: "not unknown remains unknown", operator: "not", in: trinary.Unknown, want: trinary.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ast.NewUnaryExpression(tt.operator, ast.NewTrinaryLiteral(tt.in, stubRange()), stubRange())
			got, _, err := evalUnary(t.Context(), NewExecutionContext(newEvalTestPolicy(), &executorImpl{}), &executorImpl{}, newEvalTestPolicy(), expr)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.Any())
		})
	}
}

func TestEvalUnaryNumberOperatorsAndErrors(t *testing.T) {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	plus, _, err := evalUnary(t.Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("+", ast.NewIntegerLiteral(7, stubRange()), stubRange()))
	require.NoError(t, err)
	require.Equal(t, 7.0, plus.Any())

	minus, _, err := evalUnary(t.Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("-", ast.NewIntegerLiteral(7, stubRange()), stubRange()))
	require.NoError(t, err)
	require.Equal(t, -7.0, minus.Any())

	_, _, err = evalUnary(t.Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("+", ast.NewStringLiteral("x", stubRange()), stubRange()))
	require.ErrorContains(t, err, "unary + requires number")

	_, _, err = evalUnary(t.Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("-", ast.NewStringLiteral("x", stubRange()), stubRange()))
	require.ErrorContains(t, err, "unary - requires number")
}

func TestEvalUnaryUndefinedPassthrough(t *testing.T) {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	missing := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())

	got, _, err := evalUnary(t.Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("!", missing, stubRange()))
	require.NoError(t, err)
	require.True(t, got.IsUndefined())
}
