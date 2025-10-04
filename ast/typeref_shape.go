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

type ShapeTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
	Ref         FQN // Fully Qualified Name (FQN) of the shape
}

var _ TypeRef = &ShapeTypeRef{}
var _ Node = &ShapeTypeRef{}

func (s *ShapeTypeRef) typeref()                  {}
func (s *ShapeTypeRef) Position() tokens.Position { return s.Pos }
func (s *ShapeTypeRef) String() string            { return s.Ref.String() }
func (s *ShapeTypeRef) GetConstraints() []*TypeRefConstraint {
	return s.constraints
}
func (s *ShapeTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, shapeConstraints); err != nil {
		return err
	}
	s.constraints = append(s.constraints, constraint)
	return nil
}

var shapeConstraints = func() map[string]int {
	constraints := [...]v{
		{name: "required", arglen: 0},
		{name: "optional", arglen: 0},
	}
	constraintsMap := make(map[string]int)
	for _, v := range constraints {
		constraintsMap[v.name] = v.arglen
	}
	return constraintsMap
}()
