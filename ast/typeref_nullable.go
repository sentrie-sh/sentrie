// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type NullableTypeRef struct {
	*baseTypeRef
	Inner TypeRef
}

func NewNullableTypeRef(inner TypeRef, ssp tokens.Range) *NullableTypeRef {
	return &NullableTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "nullable_typeref",
			},
			validConstraints: map[string]int{},
		},
		Inner: inner,
	}
}

func (n *NullableTypeRef) String() string {
	return n.Inner.String() + "?"
}

func (n *NullableTypeRef) GetConstraints() []*TypeRefConstraint {
	return n.Inner.GetConstraints()
}

func (n *NullableTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	err := n.Inner.AddConstraint(constraint)
	if err != nil {
		return err
	}
	n.Rnge.To = n.Inner.Span().To
	return nil
}

func IsNullableTypeRef(typeRef TypeRef) bool {
	_, ok := typeRef.(*NullableTypeRef)
	return ok
}

func UnwrapNullableTypeRef(typeRef TypeRef) TypeRef {
	n, ok := typeRef.(*NullableTypeRef)
	if !ok {
		return typeRef
	}
	return n.Inner
}

var _ TypeRef = &NullableTypeRef{}
var _ Node = &NullableTypeRef{}
