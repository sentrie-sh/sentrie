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

type DocumentTypeRef struct {
	constraints []*TypeRefConstraint
	Pos         tokens.Position
}

var _ TypeRef = &DocumentTypeRef{}

func (d *DocumentTypeRef) typeref()                  {}
func (d *DocumentTypeRef) Position() tokens.Position { return d.Pos }
func (d *DocumentTypeRef) String() string            { return "document" }
func (d *DocumentTypeRef) GetConstraints() []*TypeRefConstraint {
	return d.constraints
}

func (d *DocumentTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, documentConstraints); err != nil {
		return err
	}
	d.constraints = append(d.constraints, constraint)
	return nil
}

var documentConstraints = func() map[string]int {
	constraints := [...]v{
		{name: "minlength", arglen: 1},
		{name: "maxlength", arglen: 1},
		{name: "length", arglen: 1},
		{name: "keys", arglen: 1},
		{name: "required", arglen: 1},
		{name: "schema", arglen: 1},
	}
	constraintsMap := make(map[string]int)
	for _, v := range constraints {
		constraintsMap[v.name] = v.arglen
	}
	return constraintsMap
}()
