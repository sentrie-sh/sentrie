// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// DataStatement declares that a policy reads one entry of the pack's static
// `[data]` table: `data <name> (':' <type>)? ('as' <alias>)?`. The value is
// bound read-only under Alias for the whole policy.
type DataStatement struct {
	*baseNode
	Name  string  // Entry name in the pack manifest's [data] table
	Type  TypeRef // Optional type the value is validated against; nil when absent
	Alias string  // Name the policy reads the value by (defaults to Name)
}

func NewDataStatement(name string, typeRef TypeRef, alias string, ssp tokens.Range) *DataStatement {
	return &DataStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "data",
		},
		Name:  name,
		Type:  typeRef,
		Alias: alias,
	}
}

func (d DataStatement) String() string {
	return d.Name
}

func (d DataStatement) statementNode() {}

var _ Statement = &DataStatement{}
var _ Node = &DataStatement{}
