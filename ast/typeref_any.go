// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// AnyTypeName is the type name that denotes the top type. It is not a lexer
// keyword: `any` stays an identifier everywhere else (the `any` quantifier
// builtin), and only a bare `any` in a type position means AnyTypeRef.
const AnyTypeName = "any"

// AnyTypeRef is the top type `any`: it accepts every defined value, scalars
// and null included. It permits no constraints.
type AnyTypeRef struct {
	*baseTypeRef
}

var _ TypeRef = &AnyTypeRef{}
var _ Node = &AnyTypeRef{}

func NewAnyTypeRef(ssp tokens.Range) *AnyTypeRef {
	return &AnyTypeRef{
		baseTypeRef: &baseTypeRef{
			baseNode: &baseNode{
				Rnge:  ssp,
				Kind_: "any_typeref",
			},
			validConstraints: map[string]int{},
		},
	}
}

func (a *AnyTypeRef) String() string { return AnyTypeName }

// notATypeHints maps the type names other languages use, which are not
// Sentrie types, to the Sentrie type to use instead (#144). They are not
// lexer keywords, so they stay ordinary identifiers and field names
// everywhere outside a type position.
var notATypeHints = map[string]string{
	"bool":  "boolean",
	"int":   "number",
	"float": "number",
}

// NotATypeHint reports whether name is one of the familiar type names that
// Sentrie does not have (`bool`, `int`, `float`), and the type to suggest.
func NotATypeHint(name string) (string, bool) {
	hint, ok := notATypeHints[name]
	return hint, ok
}
