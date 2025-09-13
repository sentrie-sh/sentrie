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

type MapTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
	ValueType   TypeRef
}

var _ TypeRef = &MapTypeRef{}

func (m *MapTypeRef) typeref()                  {}
func (m *MapTypeRef) Position() tokens.Position { return m.Pos }
func (m *MapTypeRef) String() string            { return "map[" + m.ValueType.String() + "]" }
func (m *MapTypeRef) GetConstraints() []*TypeRefConstraint {
	return m.constraints
}
func (m *MapTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, mapConstraints); err != nil {
		return err
	}
	m.constraints = append(m.constraints, constraint)
	return nil
}

var mapConstraints = func() map[string]int {
	constraints := [...]v{
		{name: "minlength", arglen: 1},
		{name: "maxlength", arglen: 1},
		{name: "length", arglen: 1},
		{name: "keys", arglen: 1},
	}
	constraintsMap := make(map[string]int)
	for _, v := range constraints {
		constraintsMap[v.name] = v.arglen
	}
	return constraintsMap
}()
