// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

type ShapeStatement struct {
	*baseNode
	Name    string
	Simple  TypeRef
	Complex *Cmplx
}

type Cmplx struct {
	Range  tokens.Range
	With   *FQN
	Node   Node
	Fields map[string]*ShapeField
	// Comments holds the comments inside the shape body, in source order.
	// Evaluation ignores them; a printer merges them with Fields by position.
	Comments []*CommentStatement
}

type ShapeField struct {
	Range    tokens.Range
	Name     string
	Optional bool
	Type     TypeRef
	Node     Node
}

func NewShapeStatement(name string, simple TypeRef, complex *Cmplx, ssp tokens.Range) *ShapeStatement {
	return &ShapeStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "shape",
		},
		Name:    name,
		Simple:  simple,
		Complex: complex,
	}
}

func (s *ShapeStatement) String() string {
	return s.Name
}

func (s *ShapeStatement) statementNode() {}

var _ Statement = &ShapeStatement{}
var _ Node = &ShapeStatement{}
