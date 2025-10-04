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

// 'import value|decision @ident from @string { @WithClause }'
type ImportClause struct {
	Pos           tokens.Position // Position in the source code
	RuleToImport  string          // The name of the rule being imported
	FromPolicyFQN FQN             // The source identifier - segmented by '/'
	Withs         []*WithClause   // Inline with import clause
}

// 'with @ident as @string'
// Represents a 'with' clause in an import statement, allowing for additional context or configuration.
type WithClause struct {
	Pos  tokens.Position // Position in the source code
	Name string          // Name of the with clause - this is also the name that the target policy exposes
	Expr Expression      // Value associated with the with clause
}

func (i ImportClause) String() string {
	return i.RuleToImport
}

func (i ImportClause) Position() tokens.Position {
	return i.Pos
}

func (i ImportClause) expressionNode() {}

var _ Expression = &ImportClause{}
var _ Node = &ImportClause{}
