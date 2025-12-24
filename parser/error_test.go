// SPDX-License-Identifier: Apache-2.0
//
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

// ErrorTestSuite provides tests for error handling
type ErrorTestSuite struct {
	suite.Suite
}

// SetupSuite initializes the test suite
func (s *ErrorTestSuite) SetupSuite() {
	slog.Info("ErrorTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *ErrorTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "ErrorTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *ErrorTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "ErrorTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *ErrorTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestParseErrorEmptyInput tests parsing empty input
func (s *ErrorTestSuite) TestParseErrorEmptyInput() {
	parser := NewParserFromString("", "test.sentra")
	_, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Empty input should not error")
}

// TestParseErrorWhitespaceOnly tests parsing whitespace-only input
func (s *ErrorTestSuite) TestParseErrorWhitespaceOnly() {
	parser := NewParserFromString("   \n\t   ", "test.sentra")
	_, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Whitespace-only input should not error")
}

// TestParseErrorCommentOnly tests parsing comment-only input
func (s *ErrorTestSuite) TestParseErrorCommentOnly() {
	parser := NewParserFromString("-- This is a comment", "test.sentra")
	_, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Comment-only input should not error")
}

// TestParseErrorUnexpectedTokens tests parsing with unexpected tokens
func (s *ErrorTestSuite) TestParseErrorUnexpectedTokens() {
	testCases := []string{
		"rule check { true }",       // Rule at top level
		"fact name:string",          // Fact at top level
		"use fn from @lib as alias", // Use at top level
		"export rule check",         // Export at top level
		"shape User { }",            // Shape before namespace
		"policy user { }",           // Policy before namespace
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for: %s", tc)
	}
}

// TestParseErrorIncompleteStatement tests parsing incomplete statements
func (s *ErrorTestSuite) TestParseErrorIncompleteStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"namespace", "expected Ident, got EOF at"},
		{"policy", "expected Ident, got EOF"},
		{"rule", "unexpected token"},
		{"fact", "unexpected token"},
		{"shape", "expected Ident, got EOF"},
		{"export", "expected 'shape', got EOF"},
		{"import", "unexpected token"},
		{"use", "unexpected token"},
		{"when", "unexpected token"},
		{"default", "unexpected token"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for: %s", tc.input)
		s.Contains(err.Error(), tc.expected, "Error message should contain: %s", tc.expected)
	}
}

// TestParseErrorMismatchedBrackets tests parsing with mismatched brackets
func (s *ErrorTestSuite) TestParseErrorMismatchedBrackets() {
	testCases := []string{
		"namespace com/example; policy user {",     // Missing closing brace
		"namespace com/example; policy user { } }", // Extra closing brace
		"namespace com/example; shape User {",      // Missing closing brace
		"namespace com/example; rule check { true", // Missing closing brace
		"namespace com/example; (x + y",            // Missing closing parenthesis
		"namespace com/example; [1, 2, 3",          // Missing closing bracket
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for mismatched brackets: %s", tc)
	}
}

// TestParseErrorInvalidIdentifiers tests parsing with invalid identifiers
func (s *ErrorTestSuite) TestParseErrorInvalidIdentifiers() {
	testCases := []string{
		"namespace 123;",                         // Invalid namespace identifier
		"namespace com/123;",                     // Invalid namespace segment
		"namespace com/example; policy 123 { }",  // Invalid policy identifier
		"namespace com/example; rule 123 { }",    // Invalid rule identifier
		"namespace com/example; fact 123:string", // Invalid fact identifier
		"namespace com/example; shape 123 { }",   // Invalid shape identifier
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid identifier: %s", tc)
	}
}

// TestParseErrorInvalidExpressions tests parsing with invalid expressions
func (s *ErrorTestSuite) TestParseErrorInvalidExpressions() {
	testCases := []string{
		"namespace com/example; rule check { + }",       // Invalid unary operator
		"namespace com/example; rule check { 1 + }",     // Incomplete binary expression
		"namespace com/example; rule check { (1 + }",    // Mismatched parentheses
		"namespace com/example; rule check { 1 + + 2 }", // Invalid operator sequence
		"namespace com/example; rule check { .field }",  // Invalid field access
		"namespace com/example; rule check { [1, 2, }",  // Invalid array literal
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid expression: %s", tc)
	}
}

// TestParseErrorInvalidTypes tests parsing with invalid types
func (s *ErrorTestSuite) TestParseErrorInvalidTypes() {
	testCases := []string{
		"namespace com/example; fact name:123;",    // Invalid type
		"namespace com/example; fact name:;",       // Missing type
		"namespace com/example; shape User = 123;", // Invalid type reference
		"namespace com/example; shape User = ;",    // Missing type reference
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid type: %s", tc)
	}
}

// TestParseErrorInvalidLiterals tests parsing with invalid literals
func (s *ErrorTestSuite) TestParseErrorInvalidLiterals() {
	testCases := []string{
		"namespace com/example; fact name:string default \"unclosed;", // Unclosed string
		"namespace com/example; fact name:string default 'unclosed;",  // Unclosed string
		"namespace com/example; fact name:string default 123abc;",     // Invalid number
		"namespace com/example; fact name:string default true;",       // Invalid boolean (not supported)
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid literal: %s", tc)
	}
}

// TestParseErrorInvalidOperators tests parsing with invalid operators
func (s *ErrorTestSuite) TestParseErrorInvalidOperators() {
	testCases := []string{
		"namespace com/example; rule check { 1 ** 2 }",  // Invalid operator
		"namespace com/example; rule check { 1 // 2 }",  // Invalid operator
		"namespace com/example; rule check { 1 %% 2 }",  // Invalid operator
		"namespace com/example; rule check { 1 &&& 2 }", // Invalid operator
		"namespace com/example; rule check { 1 ||| 2 }", // Invalid operator
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid operator: %s", tc)
	}
}

// TestParseErrorInvalidKeywords tests parsing with invalid keywords
func (s *ErrorTestSuite) TestParseErrorInvalidKeywords() {
	testCases := []string{
		"namespace com/example; invalid { }",                       // Invalid keyword
		"namespace com/example; policy user { invalid { } }",       // Invalid keyword in policy
		"namespace com/example; rule check { invalid }",            // Invalid keyword in rule
		"namespace com/example; fact name:string default invalid;", // Invalid keyword in fact
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid keyword: %s", tc)
	}
}

// TestParseErrorInvalidSyntax tests parsing with invalid syntax
func (s *ErrorTestSuite) TestParseErrorInvalidSyntax() {
	testCases := []string{
		"namespace com/example; policy user { rule check { true } fact name:string; }",  // Missing semicolon
		"namespace com/example; policy user { rule check { true }; fact name:string }",  // Missing semicolon
		"namespace com/example; policy user { rule check { true } fact name:string; }",  // Missing semicolon
		"namespace com/example; policy user { rule check { true }; fact name:string; }", // Missing semicolon
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for invalid syntax: %s", tc)
	}
}

// TestParseErrorEdgeCases tests parsing edge cases
func (s *ErrorTestSuite) TestParseErrorEdgeCases() {
	testCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		{"namespace com/example;", false, "Valid namespace"},
		{"namespace com/example; policy user { }", false, "Valid namespace and policy"},
		{"namespace com/example; shape User { }", false, "Valid namespace and shape"},
		{"namespace com/example; rule check { true }", true, "Rule at top level"},
		{"namespace com/example; fact name:string", true, "Fact at top level"},
		{"namespace com/example; use fn from @lib as alias", true, "Use at top level"},
		{"namespace com/example; export rule check", true, "Export at top level"},
		{"namespace com/example; import from @lib", true, "Import at top level"},
		{"namespace com/example; when condition { }", true, "When at top level"},
		{"namespace com/example; default { }", true, "Default at top level"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		_, err := parser.ParseProgram(s.T().Context())

		if tc.shouldError {
			s.Error(err, "Expected error for: %s (%s)", tc.input, tc.description)
		} else {
			s.NoError(err, "Expected no error for: %s (%s)", tc.input, tc.description)
		}
	}
}

// TestErrorTestSuite runs the error test suite
func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}
