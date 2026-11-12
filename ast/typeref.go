// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

type TypeRef interface {
	Node
	typeref()
	GetConstraints() []*TypeRefConstraint
	AddConstraint(*TypeRefConstraint) error
}

type TypeRefConstraint struct {
	*baseNode
	Name string
	Args []Expression
}

func NewTypeRefConstraint(name string, args []Expression, ssp tokens.Range) *TypeRefConstraint {
	return &TypeRefConstraint{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "typeref_constraint",
		},
		Name: name,
		Args: args,
	}
}

type baseTypeRef struct {
	*baseNode
	constraints      []*TypeRefConstraint
	validConstraints map[string]int
}

func (b *baseTypeRef) typeref() {}

func (b *baseTypeRef) AddConstraint(constraint *TypeRefConstraint) error {
	if err := validateConstraint(constraint, b.validConstraints); err != nil {
		return err
	}
	b.constraints = append(b.constraints, constraint)
	b.Rnge.To = constraint.Rnge.To
	return nil
}

func (b *baseTypeRef) GetConstraints() []*TypeRefConstraint {
	return b.constraints
}

// validateConstraint checks if a constraint is valid for the given type
func validateConstraint(constraint *TypeRefConstraint, constraints map[string]int) error {
	expectedArgs, ok := constraints[constraint.Name]
	if !ok {
		return xerr.NotFoundError{}
	}
	if expectedArgs == -2 {
		// Any number of arguments, including none.
		return nil
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
