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

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

type TypeRef interface {
	Node
	typeref()
	String() string
	Position() tokens.Position
	GetConstraints() []*TypeRefConstraint
	AddConstraint(*TypeRefConstraint) error
}

type TypeRefConstraint struct {
	Pos  tokens.Position
	Name string
	Args []Expression
}

// validateConstraint checks if a constraint is valid for the given type
func validateConstraint(constraint *TypeRefConstraint, constraints map[string]int) error {
	expectedArgs, ok := constraints[constraint.Name]
	if !ok {
		return xerr.NotFoundError{}
	}
	if expectedArgs == -1 {
		// Variable arguments - at least 1 required
		if len(constraint.Args) < 1 {
			return fmt.Errorf("constraint %s requires at least 1 argument", constraint.Name)
		}
	} else if len(constraint.Args) != expectedArgs {
		return fmt.Errorf("invalid number of arguments for constraint %s", constraint.Name)
	}
	return nil
}
