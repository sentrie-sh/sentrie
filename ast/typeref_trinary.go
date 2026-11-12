// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// TrinaryTypeRef is the `trinary` type (true, false or unknown) and, with
// Boolean set, the `boolean` type (true or false only).
type TrinaryTypeRef struct {
	*baseTypeRef
	// Boolean records that the source spelled the type `boolean` rather than
	// `trinary`. A boolean is a trinary that rejects unknown, like
	// `trinary @not_unknown()`; the runtime validator and the union member
	// check enforce it (#144).
	Boolean bool
}

func NewTrinaryTypeRef(ssp tokens.Range) *TrinaryTypeRef {
	return &TrinaryTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "trinary_typeref",
			},
			validConstraints: genTrinaryConstraints,
		},
	}
}

var _ TypeRef = &TrinaryTypeRef{}
var _ Node = &TrinaryTypeRef{}

func (b *TrinaryTypeRef) String() string {
	if b.Boolean {
		return "boolean"
	}
	return "trinary"
}
