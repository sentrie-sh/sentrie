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
	constraints []*TypeRefConstraint
	Range       tokens.Range
	ElemType    TypeRef
}

var _ TypeRef = &ListTypeRef{}
var _ Node = &ListTypeRef{}

func (a *ListTypeRef) typeref()           {}
func (a *ListTypeRef) Span() tokens.Range { return a.Range }
func (a *ListTypeRef) Kind() string       { return "list_typeref" }
func (a *ListTypeRef) String() string     { return "array[" + a.ElemType.String() + "]" }
func (a *ListTypeRef) GetConstraints() []*TypeRefConstraint {
	return a.constraints
}
func (a *ListTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, genListConstraints); err != nil {
		return err
	}
	a.constraints = append(a.constraints, constraint)
	a.Range.To = constraint.Range.To
	return nil
}
