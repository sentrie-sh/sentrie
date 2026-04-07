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
	"github.com/sentrie-sh/sentrie/tokens"
)

// DescriptionStatement is a policy metadata line: description "…".
type DescriptionStatement struct {
	*baseNode
	Value string
}

func NewDescriptionStatement(value string, ssp tokens.Range) *DescriptionStatement {
	return &DescriptionStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "description",
		},
		Value: value,
	}
}

func (s *DescriptionStatement) String() string { return "description" }

func (s *DescriptionStatement) statementNode() {}

var _ Statement = (*DescriptionStatement)(nil)
var _ Node = (*DescriptionStatement)(nil)
