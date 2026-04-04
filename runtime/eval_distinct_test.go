// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE/2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *RuntimeTestSuite) TestEvalDistinctIntegersWithEquality() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(3, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(1), float64(2), float64(3)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctStringsWithEquality() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewStringLiteral("apple", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("banana", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("apple", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("cherry", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("banana", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{"apple", "banana", "cherry"}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctMixedTypes() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("hello", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("world", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewStringLiteral("hello", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(1), "hello", "world"}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctEmptyCollection() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctSingleItemCollection() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(42, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(42)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctAllIdenticalItems() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(5)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctAlreadyDistinctItems() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(3, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(4, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(1), float64(2), float64(3), float64(4)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctNonListInput() {
	// Test with non-list collection (should return error)
	distinctExpr := ast.NewDistinctExpression(
		ast.NewStringLiteral("not a list", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.Error(err)
	s.False(result.IsValid())
	s.Contains(err.Error(), "distinct expects list source")
}

func (s *RuntimeTestSuite) TestEvalDistinctByAbsoluteValue() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(-1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(-2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(-1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewInfixExpression(
				ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"*",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			ast.NewInfixExpression(
				ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"*",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(-1), float64(-2)}, result.Any()) // Should keep first occurrence of each absolute value
}

func (s *RuntimeTestSuite) TestEvalDistinctByModulo3() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(4, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(5, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewInfixExpression(
				ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				ast.NewIntegerLiteral(3, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"%",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			ast.NewInfixExpression(
				ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				ast.NewIntegerLiteral(3, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"%",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	s.Equal([]any{float64(1), float64(2)}, result.Any()) // Should keep first occurrence of each modulo 3 result
}

func (s *RuntimeTestSuite) TestEvalDistinctPredicateEvaluationError() {
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewCallExpression(
			ast.NewIdentifier("nonexistent_function", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			[]ast.Expression{
				ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
			},
			false,
			nil,
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.Error(err)
	s.False(result.IsValid())
	s.Contains(err.Error(), "unable to resolve import")
}

func (s *RuntimeTestSuite) TestEvalDistinctWithMaps() {
	// Test with map objects - compare by id field instead of direct equality
	collectionExpr := ast.NewListLiteral(
		[]ast.Expression{
			ast.NewMapLiteral(
				[]ast.MapEntry{
					{Key: ast.NewStringLiteral("id", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
					{Key: ast.NewStringLiteral("name", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewStringLiteral("Alice", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			ast.NewMapLiteral(
				[]ast.MapEntry{
					{Key: ast.NewStringLiteral("id", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewIntegerLiteral(2, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
					{Key: ast.NewStringLiteral("name", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewStringLiteral("Bob", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			ast.NewMapLiteral(
				[]ast.MapEntry{
					{Key: ast.NewStringLiteral("id", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewIntegerLiteral(1, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
					{Key: ast.NewStringLiteral("name", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}), Value: ast.NewStringLiteral("Alice", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})},
				},
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	distinctExpr := ast.NewDistinctExpression(
		collectionExpr,
		"left",
		"right",
		ast.NewInfixExpression(
			ast.NewFieldAccessExpression(
				ast.NewIdentifier("left", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"id",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			ast.NewFieldAccessExpression(
				ast.NewIdentifier("right", tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
				"id",
				tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			),
			"==",
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)

	result, _, err := evalDistinct(s.ctx, s.ec, s.exec, s.policy, distinctExpr)

	s.NoError(err)
	list, ok := result.ListValue()
	s.True(ok)
	s.Len(list, 2) // Should have 2 distinct maps based on id field
}
