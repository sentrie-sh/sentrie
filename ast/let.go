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
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

type VarDeclaration struct {
	*baseNode
	Name  string
	Type  TypeRef
	Value Expression
}

func NewVarDeclaration(name string, typeRef TypeRef, value Expression, ssp tokens.Range) *VarDeclaration {
	return &VarDeclaration{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "let",
		},
		Name:  name,
		Type:  typeRef,
		Value: value,
	}
}
func (v VarDeclaration) String() string {
	return fmt.Sprintf("%s: %s = %s", v.Name, v.Type.String(), v.Value.String())
}

func (v VarDeclaration) statementNode() {}

var _ Statement = &VarDeclaration{}
var _ Node = &VarDeclaration{}
