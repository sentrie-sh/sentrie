// SPDX-License-Identifier: Apache-2.0

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
	s.T().Run("AdditionMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 * 3")
		s.Equal("(1 + (2 * 3))", expr.String())
	})

	s.T().Run("MultiplicationAddition", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 + 3")
		s.Equal("((1 * 2) + 3)", expr.String())
	})

	s.T().Run("AdditionChain", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 + 3")
		s.Equal("((1 + 2) + 3)", expr.String())
	})

	s.T().Run("MultiplicationChain", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 * 3")
		s.Equal("((1 * 2) * 3)", expr.String())
	})

	s.T().Run("MixedAdditionMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 * 3 + 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 * 3 + 4")
		s.Equal("((1 + (2 * 3)) + 4)", expr.String())
	})

	s.T().Run("MixedMultiplicationAddition", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 + 3 * 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 + 3 * 4")
		s.Equal("((1 * 2) + (3 * 4))", expr.String())
	})

	s.T().Run("AdditionDivision", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 / 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 / 3")
		s.Equal("(1 + (2 / 3))", expr.String())
	})

	s.T().Run("DivisionAddition", func(t *testing.T) {
		parser := NewParserFromString("1 / 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 / 2 + 3")
		s.Equal("((1 / 2) + 3)", expr.String())
	})

	s.T().Run("AdditionModulo", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 % 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 % 3")
		s.Equal("(1 + (2 % 3))", expr.String())
	})

	s.T().Run("ModuloAddition", func(t *testing.T) {
		parser := NewParserFromString("1 % 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 % 2 + 3")
		s.Equal("((1 % 2) + 3)", expr.String())
	})

	s.T().Run("SubtractionMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 - 2 * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 - 2 * 3")
		s.Equal("(1 - (2 * 3))", expr.String())
	})

	s.T().Run("MultiplicationSubtraction", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 - 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 - 3")
		s.Equal("((1 * 2) - 3)", expr.String())
	})
}

// TestPrecedenceComparison tests comparison operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceComparison() {
	s.T().Run("LessThanAndGreaterThan", func(t *testing.T) {
		parser := NewParserFromString("1 < 2 and 3 > 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 < 2 and 3 > 4")
		s.Equal("((1 < 2) and (3 > 4))", expr.String())
	})

	s.T().Run("LessEqualAndGreaterEqual", func(t *testing.T) {
		parser := NewParserFromString("1 <= 2 and 3 >= 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 <= 2 and 3 >= 4")
		s.Equal("((1 <= 2) and (3 >= 4))", expr.String())
	})

	s.T().Run("EqualAndNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 and 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 and 3 != 4")
		s.Equal("((1 == 2) and (3 != 4))", expr.String())
	})

	s.T().Run("LessThanOrGreaterThan", func(t *testing.T) {
		parser := NewParserFromString("1 < 2 or 3 > 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 < 2 or 3 > 4")
		s.Equal("((1 < 2) or (3 > 4))", expr.String())
	})

	s.T().Run("LessEqualOrGreaterEqual", func(t *testing.T) {
		parser := NewParserFromString("1 <= 2 or 3 >= 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 <= 2 or 3 >= 4")
		s.Equal("((1 <= 2) or (3 >= 4))", expr.String())
	})

	s.T().Run("EqualOrNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 or 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 or 3 != 4")
		s.Equal("((1 == 2) or (3 != 4))", expr.String())
	})

	s.T().Run("LessThanXorGreaterThan", func(t *testing.T) {
		parser := NewParserFromString("1 < 2 xor 3 > 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 < 2 xor 3 > 4")
		s.Equal("((1 < 2) xor (3 > 4))", expr.String())
	})

	s.T().Run("LessEqualXorGreaterEqual", func(t *testing.T) {
		parser := NewParserFromString("1 <= 2 xor 3 >= 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 <= 2 xor 3 >= 4")
		s.Equal("((1 <= 2) xor (3 >= 4))", expr.String())
	})

	s.T().Run("EqualXorNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 xor 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 xor 3 != 4")
		s.Equal("((1 == 2) xor (3 != 4))", expr.String())
	})
}

// TestPrecedenceLogical tests logical operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceLogical() {
	s.T().Run("AndOr", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 or 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 or 3")
		s.Equal("((1 and 2) or 3)", expr.String())
	})

	s.T().Run("OrAnd", func(t *testing.T) {
		parser := NewParserFromString("1 or 2 and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 or 2 and 3")
		s.Equal("(1 or (2 and 3))", expr.String())
	})

	s.T().Run("AndXor", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 xor 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 xor 3")
		s.Equal("((1 and 2) xor 3)", expr.String())
	})

	s.T().Run("XorAnd", func(t *testing.T) {
		parser := NewParserFromString("1 xor 2 and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 xor 2 and 3")
		s.Equal("(1 xor (2 and 3))", expr.String())
	})

	s.T().Run("OrXor", func(t *testing.T) {
		parser := NewParserFromString("1 or 2 xor 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 or 2 xor 3")
		s.Equal("(1 or (2 xor 3))", expr.String())
	})

	s.T().Run("XorOr", func(t *testing.T) {
		parser := NewParserFromString("1 xor 2 or 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 xor 2 or 3")
		s.Equal("((1 xor 2) or 3)", expr.String())
	})

	s.T().Run("AndChain", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 and 3")
		s.Equal("((1 and 2) and 3)", expr.String())
	})

	s.T().Run("OrChain", func(t *testing.T) {
		parser := NewParserFromString("1 or 2 or 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 or 2 or 3")
		s.Equal("((1 or 2) or 3)", expr.String())
	})

	s.T().Run("XorChain", func(t *testing.T) {
		parser := NewParserFromString("1 xor 2 xor 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 xor 2 xor 3")
		s.Equal("((1 xor 2) xor 3)", expr.String())
	})
}

// TestPrecedenceEquality tests equality operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceEquality() {
	s.T().Run("EqualAndNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 and 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 and 3 != 4")
		s.Equal("((1 == 2) and (3 != 4))", expr.String())
	})

	s.T().Run("NotEqualAndEqual", func(t *testing.T) {
		parser := NewParserFromString("1 != 2 and 3 == 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 != 2 and 3 == 4")
		s.Equal("((1 != 2) and (3 == 4))", expr.String())
	})

	s.T().Run("EqualOrNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 or 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 or 3 != 4")
		s.Equal("((1 == 2) or (3 != 4))", expr.String())
	})

	s.T().Run("NotEqualOrEqual", func(t *testing.T) {
		parser := NewParserFromString("1 != 2 or 3 == 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 != 2 or 3 == 4")
		s.Equal("((1 != 2) or (3 == 4))", expr.String())
	})

	s.T().Run("EqualXorNotEqual", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 xor 3 != 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 xor 3 != 4")
		s.Equal("((1 == 2) xor (3 != 4))", expr.String())
	})

	s.T().Run("NotEqualXorEqual", func(t *testing.T) {
		parser := NewParserFromString("1 != 2 xor 3 == 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 != 2 xor 3 == 4")
		s.Equal("((1 != 2) xor (3 == 4))", expr.String())
	})

	s.T().Run("IsAndIsNot", func(t *testing.T) {
		parser := NewParserFromString("1 is 2 and 3 is not 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 is 2 and 3 is not 4")
		s.Equal("((1 is 2) and not(3 is 4))", expr.String())
	})

	s.T().Run("IsNotAndIs", func(t *testing.T) {
		parser := NewParserFromString("1 is not 2 and 3 is 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 is not 2 and 3 is 4")
		s.Equal("(not(1 is 2) and (3 is 4))", expr.String())
	})
}

// TestPrecedenceTernary tests ternary operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceTernary() {
	s.T().Run("BasicTernary", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 : 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 : 3")
		s.Equal("(1 ? 2 : 3)", expr.String())
	})

	s.T().Run("AndTernary", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 ? 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 ? 3 : 4")
		s.Equal("((1 and 2) ? 3 : 4)", expr.String())
	})

	s.T().Run("TernaryAnd", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 and 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 and 3 : 4")
		s.Equal("(1 ? (2 and 3) : 4)", expr.String())
	})

	s.T().Run("TernaryAndResult", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 : 3 and 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 : 3 and 4")
		s.Equal("(1 ? 2 : (3 and 4))", expr.String())
	})

	s.T().Run("AdditionTernary", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 ? 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 ? 3 : 4")
		s.Equal("((1 + 2) ? 3 : 4)", expr.String())
	})

	s.T().Run("TernaryAddition", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 + 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 + 3 : 4")
		s.Equal("(1 ? (2 + 3) : 4)", expr.String())
	})

	s.T().Run("TernaryAdditionResult", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 : 3 + 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 : 3 + 4")
		s.Equal("(1 ? 2 : (3 + 4))", expr.String())
	})

	s.T().Run("MultiplicationTernary", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 ? 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 ? 3 : 4")
		s.Equal("((1 * 2) ? 3 : 4)", expr.String())
	})

	s.T().Run("TernaryMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 * 3 : 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 * 3 : 4")
		s.Equal("(1 ? (2 * 3) : 4)", expr.String())
	})

	s.T().Run("TernaryMultiplicationResult", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 : 3 * 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 : 3 * 4")
		s.Equal("(1 ? 2 : (3 * 4))", expr.String())
	})
}

// TestPrecedenceUnary tests unary operator precedence
func (s *PrecedenceTestSuite) TestPrecedenceUnary() {
	s.T().Run("NotTrue", func(t *testing.T) {
		parser := NewParserFromString("!true", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: !true")
		s.Equal("!true", expr.String())
	})

	s.T().Run("NegativeInt", func(t *testing.T) {
		parser := NewParserFromString("-42", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: -42")
		s.Equal("-42", expr.String())
	})

	s.T().Run("PositiveFloat", func(t *testing.T) {
		parser := NewParserFromString("+3.14", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: +3.14")
		s.Equal("+3.14", expr.String())
	})

	s.T().Run("NotAddition", func(t *testing.T) {
		parser := NewParserFromString("!1 + 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: !1 + 2")
		s.Equal("(!1 + 2)", expr.String())
	})

	s.T().Run("NegativeAddition", func(t *testing.T) {
		parser := NewParserFromString("-1 + 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: -1 + 2")
		s.Equal("(-1 + 2)", expr.String())
	})

	s.T().Run("PositiveAddition", func(t *testing.T) {
		parser := NewParserFromString("+1 + 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: +1 + 2")
		s.Equal("(+1 + 2)", expr.String())
	})

	s.T().Run("NotMultiplication", func(t *testing.T) {
		parser := NewParserFromString("!1 * 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: !1 * 2")
		s.Equal("(!1 * 2)", expr.String())
	})

	s.T().Run("NegativeMultiplication", func(t *testing.T) {
		parser := NewParserFromString("-1 * 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: -1 * 2")
		s.Equal("(-1 * 2)", expr.String())
	})

	s.T().Run("PositiveMultiplication", func(t *testing.T) {
		parser := NewParserFromString("+1 * 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: +1 * 2")
		s.Equal("(+1 * 2)", expr.String())
	})

	s.T().Run("NotAnd", func(t *testing.T) {
		parser := NewParserFromString("!1 and 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: !1 and 2")
		s.Equal("(!1 and 2)", expr.String())
	})

	s.T().Run("NegativeAnd", func(t *testing.T) {
		parser := NewParserFromString("-1 and 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: -1 and 2")
		s.Equal("(-1 and 2)", expr.String())
	})

	s.T().Run("PositiveAnd", func(t *testing.T) {
		parser := NewParserFromString("+1 and 2", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: +1 and 2")
		s.Equal("(+1 and 2)", expr.String())
	})
}

// TestPrecedenceCall tests function call precedence
func (s *PrecedenceTestSuite) TestPrecedenceCall() {
	s.T().Run("BasicCall", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1, 2)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1, 2)")
		s.Equal("myFunction(1, 2)", expr.String())
	})

	s.T().Run("CallWithAddition", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1 + 2, 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1 + 2, 3)")
		s.Equal("myFunction((1 + 2), 3)", expr.String())
	})

	s.T().Run("CallWithAdditionSecond", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1, 2 + 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1, 2 + 3)")
		s.Equal("myFunction(1, (2 + 3))", expr.String())
	})

	s.T().Run("CallWithMultiplication", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1 * 2, 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1 * 2, 3)")
		s.Equal("myFunction((1 * 2), 3)", expr.String())
	})

	s.T().Run("CallWithMultiplicationSecond", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1, 2 * 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1, 2 * 3)")
		s.Equal("myFunction(1, (2 * 3))", expr.String())
	})

	s.T().Run("CallWithAnd", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1 and 2, 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1 and 2, 3)")
		s.Equal("myFunction((1 and 2), 3)", expr.String())
	})

	s.T().Run("CallWithAndSecond", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1, 2 and 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1, 2 and 3)")
		s.Equal("myFunction(1, (2 and 3))", expr.String())
	})

	s.T().Run("CallWithTernary", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1 ? 2 : 3, 4)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1 ? 2 : 3, 4)")
		s.Equal("myFunction((1 ? 2 : 3), 4)", expr.String())
	})

	s.T().Run("CallWithTernarySecond", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1, 2 ? 3 : 4)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1, 2 ? 3 : 4)")
		s.Equal("myFunction(1, (2 ? 3 : 4))", expr.String())
	})
}

// TestPrecedenceIndex tests index access precedence
func (s *PrecedenceTestSuite) TestPrecedenceIndex() {
	s.T().Run("ArrayIndex", func(t *testing.T) {
		parser := NewParserFromString("array[0]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[0]")
		s.Equal("array[0]", expr.String())
	})

	s.T().Run("ObjectField", func(t *testing.T) {
		parser := NewParserFromString("obj.field", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: obj.field")
		s.Equal("obj.field", expr.String())
	})

	s.T().Run("ArrayIndexAddition", func(t *testing.T) {
		parser := NewParserFromString("array[1 + 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[1 + 2]")
		s.Equal("array[(1 + 2)]", expr.String())
	})

	s.T().Run("ObjectIndexAddition", func(t *testing.T) {
		parser := NewParserFromString("obj[1 + 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: obj[1 + 2]")
		s.Equal("obj[(1 + 2)]", expr.String())
	})

	s.T().Run("ArrayIndexMultiplication", func(t *testing.T) {
		parser := NewParserFromString("array[1 * 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[1 * 2]")
		s.Equal("array[(1 * 2)]", expr.String())
	})

	s.T().Run("ObjectIndexMultiplication", func(t *testing.T) {
		parser := NewParserFromString("obj[1 * 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: obj[1 * 2]")
		s.Equal("obj[(1 * 2)]", expr.String())
	})

	s.T().Run("ArrayIndexAnd", func(t *testing.T) {
		parser := NewParserFromString("array[1 and 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[1 and 2]")
		s.Equal("array[(1 and 2)]", expr.String())
	})

	s.T().Run("ObjectIndexAnd", func(t *testing.T) {
		parser := NewParserFromString("obj[1 and 2]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: obj[1 and 2]")
		s.Equal("obj[(1 and 2)]", expr.String())
	})

	s.T().Run("ArrayIndexTernary", func(t *testing.T) {
		parser := NewParserFromString("array[1 ? 2 : 3]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[1 ? 2 : 3]")
		s.Equal("array[(1 ? 2 : 3)]", expr.String())
	})

	s.T().Run("ObjectIndexTernary", func(t *testing.T) {
		parser := NewParserFromString("obj[1 ? 2 : 3]", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: obj[1 ? 2 : 3]")
		s.Equal("obj[(1 ? 2 : 3)]", expr.String())
	})
}

// TestPrecedenceGrouping tests grouping with parentheses
func (s *PrecedenceTestSuite) TestPrecedenceGrouping() {
	s.T().Run("GroupedAdditionMultiplication", func(t *testing.T) {
		parser := NewParserFromString("(1 + 2) * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 + 2) * 3")
		s.Equal("((1 + 2) * 3)", expr.String())
	})

	s.T().Run("AdditionGroupedMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 + (2 * 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + (2 * 3)")
		s.Equal("(1 + (2 * 3))", expr.String())
	})

	s.T().Run("GroupedAdditionAddition", func(t *testing.T) {
		parser := NewParserFromString("(1 + 2) + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 + 2) + 3")
		s.Equal("((1 + 2) + 3)", expr.String())
	})

	s.T().Run("AdditionGroupedAddition", func(t *testing.T) {
		parser := NewParserFromString("1 + (2 + 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + (2 + 3)")
		s.Equal("(1 + (2 + 3))", expr.String())
	})

	s.T().Run("GroupedMultiplicationMultiplication", func(t *testing.T) {
		parser := NewParserFromString("(1 * 2) * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 * 2) * 3")
		s.Equal("((1 * 2) * 3)", expr.String())
	})

	s.T().Run("MultiplicationGroupedMultiplication", func(t *testing.T) {
		parser := NewParserFromString("1 * (2 * 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * (2 * 3)")
		s.Equal("(1 * (2 * 3))", expr.String())
	})

	s.T().Run("GroupedAndOr", func(t *testing.T) {
		parser := NewParserFromString("(1 and 2) or 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 and 2) or 3")
		s.Equal("((1 and 2) or 3)", expr.String())
	})

	s.T().Run("AndGroupedOr", func(t *testing.T) {
		parser := NewParserFromString("1 and (2 or 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and (2 or 3)")
		s.Equal("(1 and (2 or 3))", expr.String())
	})

	s.T().Run("GroupedOrAnd", func(t *testing.T) {
		parser := NewParserFromString("(1 or 2) and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 or 2) and 3")
		s.Equal("((1 or 2) and 3)", expr.String())
	})

	s.T().Run("OrGroupedAnd", func(t *testing.T) {
		parser := NewParserFromString("1 or (2 and 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 or (2 and 3)")
		s.Equal("(1 or (2 and 3))", expr.String())
	})

	s.T().Run("GroupedEqualAnd", func(t *testing.T) {
		parser := NewParserFromString("(1 == 2) and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 == 2) and 3")
		s.Equal("((1 == 2) and 3)", expr.String())
	})

	s.T().Run("EqualGroupedAnd", func(t *testing.T) {
		parser := NewParserFromString("1 == (2 and 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == (2 and 3)")
		s.Equal("(1 == (2 and 3))", expr.String())
	})

	s.T().Run("GroupedLessThanAnd", func(t *testing.T) {
		parser := NewParserFromString("(1 < 2) and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: (1 < 2) and 3")
		s.Equal("((1 < 2) and 3)", expr.String())
	})

	s.T().Run("LessThanGroupedAnd", func(t *testing.T) {
		parser := NewParserFromString("1 < (2 and 3)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 < (2 and 3)")
		s.Equal("(1 < (2 and 3))", expr.String())
	})
}

// TestPrecedenceComplex tests complex precedence combinations
func (s *PrecedenceTestSuite) TestPrecedenceComplex() {
	s.T().Run("ComplexArithmeticEquality", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 * 3 == 4 + 5 * 6", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 * 3 == 4 + 5 * 6")
		s.Equal("((1 + (2 * 3)) == (4 + (5 * 6)))", expr.String())
	})

	s.T().Run("ComplexArithmeticComparison", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 + 3 < 4 * 5 + 6", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 + 3 < 4 * 5 + 6")
		s.Equal("(((1 * 2) + 3) < ((4 * 5) + 6))", expr.String())
	})

	s.T().Run("AndArithmetic", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 + 3 * 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 + 3 * 4")
		s.Equal("(1 and (2 + (3 * 4)))", expr.String())
	})

	s.T().Run("ArithmeticAnd", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 and 3 * 4", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 and 3 * 4")
		s.Equal("((1 + 2) and (3 * 4))", expr.String())
	})

	s.T().Run("TernaryArithmetic", func(t *testing.T) {
		parser := NewParserFromString("1 ? 2 + 3 : 4 * 5", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 ? 2 + 3 : 4 * 5")
		s.Equal("(1 ? (2 + 3) : (4 * 5))", expr.String())
	})

	s.T().Run("ArithmeticTernary", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 ? 3 * 4 : 5 + 6", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 ? 3 * 4 : 5 + 6")
		s.Equal("((1 + 2) ? (3 * 4) : (5 + 6))", expr.String())
	})

	s.T().Run("FunctionArithmetic", func(t *testing.T) {
		parser := NewParserFromString("myFunction(1 + 2, 3 * 4)", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: myFunction(1 + 2, 3 * 4)")
		s.Equal("myFunction((1 + 2), (3 * 4))", expr.String())
	})

	s.T().Run("IndexArithmetic", func(t *testing.T) {
		parser := NewParserFromString("array[1 + 2] + obj.field", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: array[1 + 2] + obj.field")
		s.Equal("(array[(1 + 2)] + obj.field)", expr.String())
	})

	s.T().Run("NotArithmetic", func(t *testing.T) {
		parser := NewParserFromString("!1 + 2 * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: !1 + 2 * 3")
		s.Equal("(!1 + (2 * 3))", expr.String())
	})

	s.T().Run("NegativeArithmetic", func(t *testing.T) {
		parser := NewParserFromString("-1 * 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: -1 * 2 + 3")
		s.Equal("((-1 * 2) + 3)", expr.String())
	})
}

// TestPrecedenceAssociativity tests operator associativity
func (s *PrecedenceTestSuite) TestPrecedenceAssociativity() {
	s.T().Run("AdditionChain", func(t *testing.T) {
		parser := NewParserFromString("1 + 2 + 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 + 2 + 3")
		s.Equal("((1 + 2) + 3)", expr.String())
	})

	s.T().Run("MultiplicationChain", func(t *testing.T) {
		parser := NewParserFromString("1 * 2 * 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 * 2 * 3")
		s.Equal("((1 * 2) * 3)", expr.String())
	})

	s.T().Run("AndChain", func(t *testing.T) {
		parser := NewParserFromString("1 and 2 and 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 and 2 and 3")
		s.Equal("((1 and 2) and 3)", expr.String())
	})

	s.T().Run("OrChain", func(t *testing.T) {
		parser := NewParserFromString("1 or 2 or 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 or 2 or 3")
		s.Equal("((1 or 2) or 3)", expr.String())
	})

	s.T().Run("XorChain", func(t *testing.T) {
		parser := NewParserFromString("1 xor 2 xor 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 xor 2 xor 3")
		s.Equal("((1 xor 2) xor 3)", expr.String())
	})

	s.T().Run("EqualChain", func(t *testing.T) {
		parser := NewParserFromString("1 == 2 == 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 == 2 == 3")
		s.Equal("((1 == 2) == 3)", expr.String())
	})

	s.T().Run("NotEqualChain", func(t *testing.T) {
		parser := NewParserFromString("1 != 2 != 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 != 2 != 3")
		s.Equal("((1 != 2) != 3)", expr.String())
	})

	s.T().Run("LessThanChain", func(t *testing.T) {
		parser := NewParserFromString("1 < 2 < 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 < 2 < 3")
		s.Equal("((1 < 2) < 3)", expr.String())
	})

	s.T().Run("GreaterThanChain", func(t *testing.T) {
		parser := NewParserFromString("1 > 2 > 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 > 2 > 3")
		s.Equal("((1 > 2) > 3)", expr.String())
	})

	s.T().Run("LessEqualChain", func(t *testing.T) {
		parser := NewParserFromString("1 <= 2 <= 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 <= 2 <= 3")
		s.Equal("((1 <= 2) <= 3)", expr.String())
	})

	s.T().Run("GreaterEqualChain", func(t *testing.T) {
		parser := NewParserFromString("1 >= 2 >= 3", "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "Failed to parse: 1 >= 2 >= 3")
		s.Equal("((1 >= 2) >= 3)", expr.String())
	})
}

// TestPrecedenceTestSuite runs the precedence test suite
func TestPrecedenceTestSuite(t *testing.T) {
	suite.Run(t, new(PrecedenceTestSuite))
}
