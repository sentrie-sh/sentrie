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
	"math"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func stubRange() tokens.Range {
	return tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}
}

func (s *RuntimeTestSuite) TestCalculateHashKeyDistinguishesUndefinedAndNull() {
	node := &ast.CallExpression{}
	undefinedHash := calculateHashKey(node, []box.Value{box.Undefined()})
	nullHash := calculateHashKey(node, []box.Value{box.Null()})

	s.Require().NotEmpty(undefinedHash)
	s.Require().NotEmpty(nullHash)
	s.Require().NotEqual(undefinedHash, nullHash)
}

func (s *RuntimeTestSuite) TestGetTargetBuiltinPreservesUndefined() {
	ec := NewExecutionContext(&index.Policy{}, &executorImpl{})
	call := ast.NewCallExpression(
		ast.NewIdentifier("as_list", stubRange()),
		[]ast.Expression{},
		false,
		nil,
		stubRange(),
	)

	target, err := getTarget(context.Background(), ec, &index.Policy{}, call)
	s.Require().NoError(err)

	out, err := target(context.Background(), box.Undefined())
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestGetTargetBuiltinPreservesNestedUndefined() {
	ec := NewExecutionContext(&index.Policy{}, &executorImpl{})
	call := ast.NewCallExpression(
		ast.NewIdentifier("flatten_deep", stubRange()),
		[]ast.Expression{},
		false,
		nil,
		stubRange(),
	)

	target, err := getTarget(context.Background(), ec, &index.Policy{}, call)
	s.Require().NoError(err)

	arg := box.List([]box.Value{
		box.List([]box.Value{
			box.Number(1),
			box.Undefined(),
		}),
	})
	out, err := target(context.Background(), arg)
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestCalculateHashKeyMapKeyOrderStable() {
	node := &ast.CallExpression{}
	arg1 := box.Map(map[string]box.Value{"a": box.Number(1), "b": box.Number(2)})
	arg2 := box.Map(map[string]box.Value{"b": box.Number(2), "a": box.Number(1)})
	hash1 := calculateHashKey(node, []box.Value{arg1})
	hash2 := calculateHashKey(node, []box.Value{arg2})
	s.Require().Equal(hash1, hash2)
}

func (s *RuntimeTestSuite) TestCalculateHashKeyNestedStructureStable() {
	node := &ast.CallExpression{}
	arg := box.List([]box.Value{
		box.Map(map[string]box.Value{"k": box.List([]box.Value{box.Number(1), box.String("x")})}),
	})
	hash := calculateHashKey(node, []box.Value{arg})
	s.Require().NotEmpty(hash)
}

func (s *RuntimeTestSuite) TestCalculateHashKeyNumericEdges() {
	node := &ast.CallExpression{}
	hashNegZero := calculateHashKey(node, []box.Value{box.Number(math.Copysign(0, -1))})
	hashPosZero := calculateHashKey(node, []box.Value{box.Number(0)})
	hashNaN := calculateHashKey(node, []box.Value{box.Number(math.NaN())})
	hashInf := calculateHashKey(node, []box.Value{box.Number(math.Inf(1))})

	s.Require().NotEmpty(hashNaN)
	s.Require().NotEmpty(hashInf)
	s.Require().NotEqual(hashNegZero, hashPosZero)
}
