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
	"github.com/binaek/sentra/tokens"
)

type NamespaceStatement struct {
	Pos  tokens.Position
	Name FQN // Fully Qualified Name (FQN) of the namespace
}

func (n NamespaceStatement) String() string {
	return n.Name.String()
}

func (n NamespaceStatement) Position() tokens.Position {
	return n.Pos
}

func (n NamespaceStatement) statementNode() {}

type PolicyStatement struct {
	Pos        tokens.Position
	Name       string
	Statements []Statement
}

func (p PolicyStatement) String() string {
	return p.Name
}

func (p PolicyStatement) Position() tokens.Position {
	return p.Pos
}

func (p PolicyStatement) statementNode() {}

var _ Statement = &PolicyStatement{}

type Program struct {
	Statements []Statement
	Reference  string
}
