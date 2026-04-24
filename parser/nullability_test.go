// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) TestParseShapeFieldNullabilityMatrix() {
	parser := NewParserFromString("shape Person { name:string age?:number middle_name:string? nickname?:string? }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	s.Require().NotNil(shapeStmt.Complex)

	s.False(shapeStmt.Complex.Fields["name"].Optional)
	s.False(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["name"].Type))

	s.True(shapeStmt.Complex.Fields["age"].Optional)
	s.False(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["age"].Type))

	s.False(shapeStmt.Complex.Fields["middle_name"].Optional)
	s.True(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["middle_name"].Type))

	s.True(shapeStmt.Complex.Fields["nickname"].Optional)
	s.True(ast.IsNullableTypeRef(shapeStmt.Complex.Fields["nickname"].Type))
}

func (s *ParserTestSuite) TestParseShapeFieldRejectsLegacyBangSyntax() {
	testCases := []struct {
		input   string
		message string
	}{
		{"shape Person { name!: string }", "name!: T"},
		{"shape Person { phone!?: string }", "phone!?: T"},
		{"shape Person { phone?!: string }", "phone?!: T"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseShapeStatement(context.Background(), parser)
		s.Require().Nil(stmt)
		s.Require().Error(parser.err)
		s.Contains(parser.err.Error(), tc.message)
	}
}

func (s *ParserTestSuite) TestParseFactNullableTypeRef() {
	parser := NewParserFromString("fact input?: string?", "test.sentra")
	stmt := parseFactStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	factStmt, ok := stmt.(*ast.FactStatement)
	s.Require().True(ok)
	s.True(factStmt.Optional)
	s.True(ast.IsNullableTypeRef(factStmt.Type))
}

func (s *ParserTestSuite) TestParseNullableTypeRefWithConstraint() {
	parser := NewParserFromString("shape Person { middle_name: string? @minlength(1) }", "test.sentra")
	stmt := parseShapeStatement(context.Background(), parser)
	s.Require().NoError(parser.err)
	s.Require().NotNil(stmt)

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.Require().True(ok)
	typ := shapeStmt.Complex.Fields["middle_name"].Type
	s.True(ast.IsNullableTypeRef(typ))
	s.Len(ast.UnwrapNullableTypeRef(typ).GetConstraints(), 1)
}
