// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type CastExpression struct {
	*baseNode
	Expr       Expression
	TargetType TypeRef
}

func NewCastExpression(expr Expression, targetType TypeRef, ssp tokens.Range) *CastExpression {
	return &CastExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "cast",
		},
		Expr:       expr,
		TargetType: targetType,
	}
}
func (c *CastExpression) String() string {
	return c.Expr.String() + " as " + c.TargetType.String()
}

func (c *CastExpression) expressionNode() {}

var _ Expression = (*CastExpression)(nil)
var _ Node = (*CastExpression)(nil)
