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
	"github.com/sentrie-sh/sentrie/tokens"
)

type NamespaceStatement struct {
	*baseNode
	Name FQN // Fully Qualified Name (FQN) of the namespace
}

func NewNamespaceStatement(name FQN, ssp tokens.Range) *NamespaceStatement {
	return &NamespaceStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "namespace",
		},
		Name: name,
	}
}
func (n NamespaceStatement) String() string {
	return n.Name.String()
}

func (n NamespaceStatement) statementNode() {}

var _ Statement = &NamespaceStatement{}
var _ Node = &NamespaceStatement{}

type PolicyStatement struct {
	*baseNode
	Name       string
	Statements []Statement
}

func NewPolicyStatement(name string, statements []Statement, ssp tokens.Range) *PolicyStatement {
	return &PolicyStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "policy",
		},
		Name:       name,
		Statements: statements,
	}
}

func (p PolicyStatement) String() string {
	return p.Name
}

func (p PolicyStatement) statementNode() {}

var _ Statement = &PolicyStatement{}
var _ Node = &PolicyStatement{}

type Program struct {
	Statements []Statement
	Reference  string
}
