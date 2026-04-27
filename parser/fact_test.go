// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
)

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
