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

type StringTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
}

var _ TypeRef = &StringTypeRef{}

func (s *StringTypeRef) typeref()                  {}
func (s *StringTypeRef) Position() tokens.Position { return s.Pos }
func (s *StringTypeRef) String() string            { return "string" }
func (s *StringTypeRef) GetConstraints() []*TypeRefConstraint {
	return s.constraints
}

func (s *StringTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, stringConstraints); err != nil {
		return err
	}
	s.constraints = append(s.constraints, constraint)
	return nil
}

var stringConstraints = func() map[string]int {
	constraints := [...]v{
		{name: "minlength", arglen: 1},
		{name: "maxlength", arglen: 1},
		{name: "length", arglen: 1},
		{name: "regexp", arglen: 1},
		{name: "starts_with", arglen: 1},
		{name: "ends_with", arglen: 1},
		{name: "has_substring", arglen: 1},
		{name: "not_has_substring", arglen: 1},
		{name: "email", arglen: 0},
		{name: "url", arglen: 0},
		{name: "uuid", arglen: 0},
		{name: "alphanumeric", arglen: 0},
		{name: "alpha", arglen: 0},
		{name: "numeric", arglen: 0},
		{name: "lowercase", arglen: 0},
		{name: "uppercase", arglen: 0},
		{name: "trimmed", arglen: 0},
		{name: "not_empty", arglen: 0},
		{name: "one_of", arglen: -1}, // variable args
	}
	constraintsMap := make(map[string]int)
	for _, v := range constraints {
		constraintsMap[v.name] = v.arglen
	}
	return constraintsMap
}()
