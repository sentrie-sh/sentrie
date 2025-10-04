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

type MapEntry struct {
	Key   string
	Value Expression
}

type MapLiteral struct {
	Pos     tokens.Position
	Entries []MapEntry
}

func (m *MapLiteral) String() string {
	result := "{"
	entries := []string{}
	for _, entry := range m.Entries {
		entries = append(entries, entry.Key+": "+entry.Value.String())
	}
	result += strings.Join(entries, ", ")
	result += "}"
	return result
}

func (m *MapLiteral) Position() tokens.Position {
	return m.Pos
}

func (m *MapLiteral) expressionNode() {}

var _ Expression = &MapLiteral{}
var _ Node = &MapLiteral{}
