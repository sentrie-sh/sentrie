// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/suite"
)

type EvalDistinctTestSuite struct {
	suite.Suite
	ctx    context.Context
	ec     *ExecutionContext
	exec   *executorImpl
	policy *index.Policy
}

func (r *EvalDistinctTestSuite) SetupSuite() {
	r.ctx = context.Background()
	r.ec = &ExecutionContext{}
	r.exec = &executorImpl{}
	r.policy = &index.Policy{
		Namespace: &index.Namespace{
			FQN: ast.NewFQN([]string{"test", "namespace"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
	}
}

func (r *EvalDistinctTestSuite) TestEvalDistinctIntegersWithEquality() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(1), float64(2), float64(3)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctStringsWithEquality() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{"apple", "banana", "cherry"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctMixedTypes() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(1), "hello", "world"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctEmptyCollection() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctSingleItemCollection() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(42)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAllIdenticalItems() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(5)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAlreadyDistinctItems() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(1), float64(2), float64(3), float64(4)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctNonListInput() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.Error(err)
	r.Nil(result)
	r.Contains(err.Error(), "distinct expects list source")
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByAbsoluteValue() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(-1), float64(-2)}, result) // Should keep first occurrence of each absolute value
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByModulo3() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{float64(1), float64(2)}, result) // Should keep first occurrence of each modulo 3 result
}

func (r *EvalDistinctTestSuite) TestEvalDistinctPredicateEvaluationError() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.Error(err)
	r.Nil(result)
	r.Contains(err.Error(), "unable to resolve import")
}

func (r *EvalDistinctTestSuite) TestEvalDistinctWithMaps() {
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

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Len(result, 2) // Should have 2 distinct maps based on id field
}

func TestEvalDistinctTestSuite(t *testing.T) {
	suite.Run(t, new(EvalDistinctTestSuite))
}
