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
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

// TrinaryLiteral represents a trinary literal
type TrinaryLiteral struct {
	*baseNode
	Value trinary.Value
}

func NewTrinaryLiteral(value trinary.Value, ssp tokens.Range) *TrinaryLiteral {
	return &TrinaryLiteral{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "trinary_literal",
		},
		Value: value,
	}
}

func (b *TrinaryLiteral) String() string {
	return b.Value.String()
}

func (b *TrinaryLiteral) expressionNode() {}

var _ Expression = &TrinaryLiteral{}
var _ Node = &TrinaryLiteral{}
