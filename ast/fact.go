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

import "github.com/sentrie-sh/sentrie/tokens"

type FactStatement struct {
	*baseNode
	Name     string     // Name of the fact
	Type     TypeRef    // Type of the fact
	Alias    string     // Exposed name of the fact
	Default  Expression // Default value expression (optional)
	Required bool       // Whether the fact is required
}

func NewFactStatement(name string, typeRef TypeRef, alias string, defaultExpr Expression, required bool, ssp tokens.Range) *FactStatement {
	return &FactStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "fact",
		},
		Name:     name,
		Type:     typeRef,
		Alias:    alias,
		Default:  defaultExpr,
		Required: required,
	}
}

func (f FactStatement) String() string {
	return f.Name
}

func (f FactStatement) statementNode() {}

var _ Statement = &FactStatement{}
var _ Node = &FactStatement{}
