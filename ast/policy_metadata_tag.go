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

// TagStatement is a policy metadata line: tag "key" = "value".
type TagStatement struct {
	*baseNode
	Key   string
	Value string
}

func NewTagStatement(key, value string, ssp tokens.Range) *TagStatement {
	return &TagStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "tag",
		},
		Key:   key,
		Value: value,
	}
}

func (s *TagStatement) String() string { return "tag" }

func (s *TagStatement) statementNode() {}

var _ Statement = (*TagStatement)(nil)
var _ Node = (*TagStatement)(nil)
