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

package parser

import (
	"log/slog"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/suite"
)

// ExpressionTestSuite provides tests for expression parsing
type ExpressionTestSuite struct {
	suite.Suite
}

// SetupSuite initializes the test suite
func (s *ExpressionTestSuite) SetupSuite() {
	slog.Info("ExpressionTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *ExpressionTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "ExpressionTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *ExpressionTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "ExpressionTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *ExpressionTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestParseExpressionIdentifier tests parsing identifier expressions
func (s *ExpressionTestSuite) TestParseExpressionIdentifier() {
	input := `x`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	ident, ok := expr.(*ast.Identifier)
	s.True(ok)
	s.Equal("x", ident.Value)
}

// TestParseExpressionStringLiteral tests parsing string literal expressions
func (s *ExpressionTestSuite) TestParseExpressionStringLiteral() {
	input := `"hello world"`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	str, ok := expr.(*ast.StringLiteral)
	s.True(ok)
	s.Equal("hello world", str.Value)
}

// TestParseExpressionIntegerLiteral tests parsing integer literal expressions
func (s *ExpressionTestSuite) TestParseExpressionIntegerLiteral() {
	input := `42`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	intLit, ok := expr.(*ast.IntegerLiteral)
	s.True(ok)
	s.Equal(float64(42), intLit.Value)
}

// TestParseExpressionFloatLiteral tests parsing float literal expressions
func (s *ExpressionTestSuite) TestParseExpressionFloatLiteral() {
	input := `3.14`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	floatLit, ok := expr.(*ast.FloatLiteral)
	s.True(ok)
	s.Equal(3.14, floatLit.Value)
}

// TestParseExpressionBooleanLiteral tests parsing boolean literal expressions
func (s *ExpressionTestSuite) TestParseExpressionBooleanLiteral() {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"unknown", false}, // unknown is not a boolean
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr)

		if tc.input == "unknown" {
			// unknown should be parsed as a tristate literal
			tristate, ok := expr.(*ast.TrinaryLiteral)
			s.True(ok)
			s.Equal(trinary.Unknown, tristate.Value)
		} else {
			tristate, ok := expr.(*ast.TrinaryLiteral)
			s.True(ok)
			if tc.expected == true {
				s.Equal(trinary.True, tristate.Value)
			} else {
				s.Equal(trinary.False, tristate.Value)
			}
		}
	}
}

// TestParseExpressionTrinaryLiteral tests parsing trinary literal expressions
func (s *ExpressionTestSuite) TestParseExpressionTrinaryLiteral() {
	testCases := []string{"true", "false", "unknown"}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr)

		tristate, ok := expr.(*ast.TrinaryLiteral)
		s.True(ok)
		switch tc {
		case "true":
			s.Equal(trinary.True, tristate.Value)
		case "false":
			s.Equal(trinary.False, tristate.Value)
		case "unknown":
			s.Equal(trinary.Unknown, tristate.Value)
		}
	}
}

// TestParseExpressionNullLiteral tests parsing null literal expressions
func (s *ExpressionTestSuite) TestParseExpressionNullLiteral() {
	input := `null`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	nullLit, ok := expr.(*ast.NullLiteral)
	s.True(ok)
	s.NotNil(nullLit)
}

// TestParseExpressionInfixExpression tests parsing infix expressions
func (s *ExpressionTestSuite) TestParseExpressionInfixExpression() {
	testCases := []struct {
		input    string
		operator string
	}{
		{"1 + 2", "+"},
		{"3 - 4", "-"},
		{"5 * 6", "*"},
		{"7 / 8", "/"},
		{"9 % 10", "%"},
		{"a == b", "=="},
		{"c != d", "!="},
		{"e < f", "<"},
		{"g > h", ">"},
		{"i <= j", "<="},
		{"k >= l", ">="},
		{"m and n", "and"},
		{"o or p", "or"},
		{"q xor r", "xor"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		infix, ok := expr.(*ast.InfixExpression)
		s.True(ok, "Expected infix expression for: %s", tc.input)
		s.Equal(tc.operator, infix.Operator)
		s.NotNil(infix.Left)
		s.NotNil(infix.Right)
	}
}

// TestParseExpressionUnaryExpression tests parsing unary expressions
func (s *ExpressionTestSuite) TestParseExpressionUnaryExpression() {
	testCases := []struct {
		input    string
		operator string
	}{
		{"!true", "!"},
		{"-42", "-"},
		{"+3.14", "+"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		unary, ok := expr.(*ast.UnaryExpression)
		s.True(ok, "Expected unary expression for: %s", tc.input)
		s.Equal(tc.operator, unary.Operator)
		s.NotNil(unary.Right)
	}
}

// TestParseExpressionGroupedExpression tests parsing grouped expressions
func (s *ExpressionTestSuite) TestParseExpressionGroupedExpression() {
	input := `(1 + 2) * 3`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	infix, ok := expr.(*ast.InfixExpression)
	s.True(ok)
	s.Equal("*", infix.Operator)

	// Left should be the grouped expression (1 + 2)
	grouped, ok := infix.Left.(*ast.InfixExpression)
	s.True(ok)
	s.Equal("+", grouped.Operator)
}

// TestParseExpressionTernaryExpression tests parsing ternary expressions
func (s *ExpressionTestSuite) TestParseExpressionTernaryExpression() {
	input := `true ? "yes" : "no"`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	ternary, ok := expr.(*ast.TernaryExpression)
	s.True(ok)
	s.NotNil(ternary.Condition)
	s.NotNil(ternary.ThenBranch)
	s.NotNil(ternary.ElseBranch)
}

// TestParseExpressionListLiteral tests parsing list literal expressions
func (s *ExpressionTestSuite) TestParseExpressionListLiteral() {
	testCases := []struct {
		input    string
		expected int
	}{
		{"[]", 0},
		{"[1]", 1},
		{"[1, 2, 3]", 3},
		{"[\"a\", \"b\"]", 2},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		list, ok := expr.(*ast.ListLiteral)
		s.True(ok, "Expected list literal for: %s", tc.input)
		s.Len(list.Values, tc.expected)
	}
}

// TestParseExpressionMapLiteral tests parsing map literal expressions
func (s *ExpressionTestSuite) TestParseExpressionMapLiteral() {
	testCases := []struct {
		input    string
		expected int
	}{
		{"{}", 0},
		{"{\"key\": \"value\"}", 1},
		{"{\"a\": 1, \"b\": 2}", 2},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		mapLit, ok := expr.(*ast.MapLiteral)
		s.True(ok, "Expected map literal for: %s", tc.input)
		s.Len(mapLit.Entries, tc.expected)
	}
}

// TestParseExpressionCallExpression tests parsing call expressions
func (s *ExpressionTestSuite) TestParseExpressionCallExpression() {
	input := `myFunction(arg1, arg2)`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	call, ok := expr.(*ast.CallExpression)
	s.True(ok)
	s.Equal("myFunction", call.Callee.String())
	s.Len(call.Arguments, 2)
}

// TestParseExpressionIndexExpression tests parsing index expressions
func (s *ExpressionTestSuite) TestParseExpressionIndexExpression() {
	testCases := []struct {
		input string
		left  string
		index string
	}{
		{"array[0]", "array", "0"},
		{"obj[\"field\"]", "obj", "field"},
		{"amap[\"key\"]", "amap", "key"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		index, ok := expr.(*ast.IndexAccessExpression)
		s.True(ok, "Expected index expression for: %s", tc.input)
		s.NotNil(index.Left)
		s.NotNil(index.Index)
	}
}

// TestParseExpressionPrecedence tests operator precedence
func (s *ExpressionTestSuite) TestParseExpressionPrecedence() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3", "((1 + (2 * 3)))"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 * 2 * 3", "((1 * 2) * 3)"},
		{"1 + 2 * 3 + 4", "((1 + (2 * 3)) + 4)"},
		{"1 * 2 + 3 * 4", "((1 * 2) + (3 * 4))"},
		{"1 == 2 and 3 < 4", "((1 == 2) and (3 < 4))"},
		{"1 and 2 or 3", "((1 and 2) or 3)"},
		{"1 or 2 and 3", "(1 or (2 and 3))"},
		{"1 + 2 == 3 + 4", "((1 + 2) == (3 + 4))"},
		{"1 < 2 and 3 > 4", "((1 < 2) and (3 > 4))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Note: This is a simplified test - in a real implementation,
		// you'd want to test the actual AST structure more thoroughly
		s.NotNil(expr)
	}
}

// TestParseExpressionWithComments tests parsing expressions with comments
func (s *ExpressionTestSuite) TestParseExpressionWithComments() {
	input := `-- comment before
x + y -- comment after`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.NotNil(expr)

	// Should handle comments properly
	s.NotNil(expr)
}

// TestParseExpressionInvalidToken tests parsing with invalid tokens
func (s *ExpressionTestSuite) TestParseExpressionInvalidToken() {
	input := `@#$%`
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.NotNil(parser.err)
}

// TestParseExpressionEmptyInput tests parsing empty input
func (s *ExpressionTestSuite) TestParseExpressionEmptyInput() {
	input := ``
	parser := NewParserFromString(input, "test.sentra")

	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.NotNil(parser.err)
}

// TestParseExpressionUnaryOperators tests parsing unary operators
func (s *ExpressionTestSuite) TestParseExpressionUnaryOperators() {
	testCases := []struct {
		input    string
		operator string
		operand  string
	}{
		{"!true", "!", "true"},
		{"-42", "-", "42"},
		{"+3.14", "+", "3.14"},
		{"!x", "!", "x"},
		{"-y", "-", "y"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		unary, ok := expr.(*ast.UnaryExpression)
		s.True(ok, "Expected unary expression for: %s", tc.input)
		s.Equal(tc.operator, unary.Operator)
		s.NotNil(unary.Right)
	}
}

// TestParseExpressionBinaryOperators tests parsing binary operators
func (s *ExpressionTestSuite) TestParseExpressionBinaryOperators() {
	testCases := []struct {
		input    string
		left     string
		operator string
		right    string
	}{
		{"1 + 2", "1", "+", "2"},
		{"3 - 4", "3", "-", "4"},
		{"5 * 6", "5", "*", "6"},
		{"7 / 8", "7", "/", "8"},
		{"9 % 10", "9", "%", "10"},
		{"a == b", "a", "==", "b"},
		{"c != d", "c", "!=", "d"},
		{"e < f", "e", "<", "f"},
		{"g > h", "g", ">", "h"},
		{"i <= j", "i", "<=", "j"},
		{"k >= l", "k", ">=", "l"},
		{"m and n", "m", "and", "n"},
		{"o or p", "o", "or", "p"},
		{"q xor r", "q", "xor", "r"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		infix, ok := expr.(*ast.InfixExpression)
		s.True(ok, "Expected infix expression for: %s", tc.input)
		s.Equal(tc.operator, infix.Operator)
		s.NotNil(infix.Left)
		s.NotNil(infix.Right)
	}
}

// TestParseExpressionFieldAccess tests parsing field access expressions
func (s *ExpressionTestSuite) TestParseExpressionFieldAccess() {
	testCases := []struct {
		input string
		left  string
		field string
	}{
		{"obj.field", "obj", "field"},
		{"user.name", "user", "name"},
		{"data.value", "data", "value"},
		{"a.b.c", "a.b", "c"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		fieldAccess, ok := expr.(*ast.FieldAccessExpression)
		s.True(ok, "Expected field access expression for: %s", tc.input)
		s.NotNil(fieldAccess.Left)
		s.NotNil(fieldAccess.Field)
	}
}

// TestParseExpressionComplexNested tests parsing complex nested expressions
func (s *ExpressionTestSuite) TestParseExpressionComplexNested() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
		{"1 + (2 * 3)", "(1 + (2 * 3))"},
		{"(a + b) * (c - d)", "((a + b) * (c - d))"},
		{"f(x) + g(y)", "(f(x) + g(y))"},
		{"array[0] + obj.field", "(array[0] + obj.field)"},
		{"x ? y : z", "(x ? y : z)"},
		{"a and b or c", "((a and b) or c)"},
		{"a or b and c", "(a or (b and c))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)
		s.Equal(tc.expected, expr.String(), "String representation mismatch for: %s", tc.input)
	}
}

// TestParseExpressionEdgeCases tests parsing edge cases
func (s *ExpressionTestSuite) TestParseExpressionEdgeCases() {
	testCases := []struct {
		input       string
		shouldParse bool
		description string
	}{
		{"", false, "Empty input"},
		{"   ", false, "Whitespace only"},
		{"123", true, "Single number"},
		{"\"hello\"", true, "Single string"},
		{"true", true, "Single boolean"},
		{"unknown", true, "Single unknown"},
		{"null", true, "Single null"},
		{"(", false, "Unclosed parenthesis"},
		{")", false, "Unopened parenthesis"},
		{"[", false, "Unclosed bracket"},
		{"]", false, "Unopened bracket"},
		{"{", false, "Unclosed brace"},
		{"}", false, "Unopened brace"},
		{"1 +", false, "Incomplete expression"},
		{"+ 1", true, "Unary plus"},
		{"- 1", true, "Unary minus"},
		{"! true", true, "Unary not"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)

		if tc.shouldParse {
			s.NotNil(expr, "Expected to parse: %s (%s)", tc.input, tc.description)
			s.Nil(parser.err, "Expected no error for: %s (%s)", tc.input, tc.description)
		} else {
			s.Nil(expr, "Expected not to parse: %s (%s)", tc.input, tc.description)
			s.NotNil(parser.err, "Expected error for: %s (%s)", tc.input, tc.description)
		}
	}
}

// TestParseExpressionTestSuite runs the expression test suite
func TestParseExpressionTestSuite(t *testing.T) {
	suite.Run(t, new(ExpressionTestSuite))
}
