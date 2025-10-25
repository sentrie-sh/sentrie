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
	"strings"

	"github.com/sentrie-sh/sentrie/tokens"
)

type ReduceExpression struct {
	Range         tokens.Range
	Collection    Expression
	From          Expression
	Accumulator   string
	ValueIterator string
	IndexIterator string // optional, may be ""
	Reducer       Expression
}

func (r *ReduceExpression) String() string {
	b := strings.Builder{}
	b.WriteString("reduce ")
	b.WriteString(r.Collection.String())
	b.WriteString(" from ")
	b.WriteString(r.From.String())
	b.WriteString(" as ")
	b.WriteString(r.ValueIterator)
	if r.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(r.IndexIterator)
	}
	b.WriteString(" { ")
	b.WriteString(r.Reducer.String())
	b.WriteString(" }")
	return b.String()
}

func (r *ReduceExpression) Span() tokens.Range {
	return r.Range
}

func (r *ReduceExpression) Kind() string {
	return "reduce"
}

func (r *ReduceExpression) expressionNode() {}

var _ Expression = &ReduceExpression{}
var _ Node = &ReduceExpression{}
