// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"strings"

	"github.com/sentrie-sh/sentrie/tokens"
)

// UnionTypeRef is `T1 | T2 | ...`: a value validates if it matches any
// member. The union owns no constraints; each member carries the `?` and
// `@constraint(...)` suffixes written after it, so `string | number?` is a
// union of `string` and `number?`, and `string | number@gt(0)` constrains
// only the number.
type UnionTypeRef struct {
	*baseTypeRef
	Members []TypeRef
}

var _ TypeRef = &UnionTypeRef{}
var _ Node = &UnionTypeRef{}

func NewUnionTypeRef(members []TypeRef, ssp tokens.Range) *UnionTypeRef {
	return &UnionTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "union_typeref",
			},
			validConstraints: map[string]int{},
		},
		Members: members,
	}
}

func (u *UnionTypeRef) String() string {
	parts := make([]string, len(u.Members))
	for i, m := range u.Members {
		parts[i] = m.String()
	}
	return strings.Join(parts, " | ")
}

// AcceptsNull reports whether null validates against t without reaching a
// kind check: a nullable `T?`, `any`, or a union with such a member.
func AcceptsNull(t TypeRef) bool {
	switch tr := t.(type) {
	case *NullableTypeRef, *AnyTypeRef:
		return true
	case *UnionTypeRef:
		for _, m := range tr.Members {
			if AcceptsNull(m) {
				return true
			}
		}
	}
	return false
}
