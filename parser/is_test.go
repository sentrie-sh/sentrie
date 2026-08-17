// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

// TestParseIsExpressionIncompleteRightOperand does not panic when the right side is missing.
func (s *ParserTestSuite) TestParseIsExpressionIncompleteRightOperand() {
	parser := NewParserFromString("x is", "test.sentra")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.Error(parser.err)
}
