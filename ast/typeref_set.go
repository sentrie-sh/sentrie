// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// SetTypeRef is `set[T]`: a list of distinct elements of type T (#162). It
// takes the list constraints, since a set is a list at runtime.
type SetTypeRef struct {
	*baseTypeRef
	ElemType TypeRef
}

var _ TypeRef = &SetTypeRef{}
var _ Node = &SetTypeRef{}

func (s *SetTypeRef) String() string { return "set[" + s.ElemType.String() + "]" }

func NewSetTypeRef(elemType TypeRef, ssp tokens.Range) *SetTypeRef {
	return &SetTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "set_typeref",
			},
			validConstraints: genListConstraints,
		},
		ElemType: elemType,
	}
}
