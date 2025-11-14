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
)

type Positionable interface {
	Span() tokens.Range
}

type Node interface {
	Positionable
	String() string
	Kind() string
}

type baseNode struct {
	Rnge  tokens.Range
	Kind_ string
}

func (n *baseNode) Span() tokens.Range {
	return n.Rnge
}

func (n *baseNode) Kind() string {
	return n.Kind_
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}
