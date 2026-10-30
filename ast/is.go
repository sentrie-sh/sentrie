// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type IsDefinedExpression struct {
	*baseNode
	Left Expression
}

type IsEmptyExpression struct {
	*baseNode
	Left Expression
}

func NewIsDefinedExpression(left Expression, ssp tokens.Range) *IsDefinedExpression {
	return &IsDefinedExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "is_defined",
		},
		Left: left,
	}
}

func NewIsEmptyExpression(left Expression, ssp tokens.Range) *IsEmptyExpression {
	return &IsEmptyExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "is_empty",
		},
		Left: left,
	}
}

func (e *IsEmptyExpression) String() string {
	return "is empty " + e.Left.String()
}

func (e *IsDefinedExpression) String() string {
	return "is defined " + e.Left.String()
}

func (e *IsDefinedExpression) expressionNode() {}
func (e *IsEmptyExpression) expressionNode()   {}

var _ Expression = &IsDefinedExpression{}
var _ Node = &IsDefinedExpression{}
var _ Expression = &IsEmptyExpression{}
var _ Node = &IsEmptyExpression{}

// IsExpression is `Left is <type>`: true when Left validates against the type.
//
// The parser sets Type alone when the right-hand side is type syntax (it
// starts with a type keyword, or a name carries a `?` or `@constraint`
// suffix). When the right-hand side is a bare or slash-qualified name, the
// parser cannot tell a shape from a value, so it sets both: Right holds the
// name as an expression and Type holds it as a shape reference. The index
// resolves that case against the shapes and bindings in scope and clears
// whichever reading does not apply, so after validation exactly one of Right
// and Type is set. With Right set the expression is the legacy equality
// `Left == Right`.
type IsExpression struct {
	*baseNode
	Left  Expression
	Right Expression
	Type  TypeRef
}

func NewIsExpression(left Expression, right Expression, typ TypeRef, ssp tokens.Range) *IsExpression {
	return &IsExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "is",
		},
		Left:  left,
		Right: right,
		Type:  typ,
	}
}

// Ambiguous reports whether the right-hand side is a name the index has not
// yet resolved to a shape or a value.
func (e *IsExpression) Ambiguous() bool {
	return e.Right != nil && e.Type != nil
}

func (e *IsExpression) String() string {
	if e.Type != nil {
		return e.Left.String() + " is " + e.Type.String()
	}
	return e.Left.String() + " is " + e.Right.String()
}

func (e *IsExpression) expressionNode() {}

var _ Expression = &IsExpression{}
var _ Node = &IsExpression{}
