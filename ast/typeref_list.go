// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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

type ListTypeRef struct {
	*baseTypeRef
	ElemType TypeRef
}

var _ TypeRef = &ListTypeRef{}
var _ Node = &ListTypeRef{}

func (l *ListTypeRef) String() string { return "list[" + l.ElemType.String() + "]" }
func NewListTypeRef(elemType TypeRef, ssp tokens.Range) *ListTypeRef {
	return &ListTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "list_typeref",
			},
			validConstraints: genListConstraints,
		},
		ElemType: elemType,
	}
}
