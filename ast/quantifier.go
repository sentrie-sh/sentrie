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

type FilterExpression struct {
	Range         tokens.Range
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type AnyExpression struct {
	Range         tokens.Range
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type AllExpression struct {
	Range         tokens.Range
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type FirstExpression struct {
	Range         tokens.Range
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type MapExpression struct {
	Range         tokens.Range
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Transform     Expression
}

type DistinctExpression struct {
	Range         tokens.Range
	Collection    Expression
	LeftIterator  string
	RightIterator string
	Predicate     Expression
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
	b.WriteString(a.ValueIterator)
	if a.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(a.IndexIterator)
	}
	b.WriteString(a.Predicate.String())
	return b.String()
}

func (a *AllExpression) String() string {
	b := strings.Builder{}
	b.WriteString("all ")
	b.WriteString(a.Collection.String())
	b.WriteString(" as ")
	b.WriteString(a.ValueIterator)
	if a.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(a.IndexIterator)
	}
	b.WriteString(a.Predicate.String())
	return b.String()
}

func (f *FirstExpression) String() string {
	b := strings.Builder{}
	b.WriteString("first ")
	b.WriteString(f.Collection.String())
	b.WriteString(" as ")
	b.WriteString(f.ValueIterator)
	if f.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(f.IndexIterator)
	}
	b.WriteString(f.Predicate.String())
	return b.String()
}

func (m *MapExpression) String() string {
	b := strings.Builder{}
	b.WriteString("any ")
	b.WriteString(m.Collection.String())
	b.WriteString(" as ")
	b.WriteString(m.ValueIterator)
	if m.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(m.IndexIterator)
	}
	b.WriteString(m.Transform.String())
	return b.String()
}

func (f *FilterExpression) String() string {
	b := strings.Builder{}
	b.WriteString("filter ")
	b.WriteString(f.Collection.String())
	b.WriteString(" as ")
	b.WriteString(f.ValueIterator)
	if f.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(f.IndexIterator)
	}
	b.WriteString(f.Predicate.String())
	return b.String()
}

func (d *DistinctExpression) String() string {
	b := strings.Builder{}
	b.WriteString("distinct ")
	b.WriteString(d.Collection.String())
	b.WriteString(" as ")
	b.WriteString(d.LeftIterator)
	b.WriteString(", ")
	b.WriteString(d.RightIterator)
	b.WriteString(d.Predicate.String())
	return b.String()
}

func (m *MapExpression) Span() tokens.Range {
	return m.Range
}
func (a *AnyExpression) Span() tokens.Range {
	return a.Range
}
func (a *AllExpression) Span() tokens.Range {
	return a.Range
}
func (f *FirstExpression) Span() tokens.Range {
	return f.Range
}
func (f *FilterExpression) Span() tokens.Range {
	return f.Range
}
func (d *DistinctExpression) Span() tokens.Range {
	return d.Range
}

func (m *MapExpression) expressionNode()      {}
func (f *FilterExpression) expressionNode()   {}
func (a *AnyExpression) expressionNode()      {}
func (a *AllExpression) expressionNode()      {}
func (f *FirstExpression) expressionNode()    {}
func (d *DistinctExpression) expressionNode() {}
