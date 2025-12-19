// SPDX-License-Identifier: Apache-2.0

// Copyright 2025 Binaek Sarkar
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

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type IsDefinedExpression struct {
	*baseNode
	Left Expression
}

type IsEmptyExpression struct {
	*baseNode
	Left Expression
}

func NewIsDefinedExpression(left Expression, ssp tokens.Range) *IsDefinedExpression {
	return &IsDefinedExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "is_defined",
		},
		Left: left,
	}
}

func NewIsEmptyExpression(left Expression, ssp tokens.Range) *IsEmptyExpression {
	return &IsEmptyExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "is_empty",
		},
		Left: left,
	}
}

func (e *IsEmptyExpression) String() string {
	return "is empty " + e.Left.String()
}

func (e *IsDefinedExpression) String() string {
	return "is defined " + e.Left.String()
}

func (e *IsDefinedExpression) expressionNode() {}
func (e *IsEmptyExpression) expressionNode()   {}

var _ Expression = &IsDefinedExpression{}
var _ Node = &IsDefinedExpression{}
var _ Expression = &IsEmptyExpression{}
var _ Node = &IsEmptyExpression{}
