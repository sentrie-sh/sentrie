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
	Pos           tokens.Position
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type AnyExpression struct {
	Pos           tokens.Position
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type AllExpression struct {
	Pos           tokens.Position
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type MapExpression struct {
	Pos           tokens.Position
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Transform     Expression
}

type DistinctExpression struct {
	Pos           tokens.Position
	Collection    Expression
	ValueIterator string
	IndexIterator string
	Predicate     Expression
}

type CountExpression struct {
	Pos        tokens.Position
	Collection Expression
}

var _ Expression = &FilterExpression{}
var _ Expression = &AnyExpression{}
var _ Expression = &AllExpression{}
var _ Expression = &MapExpression{}
var _ Expression = &DistinctExpression{}
var _ Expression = &CountExpression{}

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
	b.WriteString("{ ")
	b.WriteString(a.Predicate.String())
	b.WriteString(" }")
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
	b.WriteString("{ ")
	b.WriteString(a.Predicate.String())
	b.WriteString(" }")
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
	b.WriteString("{ ")
	b.WriteString(m.Transform.String())
	b.WriteString(" }")
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
	b.WriteString("{ ")
	b.WriteString(f.Predicate.String())
	b.WriteString(" }")
	return b.String()
}
func (d *DistinctExpression) String() string {
	b := strings.Builder{}
	b.WriteString("any ")
	b.WriteString(d.Collection.String())
	b.WriteString(" as ")
	b.WriteString(d.ValueIterator)
	if d.IndexIterator != "" {
		b.WriteString(", ")
		b.WriteString(d.IndexIterator)
	}
	b.WriteString("{ ")
	b.WriteString(d.Predicate.String())
	b.WriteString(" }")
	return b.String()
}

func (c *CountExpression) String() string {
	b := strings.Builder{}
	b.WriteString("count ")
	b.WriteString(c.Collection.String())
	return b.String()
}

func (m *MapExpression) Position() tokens.Position {
	return m.Pos
}
func (a *AnyExpression) Position() tokens.Position {
	return a.Pos
}
func (a *AllExpression) Position() tokens.Position {
	return a.Pos
}
func (f *FilterExpression) Position() tokens.Position {
	return f.Pos
}
func (d *DistinctExpression) Position() tokens.Position {
	return d.Pos
}
func (c *CountExpression) Position() tokens.Position {
	return c.Pos
}

func (m *MapExpression) expressionNode()      {}
func (f *FilterExpression) expressionNode()   {}
func (a *AnyExpression) expressionNode()      {}
func (a *AllExpression) expressionNode()      {}
func (d *DistinctExpression) expressionNode() {}
func (c *CountExpression) expressionNode()    {}
