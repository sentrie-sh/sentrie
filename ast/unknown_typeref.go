// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// UnknownTypeRef is a defensive TypeRef placeholder for forward-compatible kind checks.
// The current parser does not produce this node; it exists so validate-time kind
// resolution can treat unrecognized TypeRef implementations as unknown.
type UnknownTypeRef struct {
	*baseTypeRef
}

func NewUnknownTypeRef(ssp tokens.Range) TypeRef {
	return &UnknownTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "unknown_typeref",
			},
		},
	}
}

func (u *UnknownTypeRef) typeref() {}

func (u *UnknownTypeRef) String() string { return "unknown" }
