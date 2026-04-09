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

package ast

import (
	"strings"

	"github.com/sentrie-sh/sentrie/tokens"
)

// LambdaExpression is an inline block-bodied lambda: (a, b) => { yield ... }
type LambdaExpression struct {
	*baseNode
	Params []string
	Body   *BlockExpression
}

func NewLambdaExpression(params []string, body *BlockExpression, ssp tokens.Range) *LambdaExpression {
	return &LambdaExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "lambda",
		},
		Params: params,
		Body:   body,
	}
}

func (l *LambdaExpression) expressionNode() {}

func (l *LambdaExpression) String() string {
	var b strings.Builder
	b.WriteByte('(')
	for i, p := range l.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p)
	}
	b.WriteString(") => ")
	b.WriteString(l.Body.String())
	return b.String()
}

var _ Expression = &LambdaExpression{}
var _ Node = &LambdaExpression{}
