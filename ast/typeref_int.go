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

import "github.com/binaek/sentra/tokens"

type IntTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
}

var _ TypeRef = &IntTypeRef{}

func (i *IntTypeRef) typeref()                  {}
func (s *IntTypeRef) Position() tokens.Position { return s.Pos }
func (i *IntTypeRef) String() string            { return "int" }
func (i *IntTypeRef) GetConstraints() []*TypeRefConstraint {
	return i.constraints
}
func (i *IntTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, intConstraints); err != nil {
		return err
	}
	i.constraints = append(i.constraints, constraint)
	return nil
}

var intConstraints = func() map[string]int {
	constraints := [...]v{
		{name: "min", arglen: 1},
		{name: "max", arglen: 1},
		{name: "range", arglen: 2},
		{name: "multiple_of", arglen: 1},
		{name: "even", arglen: 0},
		{name: "odd", arglen: 0},
		{name: "positive", arglen: 0},
		{name: "negative", arglen: 0},
		{name: "non_negative", arglen: 0},
		{name: "non_positive", arglen: 0},
	}
	constraintsMap := make(map[string]int)
	for _, v := range constraints {
		constraintsMap[v.name] = v.arglen
	}
	return constraintsMap
}()
