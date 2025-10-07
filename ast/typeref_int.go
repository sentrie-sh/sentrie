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

type IntTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
}

var _ TypeRef = &IntTypeRef{}
var _ Node = &IntTypeRef{}

func (i *IntTypeRef) typeref()                  {}
func (s *IntTypeRef) Position() tokens.Position { return s.Pos }
func (i *IntTypeRef) String() string            { return "int" }
func (i *IntTypeRef) GetConstraints() []*TypeRefConstraint {
	return i.constraints
}
func (i *IntTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, genIntConstraints); err != nil {
		return err
	}
	i.constraints = append(i.constraints, constraint)
	return nil
}
