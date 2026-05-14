// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// DeriveStatement binds a named pure lambda at namespace or policy scope.
type DeriveStatement struct {
	*baseNode
	Name  string
	Value *LambdaExpression
}

func NewDeriveStatement(name string, value *LambdaExpression, ssp tokens.Range) *DeriveStatement {
	return &DeriveStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "derive",
		},
		Name:  name,
		Value: value,
	}
}

func (d *DeriveStatement) statementNode() {}

func (d *DeriveStatement) String() string {
	return fmt.Sprintf("derive %s = %s", d.Name, d.Value.String())
}

var _ Statement = &DeriveStatement{}
var _ Node = &DeriveStatement{}

// ExportDeriveStatement exports a namespace-level derive.
type ExportDeriveStatement struct {
	*baseNode
	Name string
}

func NewExportDeriveStatement(name string, ssp tokens.Range) *ExportDeriveStatement {
	return &ExportDeriveStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "export_derive",
		},
		Name: name,
	}
}

func (e *ExportDeriveStatement) statementNode() {}

func (e *ExportDeriveStatement) String() string {
	return "export derive " + e.Name
}

var _ Statement = &ExportDeriveStatement{}
var _ Node = &ExportDeriveStatement{}
