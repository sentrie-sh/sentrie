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

// VersionStatement is a policy metadata line: version "…" (SemVer validated at index time).
type VersionStatement struct {
	*baseNode
	Literal string
}

func NewVersionStatement(literal string, ssp tokens.Range) *VersionStatement {
	return &VersionStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "version",
		},
		Literal: literal,
	}
}

func (s *VersionStatement) String() string { return "version" }

func (s *VersionStatement) statementNode() {}

var _ Statement = (*VersionStatement)(nil)
var _ Node = (*VersionStatement)(nil)
