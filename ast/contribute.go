// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// ContributeStatement is `contribute <rule> [when <expr>] <expr>`: one
// incremental contribution to a rule declared with `rule` in the same policy.
type ContributeStatement struct {
	*baseNode
	RuleName string
	When     Expression
	Body     Expression
}

func NewContributeStatement(ruleName string, whenExpr Expression, bodyExpr Expression, ssp tokens.Range) *ContributeStatement {
	return &ContributeStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "contribute_statement",
		},
		RuleName: ruleName,
		When:     whenExpr,
		Body:     bodyExpr,
	}
}

func (c ContributeStatement) String() string {
	return "contribute " + c.RuleName
}

func (c ContributeStatement) statementNode() {}

var _ Statement = &ContributeStatement{}
var _ Node = &ContributeStatement{}
