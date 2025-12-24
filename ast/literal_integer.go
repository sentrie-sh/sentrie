// SPDX-License-Identifier: Apache-2.0
//
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

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	*baseNode
	// under the hood, all values are floats
	Value float64
}

func NewIntegerLiteral(value int64, ssp tokens.Range) *IntegerLiteral {
	return &IntegerLiteral{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "integer_literal",
		},
		Value: float64(value),
	}
}

func (i *IntegerLiteral) String() string {
	return fmt.Sprintf("%g", i.Value)
}

func (i *IntegerLiteral) expressionNode() {}

var _ Expression = &IntegerLiteral{}
var _ Node = &IntegerLiteral{}
