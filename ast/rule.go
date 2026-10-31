// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type RuleStatement struct {
	*baseNode
	RuleName string
	// Type is the optional declared type (`rule name: T = ...`); nil when
	// absent. A rule that takes contributions must declare list[T] or a
	// boolean type.
	Type    TypeRef
	Default Expression
	When    Expression
	Body    Expression
}

func NewRuleStatement(ruleName string, defaultExpr Expression, whenExpr Expression, bodyExpr Expression, ssp tokens.Range) *RuleStatement {
	return &RuleStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "rule_statement",
		},
		RuleName: ruleName,
		Default:  defaultExpr,
		When:     whenExpr,
		Body:     bodyExpr,
	}
}
func (r RuleStatement) String() string {
	return r.RuleName
}

func (r RuleStatement) statementNode() {}

var _ Statement = &RuleStatement{}
var _ Node = &RuleStatement{}
