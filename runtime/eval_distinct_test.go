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
			FQN: ast.FQN{"test", "namespace"},
		},
	}
}

func (r *EvalDistinctTestSuite) TestEvalDistinctIntegersWithEquality() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 3},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2), int64(3)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctStringsWithEquality() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "apple"},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "banana"},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "apple"},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "cherry"},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "banana"},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{"apple", "banana", "cherry"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctMixedTypes() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "hello"},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "world"},
			&ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "hello"},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), "hello", "world"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctEmptyCollection() {
	collectionExpr := &ast.ListLiteral{
		Range:    tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctSingleItemCollection() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 42},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(42)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAllIdenticalItems() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 5},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 5},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 5},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 5},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(5)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAlreadyDistinctItems() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 3},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 4},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2), int64(3), int64(4)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctNonListInput() {
	// Test with non-list collection (should return error)
	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "not a list"},
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.Error(err)
	r.Nil(result)
	r.Contains(err.Error(), "distinct expects list source")
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByAbsoluteValue() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: -1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: -2},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: -1},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left: &ast.InfixExpression{
				Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
				Operator: "*",
				Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
			},
			Operator: "==",
			Right: &ast.InfixExpression{
				Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
				Operator: "*",
				Right:    &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(-1), int64(-2)}, result) // Should keep first occurrence of each absolute value
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByModulo3() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 4},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 5},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left: &ast.InfixExpression{
				Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
				Operator: "%",
				Right:    &ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 3},
			},
			Operator: "==",
			Right: &ast.InfixExpression{
				Range:      tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:     &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
				Operator: "%",
				Right:    &ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 3},
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2)}, result) // Should keep first occurrence of each modulo 3 result
}

func (r *EvalDistinctTestSuite) TestEvalDistinctPredicateEvaluationError() {
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1},
			&ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.CallExpression{
			Range:    tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Callee: &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "nonexistent_function"},
			Arguments: []ast.Expression{
				&ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
				&ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.Error(err)
	r.Nil(result)
	r.Contains(err.Error(), "unable to resolve import")
}

func (r *EvalDistinctTestSuite) TestEvalDistinctWithMaps() {
	// Test with map objects - compare by id field instead of direct equality
	collectionExpr := &ast.ListLiteral{
		Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Values: []ast.Expression{
			&ast.MapLiteral{
				Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Entries: []ast.MapEntry{
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "id"}, Value: &ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1}},
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "name"}, Value: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "Alice"}},
				},
			},
			&ast.MapLiteral{
				Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Entries: []ast.MapEntry{
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "id"}, Value: &ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 2}},
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "name"}, Value: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "Bob"}},
				},
			},
			&ast.MapLiteral{
				Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Entries: []ast.MapEntry{
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "id"}, Value: &ast.IntegerLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: 1}},
					{Key: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "name"}, Value: &ast.StringLiteral{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "Alice"}},
				},
			},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Range:           tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
			Left: &ast.FieldAccessExpression{
				Range:   tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:  &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "left"},
				Field: "id",
			},
			Operator: "==",
			Right: &ast.FieldAccessExpression{
				Range:   tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
				Left:  &ast.Identifier{Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}, Value: "right"},
				Field: "id",
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Len(result, 2) // Should have 2 distinct maps based on id field
}

func TestEvalDistinctTestSuite(t *testing.T) {
	suite.Run(t, new(EvalDistinctTestSuite))
}
