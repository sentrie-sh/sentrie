// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
