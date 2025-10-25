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
	constraints []*TypeRefConstraint
	Range       tokens.Range
}

var _ TypeRef = &BoolTypeRef{}
var _ Node = &BoolTypeRef{}

func (b *BoolTypeRef) typeref()           {}
func (s *BoolTypeRef) Span() tokens.Range { return s.Range }
func (b *BoolTypeRef) Kind() string       { return "boolean_typeref" }
func (b *BoolTypeRef) String() string     { return "boolean" }
func (b *BoolTypeRef) GetConstraints() []*TypeRefConstraint {
	return b.constraints
}
func (b *BoolTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, genBoolConstraints); err != nil {
		return err
	}
	b.constraints = append(b.constraints, constraint)
	b.Range.To = constraint.Range.To
	return nil
}
