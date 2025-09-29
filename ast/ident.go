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

// Identifier represents an identifier
type Identifier struct {
	Pos   tokens.Position
	Value string
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *Identifier) Position() tokens.Position {
	return i.Pos
}

func (i *Identifier) expressionNode() {}

var _ Expression = &Identifier{}
