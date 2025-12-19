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

package index

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

type Rule struct {
	Node    *ast.RuleStatement
	Policy  *Policy
	Name    string
	FQN     ast.FQN
	Default ast.Expression
	When    ast.Expression
	Body    ast.Expression
}

func (r *Rule) String() string {
	return r.FQN.String()
}

func (r *Rule) Span() tokens.Range {
	return r.Node.Span()
}

func createRule(p *Policy, stmt *ast.RuleStatement) (*Rule, error) {
	return &Rule{
		Node:    stmt,
		Policy:  p,
		Name:    stmt.RuleName,
		FQN:     ast.CreateFQN(p.FQN, stmt.RuleName),
		Default: stmt.Default,
		When:    stmt.When,
		Body:    stmt.Body,
	}, nil
}
