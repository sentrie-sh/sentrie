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

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// FloatLiteral represents a float literal
type FloatLiteral struct {
	*baseNode
	Value float64
}

func NewFloatLiteral(value float64, ssp tokens.Range) *FloatLiteral {
	return &FloatLiteral{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "float_literal",
		},
		Value: value,
	}
}

func (f *FloatLiteral) String() string {
	return fmt.Sprintf("%g", f.Value)
}
func (f *FloatLiteral) expressionNode() {}

var _ Expression = &FloatLiteral{}
var _ Node = &FloatLiteral{}
