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

type NumberTypeRef struct {
	constraints []*TypeRefConstraint
	Range       tokens.Range
}

var _ TypeRef = &NumberTypeRef{}
var _ Node = &NumberTypeRef{}

func (i *NumberTypeRef) typeref()           {}
func (s *NumberTypeRef) Span() tokens.Range { return s.Range }
func (i *NumberTypeRef) Kind() string       { return "number_typeref" }
func (i *NumberTypeRef) String() string     { return "number" }
func (i *NumberTypeRef) GetConstraints() []*TypeRefConstraint {
	return i.constraints
}
func (i *NumberTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, genNumberConstraints); err != nil {
		return err
	}
	i.constraints = append(i.constraints, constraint)
	i.Range.To = constraint.Range.To
	return nil
}
