// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

type TernaryExpression struct {
	*baseNode
	Condition  Expression
	ThenBranch Expression
	ElseBranch Expression
	// Elvis is true for `a ?: b` (parsed as `?` immediately followed by `:`): null/undefined
	// selects ElseBranch; otherwise Condition's value is used and ElseBranch is not evaluated.
	// ThenBranch is the same node as Condition for this form.
	Elvis bool
}

func NewTernaryExpression(condition Expression, thenBranch Expression, elseBranch Expression, ssp tokens.Range) *TernaryExpression {
	return &TernaryExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "ternary",
		},
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
		Elvis:      false,
	}
}

// NewTernaryElvis builds `lhs ?: rhs` as a ternary-shaped node (ThenBranch == Condition).
func NewTernaryElvis(lhs, rhs Expression, ssp tokens.Range) *TernaryExpression {
	return &TernaryExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "ternary",
		},
		Condition:  lhs,
		ThenBranch: lhs,
		ElseBranch: rhs,
		Elvis:      true,
	}
}
func (t *TernaryExpression) String() string {
	if t.Elvis {
		return fmt.Sprintf("(%s ?: %s)", t.Condition.String(), t.ElseBranch.String())
	}
	return fmt.Sprintf("(%s ? %s : %s)", t.Condition.String(), t.ThenBranch.String(), t.ElseBranch.String())
}

func (t *TernaryExpression) expressionNode() {}

var _ Expression = &TernaryExpression{}
var _ Node = &TernaryExpression{}
