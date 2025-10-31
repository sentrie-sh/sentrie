package ast

import "github.com/sentrie-sh/sentrie/tokens"

type FieldAccessExpression struct {
	*baseNode
	Left  Expression
	Field string
}

func NewFieldAccessExpression(left Expression, field string, ssp tokens.Range) *FieldAccessExpression {
	return &FieldAccessExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "field_access",
		},
		Left:  left,
		Field: field,
	}
}

type IndexAccessExpression struct {
	*baseNode
	Left  Expression
	Index Expression
}

func NewIndexAccessExpression(left Expression, index Expression, ssp tokens.Range) *IndexAccessExpression {
	return &IndexAccessExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "index_access",
		},
		Left:  left,
		Index: index,
	}
}

var _ Expression = &FieldAccessExpression{}
var _ Node = &FieldAccessExpression{}

func (f *FieldAccessExpression) String() string {
	return f.Left.String() + "." + f.Field
}

func (i *IndexAccessExpression) String() string {
	return i.Left.String() + "[" + i.Index.String() + "]"
}

func (f *FieldAccessExpression) expressionNode() {}

func (i *IndexAccessExpression) expressionNode() {}

var _ Expression = &IndexAccessExpression{}
var _ Node = &IndexAccessExpression{}
