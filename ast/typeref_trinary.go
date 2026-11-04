// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type TrinaryTypeRef struct {
	*baseTypeRef
	// Boolean records that the source spelled the type `boolean` rather than
	// `trinary`. The two are the same type; this only keeps the spelling.
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

func (b *TrinaryTypeRef) String() string { return "trinary" }
