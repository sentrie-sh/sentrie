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

type quantifierExpression struct {
	*baseNode
	Collection Expression
	Iterator1  string
	Iterator2  string
	Quantifier Expression
}

type FilterExpression struct {
	*quantifierExpression
}

func NewFilterExpression(collection Expression, valueIterator string, indexIterator string, predicate Expression, ssp tokens.Range) *FilterExpression {
	return &FilterExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "filter",
			},
			Collection: collection,
			Iterator1:  valueIterator,
			Iterator2:  indexIterator,
			Quantifier: predicate,
		},
	}
}

type AnyExpression struct {
	*quantifierExpression
}

func NewAnyExpression(collection Expression, valueIterator string, indexIterator string, predicate Expression, ssp tokens.Range) *AnyExpression {
	return &AnyExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "any",
			},
			Collection: collection,
			Iterator1:  valueIterator,
			Iterator2:  indexIterator,
			Quantifier: predicate,
		},
	}
}

type AllExpression struct {
	*quantifierExpression
}

func NewAllExpression(collection Expression, valueIterator string, indexIterator string, predicate Expression, ssp tokens.Range) *AllExpression {
	return &AllExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "all",
			},
			Collection: collection,
			Iterator1:  valueIterator,
			Iterator2:  indexIterator,
			Quantifier: predicate,
		},
	}
}

type FirstExpression struct {
	*quantifierExpression
}

func NewFirstExpression(collection Expression, valueIterator string, indexIterator string, predicate Expression, ssp tokens.Range) *FirstExpression {
	return &FirstExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "first",
			},
			Collection: collection,
			Iterator1:  valueIterator,
			Iterator2:  indexIterator,
			Quantifier: predicate,
		},
	}
}

type MapExpression struct {
	*quantifierExpression
}

func NewMapExpression(collection Expression, valueIterator string, indexIterator string, transform Expression, ssp tokens.Range) *MapExpression {
	return &MapExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "map",
			},
			Collection: collection,
			Iterator1:  valueIterator,
			Iterator2:  indexIterator,
			Quantifier: transform,
		},
	}
}

type DistinctExpression struct {
	*quantifierExpression
}

func NewDistinctExpression(collection Expression, leftIterator string, rightIterator string, predicate Expression, ssp tokens.Range) *DistinctExpression {
	return &DistinctExpression{
		quantifierExpression: &quantifierExpression{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "distinct",
			},
			Collection: collection,
			Iterator1:  leftIterator,
			Iterator2:  rightIterator,
			Quantifier: predicate,
		},
	}
}

var _ Expression = &FilterExpression{}
var _ Node = &FilterExpression{}
var _ Expression = &AnyExpression{}
var _ Node = &AnyExpression{}
var _ Expression = &AllExpression{}
var _ Node = &AllExpression{}
var _ Expression = &FirstExpression{}
var _ Node = &FirstExpression{}
var _ Expression = &MapExpression{}
var _ Node = &MapExpression{}
var _ Expression = &DistinctExpression{}
var _ Node = &DistinctExpression{}

func (a *AnyExpression) String() string {
	b := strings.Builder{}
	b.WriteString("any ")
	b.WriteString(a.Collection.String())
	b.WriteString(" as ")
	b.WriteString(a.Iterator1)
	if a.Iterator2 != "" {
		b.WriteString(", ")
		b.WriteString(a.Iterator2)
	}
	b.WriteString(a.Quantifier.String())
	return b.String()
}

func (a *AllExpression) String() string {
	b := strings.Builder{}
	b.WriteString("all ")
	b.WriteString(a.Collection.String())
	b.WriteString(" as ")
	b.WriteString(a.Iterator1)
	if a.Iterator2 != "" {
		b.WriteString(", ")
		b.WriteString(a.Iterator2)
	}
	b.WriteString(a.Quantifier.String())
	return b.String()
}

func (f *FirstExpression) String() string {
	b := strings.Builder{}
	b.WriteString("first ")
	b.WriteString(f.Collection.String())
	b.WriteString(" as ")
	b.WriteString(f.Iterator1)
	if f.Iterator2 != "" {
		b.WriteString(", ")
		b.WriteString(f.Iterator2)
	}
	b.WriteString(f.Quantifier.String())
	return b.String()
}

func (m *MapExpression) String() string {
	b := strings.Builder{}
	b.WriteString("any ")
	b.WriteString(m.Collection.String())
	b.WriteString(" as ")
	b.WriteString(m.Iterator1)
	if m.Iterator2 != "" {
		b.WriteString(", ")
		b.WriteString(m.Iterator2)
	}
	b.WriteString(m.Quantifier.String())
	return b.String()
}

func (f *FilterExpression) String() string {
	b := strings.Builder{}
	b.WriteString("filter ")
	b.WriteString(f.Collection.String())
	b.WriteString(" as ")
	b.WriteString(f.Iterator1)
	if f.Iterator2 != "" {
		b.WriteString(", ")
		b.WriteString(f.Iterator2)
	}
	b.WriteString(f.Quantifier.String())
	return b.String()
}

func (d *DistinctExpression) String() string {
	b := strings.Builder{}
	b.WriteString("distinct ")
	b.WriteString(d.Collection.String())
	b.WriteString(" as ")
	b.WriteString(d.Iterator1)
	b.WriteString(", ")
	b.WriteString(d.Iterator2)
	b.WriteString(d.Quantifier.String())
	return b.String()
}

func (m *MapExpression) expressionNode()      {}
func (f *FilterExpression) expressionNode()   {}
func (a *AnyExpression) expressionNode()      {}
func (a *AllExpression) expressionNode()      {}
func (f *FirstExpression) expressionNode()    {}
func (d *DistinctExpression) expressionNode() {}
