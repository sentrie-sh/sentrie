// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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

	// Contributions are the `contribute` statements that add to this rule,
	// in declaration order. Empty for a rule without contributions.
	Contributions []*Contribution
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
