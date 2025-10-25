package ast

import "github.com/sentrie-sh/sentrie/tokens"

type FieldAccessExpression struct {
	Range tokens.Range
	Left  Expression
	Field string
}

type IndexAccessExpression struct {
	Range tokens.Range
	Left  Expression
	Index Expression
}

var _ Expression = &FieldAccessExpression{}
var _ Node = &FieldAccessExpression{}

func (f *FieldAccessExpression) String() string {
	return f.Left.String() + "." + f.Field
}

func (i *IndexAccessExpression) String() string {
	return i.Left.String() + "[" + i.Index.String() + "]"
}

func (f *FieldAccessExpression) Span() tokens.Range {
	return f.Range
}

func (f *FieldAccessExpression) Kind() string {
	return "field_access"
}

func (f *FieldAccessExpression) expressionNode() {}

func (i *IndexAccessExpression) Span() tokens.Range {
	return i.Range
}

func (i *IndexAccessExpression) Kind() string {
	return "index_access"
}

func (i *IndexAccessExpression) expressionNode() {}

var _ Expression = &IndexAccessExpression{}
var _ Node = &IndexAccessExpression{}
