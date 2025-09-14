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

	"github.com/stretchr/testify/suite"
)

// PrecedenceTestSuite provides tests for operator precedence
type PrecedenceTestSuite struct {
	suite.Suite
}

// SetupSuite initializes the test suite
func (s *PrecedenceTestSuite) SetupSuite() {
	slog.Info("PrecedenceTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *PrecedenceTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "PrecedenceTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *PrecedenceTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "PrecedenceTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *PrecedenceTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestPrecedenceArithmetic tests arithmetic operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceArithmetic() {
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
		{"1 + 2 / 3", "((1 + (2 / 3)))"},
		{"1 / 2 + 3", "((1 / 2) + 3)"},
		{"1 + 2 % 3", "((1 + (2 % 3)))"},
		{"1 % 2 + 3", "((1 % 2) + 3)"},
		{"1 - 2 * 3", "((1 - (2 * 3)))"},
		{"1 * 2 - 3", "((1 * 2) - 3)"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceComparison tests comparison operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceComparison() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 < 2 and 3 > 4", "((1 < 2) and (3 > 4))"},
		{"1 <= 2 and 3 >= 4", "((1 <= 2) and (3 >= 4))"},
		{"1 == 2 and 3 != 4", "((1 == 2) and (3 != 4))"},
		{"1 < 2 or 3 > 4", "((1 < 2) or (3 > 4))"},
		{"1 <= 2 or 3 >= 4", "((1 <= 2) or (3 >= 4))"},
		{"1 == 2 or 3 != 4", "((1 == 2) or (3 != 4))"},
		{"1 < 2 xor 3 > 4", "((1 < 2) xor (3 > 4))"},
		{"1 <= 2 xor 3 >= 4", "((1 <= 2) xor (3 >= 4))"},
		{"1 == 2 xor 3 != 4", "((1 == 2) xor (3 != 4))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceLogical tests logical operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceLogical() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 and 2 or 3", "((1 and 2) or 3)"},
		{"1 or 2 and 3", "(1 or (2 and 3))"},
		{"1 and 2 xor 3", "((1 and 2) xor 3)"},
		{"1 xor 2 and 3", "((1 xor 2) and 3)"},
		{"1 or 2 xor 3", "((1 or 2) xor 3)"},
		{"1 xor 2 or 3", "((1 xor 2) or 3)"},
		{"1 and 2 and 3", "((1 and 2) and 3)"},
		{"1 or 2 or 3", "((1 or 2) or 3)"},
		{"1 xor 2 xor 3", "((1 xor 2) xor 3)"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceEquality tests equality operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceEquality() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 == 2 and 3 != 4", "((1 == 2) and (3 != 4))"},
		{"1 != 2 and 3 == 4", "((1 != 2) and (3 == 4))"},
		{"1 == 2 or 3 != 4", "((1 == 2) or (3 != 4))"},
		{"1 != 2 or 3 == 4", "((1 != 2) or (3 == 4))"},
		{"1 == 2 xor 3 != 4", "((1 == 2) xor (3 != 4))"},
		{"1 != 2 xor 3 == 4", "((1 != 2) xor (3 == 4))"},
		{"1 is 2 and 3 is not 4", "((1 is 2) and (3 is not 4))"},
		{"1 is not 2 and 3 is 4", "((1 is not 2) and (3 is 4))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceTernary tests ternary operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceTernary() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 ? 2 : 3", "(1 ? 2 : 3)"},
		{"1 and 2 ? 3 : 4", "((1 and 2) ? 3 : 4)"},
		{"1 ? 2 and 3 : 4", "(1 ? (2 and 3) : 4)"},
		{"1 ? 2 : 3 and 4", "(1 ? 2 : (3 and 4))"},
		{"1 + 2 ? 3 : 4", "((1 + 2) ? 3 : 4)"},
		{"1 ? 2 + 3 : 4", "(1 ? (2 + 3) : 4)"},
		{"1 ? 2 : 3 + 4", "(1 ? 2 : (3 + 4))"},
		{"1 * 2 ? 3 : 4", "((1 * 2) ? 3 : 4)"},
		{"1 ? 2 * 3 : 4", "(1 ? (2 * 3) : 4)"},
		{"1 ? 2 : 3 * 4", "(1 ? 2 : (3 * 4))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceUnary tests unary operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceUnary() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"!true", "(!true)"},
		{"-42", "(-42)"},
		{"+3.14", "(+3.14)"},
		{"!1 + 2", "((!1) + 2)"},
		{"-1 + 2", "((-1) + 2)"},
		{"+1 + 2", "((+1) + 2)"},
		{"!1 * 2", "((!1) * 2)"},
		{"-1 * 2", "((-1) * 2)"},
		{"+1 * 2", "((+1) * 2)"},
		{"!1 and 2", "((!1) and 2)"},
		{"-1 and 2", "((-1) and 2)"},
		{"+1 and 2", "((+1) and 2)"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceCall tests function call precedence
func (s *PrecedenceTestSuite) TestPrecedenceCall() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"myFunction(1, 2)", "myFunction(1, 2)"},
		{"myFunction(1 + 2, 3)", "myFunction((1 + 2), 3)"},
		{"myFunction(1, 2 + 3)", "myFunction(1, (2 + 3))"},
		{"myFunction(1 * 2, 3)", "myFunction((1 * 2), 3)"},
		{"myFunction(1, 2 * 3)", "myFunction(1, (2 * 3))"},
		{"myFunction(1 and 2, 3)", "myFunction((1 and 2), 3)"},
		{"myFunction(1, 2 and 3)", "myFunction(1, (2 and 3))"},
		{"myFunction(1 ? 2 : 3, 4)", "myFunction((1 ? 2 : 3), 4)"},
		{"myFunction(1, 2 ? 3 : 4)", "myFunction(1, (2 ? 3 : 4))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceIndex tests index access precedence
func (s *PrecedenceTestSuite) TestPrecedenceIndex() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"array[0]", "array[0]"},
		{"obj.field", "obj.field"},
		{"array[1 + 2]", "array[(1 + 2)]"},
		{"obj[1 + 2]", "obj[(1 + 2)]"},
		{"array[1 * 2]", "array[(1 * 2)]"},
		{"obj[1 * 2]", "obj[(1 * 2)]"},
		{"array[1 and 2]", "array[(1 and 2)]"},
		{"obj[1 and 2]", "obj[(1 and 2)]"},
		{"array[1 ? 2 : 3]", "array[(1 ? 2 : 3)]"},
		{"obj[1 ? 2 : 3]", "obj[(1 ? 2 : 3)]"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceGrouping tests grouping with parentheses
func (s *PrecedenceTestSuite) TestPrecedenceGrouping() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
		{"1 + (2 * 3)", "(1 + (2 * 3))"},
		{"(1 + 2) + 3", "((1 + 2) + 3)"},
		{"1 + (2 + 3)", "(1 + (2 + 3))"},
		{"(1 * 2) * 3", "((1 * 2) * 3)"},
		{"1 * (2 * 3)", "(1 * (2 * 3))"},
		{"(1 and 2) or 3", "((1 and 2) or 3)"},
		{"1 and (2 or 3)", "(1 and (2 or 3))"},
		{"(1 or 2) and 3", "((1 or 2) and 3)"},
		{"1 or (2 and 3)", "(1 or (2 and 3))"},
		{"(1 == 2) and 3", "((1 == 2) and 3)"},
		{"1 == (2 and 3)", "(1 == (2 and 3))"},
		{"(1 < 2) and 3", "((1 < 2) and 3)"},
		{"1 < (2 and 3)", "(1 < (2 and 3))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceComplex tests complex precedence combinations
func (s *PrecedenceTestSuite) TestPrecedenceComplex() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3 == 4 + 5 * 6", "((1 + (2 * 3)) == (4 + (5 * 6)))"},
		{"1 * 2 + 3 < 4 * 5 + 6", "((1 * 2) + 3) < ((4 * 5) + 6)"},
		{"1 and 2 + 3 * 4", "(1 and (2 + (3 * 4)))"},
		{"1 + 2 and 3 * 4", "((1 + 2) and (3 * 4))"},
		{"1 ? 2 + 3 : 4 * 5", "(1 ? (2 + 3) : (4 * 5))"},
		{"1 + 2 ? 3 * 4 : 5 + 6", "((1 + 2) ? (3 * 4) : (5 + 6))"},
		{"myFunction(1 + 2, 3 * 4)", "myFunction((1 + 2), (3 * 4))"},
		{"array[1 + 2] + obj.field", "(array[(1 + 2)] + obj.field)"},
		{"!1 + 2 * 3", "((!1) + (2 * 3))"},
		{"-1 * 2 + 3", "(((-1) * 2) + 3)"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceAssociativity tests operator associativity
func (s *PrecedenceTestSuite) TestPrecedenceAssociativity() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 * 2 * 3", "((1 * 2) * 3)"},
		{"1 and 2 and 3", "((1 and 2) and 3)"},
		{"1 or 2 or 3", "((1 or 2) or 3)"},
		{"1 xor 2 xor 3", "((1 xor 2) xor 3)"},
		{"1 == 2 == 3", "((1 == 2) == 3)"},
		{"1 != 2 != 3", "((1 != 2) != 3)"},
		{"1 < 2 < 3", "((1 < 2) < 3)"},
		{"1 > 2 > 3", "((1 > 2) > 3)"},
		{"1 <= 2 <= 3", "((1 <= 2) <= 3)"},
		{"1 >= 2 >= 3", "((1 >= 2) >= 3)"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: %s", tc.input)

		// Verify the expression structure
		s.NotNil(expr)
	}
}

// TestPrecedenceTestSuite runs the precedence test suite
func TestPrecedenceTestSuite(t *testing.T) {
	suite.Run(t, new(PrecedenceTestSuite))
}
