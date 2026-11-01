// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"strconv"

	"github.com/sentrie-sh/sentrie/tokens"
)

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	*baseNode
	// Value is exact; the runtime boxes it as an exact integer (#166).
	Value int64
}

func NewIntegerLiteral(value int64, ssp tokens.Range) *IntegerLiteral {
	return &IntegerLiteral{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "integer_literal",
		},
		Value: value,
	}
}

func (i *IntegerLiteral) String() string {
	return strconv.FormatInt(i.Value, 10)
}

func (i *IntegerLiteral) expressionNode() {}

var _ Expression = &IntegerLiteral{}
var _ Node = &IntegerLiteral{}
