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
	Params       []string
	ParamTypes   []TypeRef // parallel; nil or index beyond slice => untyped
	ParamOpts    []bool    // parallel; true => optional (?)
	ReturnType   TypeRef   // nil => untyped return
	Body         *BlockExpression
}

// NewLambdaExpression builds an untyped-parameter lambda (legacy helper).
func NewLambdaExpression(params []string, body *BlockExpression, ssp tokens.Range) *LambdaExpression {
	return NewLambdaExpressionFull(params, nil, nil, nil, body, ssp)
}

// NewLambdaExpressionFull builds a lambda with optional per-param types, optionality, and return type.
func NewLambdaExpressionFull(params []string, paramTypes []TypeRef, paramOpts []bool, returnType TypeRef, body *BlockExpression, ssp tokens.Range) *LambdaExpression {
	return &LambdaExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "lambda",
		},
		Params:     params,
		ParamTypes: paramTypes,
		ParamOpts:  paramOpts,
		ReturnType: returnType,
		Body:       body,
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
		if l.ParamOpts != nil && i < len(l.ParamOpts) && l.ParamOpts[i] {
			b.WriteByte('?')
		}
		if l.ParamTypes != nil && i < len(l.ParamTypes) && l.ParamTypes[i] != nil {
			b.WriteString(": ")
			b.WriteString(l.ParamTypes[i].String())
		}
	}
	b.WriteByte(')')
	if l.ReturnType != nil {
		b.WriteString(": ")
		b.WriteString(l.ReturnType.String())
	}
	b.WriteString(" => ")
	b.WriteString(l.Body.String())
	return b.String()
}

var _ Expression = &LambdaExpression{}
var _ Node = &LambdaExpression{}

// RequiredLambdaArity returns the count of non-optional parameters (runtime Precheck callable arity).
func RequiredLambdaArity(lam *LambdaExpression) int {
	if lam == nil {
		return 0
	}
	n := len(lam.Params)
	required := 0
	for i := 0; i < n; i++ {
		opt := lam.ParamOpts != nil && i < len(lam.ParamOpts) && lam.ParamOpts[i]
		if !opt {
			required++
		}
	}
	return required
}
