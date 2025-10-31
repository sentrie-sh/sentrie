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

type BoolTypeRef struct {
	*baseTypeRef
}

func NewBoolTypeRef(ssp tokens.Range) *BoolTypeRef {
	return &BoolTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "boolean_typeref",
			},
			validConstraints: genBoolConstraints,
		},
	}
}

var _ TypeRef = &BoolTypeRef{}
var _ Node = &BoolTypeRef{}

func (b *BoolTypeRef) String() string { return "boolean" }
