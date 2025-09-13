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

	"github.com/binaek/sentra/tokens"
)

type BlockExpression struct /* implements Expression */ {
	Pos        tokens.Position
	Statements []Statement
	Yield      Expression
}

func (b *BlockExpression) expressionNode() {}

func (b *BlockExpression) Position() tokens.Position {
	return b.Pos
}

func (b *BlockExpression) String() string {
	var stmts []string
	for _, stmt := range b.Statements {
		stmts = append(stmts, stmt.String())
	}
	return "{" + strings.Join(stmts, ";") + "; yield " + b.Yield.String() + "}"
}
