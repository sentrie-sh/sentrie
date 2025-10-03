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
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 3},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2), int64(3)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctStringsWithEquality() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "apple"},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "banana"},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "apple"},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "cherry"},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "banana"},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{"apple", "banana", "cherry"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctMixedTypes() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "hello"},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "world"},
			&ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "hello"},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), "hello", "world"}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctEmptyCollection() {
	collectionExpr := &ast.ListLiteral{
		Pos:    tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctSingleItemCollection() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 42},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(42)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAllIdenticalItems() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 5},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 5},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 5},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 5},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(5)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctAlreadyDistinctItems() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 3},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 4},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2), int64(3), int64(4)}, result)
}

func (r *EvalDistinctTestSuite) TestEvalDistinctNonListInput() {
	// Test with non-list collection (should return error)
	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    &ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "not a list"},
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos:      tokens.Position{Line: 1, Column: 1},
			Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			Operator: "==",
			Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.Error(err)
	r.Nil(result)
	r.Contains(err.Error(), "distinct expects list source")
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByAbsoluteValue() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: -1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: -2},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: -1},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos: tokens.Position{Line: 1, Column: 1},
			Left: &ast.InfixExpression{
				Pos:      tokens.Position{Line: 1, Column: 1},
				Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
				Operator: "*",
				Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
			},
			Operator: "==",
			Right: &ast.InfixExpression{
				Pos:      tokens.Position{Line: 1, Column: 1},
				Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
				Operator: "*",
				Right:    &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(-1), int64(-2)}, result) // Should keep first occurrence of each absolute value
}

func (r *EvalDistinctTestSuite) TestEvalDistinctByModulo3() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 4},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 5},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos: tokens.Position{Line: 1, Column: 1},
			Left: &ast.InfixExpression{
				Pos:      tokens.Position{Line: 1, Column: 1},
				Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
				Operator: "%",
				Right:    &ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 3},
			},
			Operator: "==",
			Right: &ast.InfixExpression{
				Pos:      tokens.Position{Line: 1, Column: 1},
				Left:     &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
				Operator: "%",
				Right:    &ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 3},
			},
		},
	}

	result, _, err := evalDistinct(r.ctx, r.ec, r.exec, r.policy, distinctExpr)

	r.NoError(err)
	r.Equal([]any{int64(1), int64(2)}, result) // Should keep first occurrence of each modulo 3 result
}

func (r *EvalDistinctTestSuite) TestEvalDistinctPredicateEvaluationError() {
	collectionExpr := &ast.ListLiteral{
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1},
			&ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.CallExpression{
			Pos:    tokens.Position{Line: 1, Column: 1},
			Callee: &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "nonexistent_function"},
			Arguments: []ast.Expression{
				&ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
				&ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
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
		Pos: tokens.Position{Line: 1, Column: 1},
		Values: []ast.Expression{
			&ast.MapLiteral{
				Pos: tokens.Position{Line: 1, Column: 1},
				Entries: []ast.MapEntry{
					{Key: "id", Value: &ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1}},
					{Key: "name", Value: &ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "Alice"}},
				},
			},
			&ast.MapLiteral{
				Pos: tokens.Position{Line: 1, Column: 1},
				Entries: []ast.MapEntry{
					{Key: "id", Value: &ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 2}},
					{Key: "name", Value: &ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "Bob"}},
				},
			},
			&ast.MapLiteral{
				Pos: tokens.Position{Line: 1, Column: 1},
				Entries: []ast.MapEntry{
					{Key: "id", Value: &ast.IntegerLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: 1}},
					{Key: "name", Value: &ast.StringLiteral{Pos: tokens.Position{Line: 1, Column: 1}, Value: "Alice"}},
				},
			},
		},
	}

	distinctExpr := &ast.DistinctExpression{
		Pos:           tokens.Position{Line: 1, Column: 1},
		Collection:    collectionExpr,
		LeftIterator:  "left",
		RightIterator: "right",
		Predicate: &ast.InfixExpression{
			Pos: tokens.Position{Line: 1, Column: 1},
			Left: &ast.FieldAccessExpression{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Left:  &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "left"},
				Field: "id",
			},
			Operator: "==",
			Right: &ast.FieldAccessExpression{
				Pos:   tokens.Position{Line: 1, Column: 1},
				Left:  &ast.Identifier{Pos: tokens.Position{Line: 1, Column: 1}, Value: "right"},
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
