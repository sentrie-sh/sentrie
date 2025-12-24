// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License")
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
	"github.com/stretchr/testify/suite"
)

// ProgramTestSuite provides tests for program parsing
type ProgramTestSuite struct {
	suite.Suite
}

// SetupSuite initializes the test suite
func (s *ProgramTestSuite) SetupSuite() {
	slog.Info("ProgramTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *ProgramTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "ProgramTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *ProgramTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "ProgramTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *ProgramTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestParseProgramBasic tests parsing basic programs
func (s *ProgramTestSuite) TestParseProgramBasic() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"namespace com/example", "com/example"},
		{"namespace test", "test"},
		{"namespace com/example/test", "com/example/test"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		program, err := parser.ParseProgram(s.T().Context())
		s.NoError(err, "Expected no error for: %s", tc.input)
		s.NotNil(program, "Expected program for: %s", tc.input)
		// Check namespace (first statement should be namespace)
		s.Greater(len(program.Statements), 0, "Expected statements in program")
		namespaceStmt, ok := program.Statements[0].(*ast.NamespaceStatement)
		s.True(ok, "Expected first statement to be namespace")
		s.Equal(tc.expected, namespaceStmt.Name.String(), "Expected namespace: %s", tc.expected)
	}
}

// TestParseProgramWithPolicies tests parsing programs with policies
func (s *ProgramTestSuite) TestParseProgramWithPolicies() {
	input := `
namespace com/example
policy user {
	rule check = { yield true }
	fact name:string default "john"
}
policy admin {
	rule validate = { yield x > 0 }
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for program with policies")
	s.NotNil(program, "Expected program for program with policies")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	namespaceStmt, ok := program.Statements[0].(*ast.NamespaceStatement)
	s.True(ok, "Expected first statement to be namespace")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseProgramWithShapes tests parsing programs with shapes
func (s *ProgramTestSuite) TestParseProgramWithShapes() {
	input := `
namespace com/example
shape User {
	name:string
	age:int
}
shape Person string
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for program with shapes")
	s.NotNil(program, "Expected program for program with shapes")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	namespaceStmt, ok := program.Statements[0].(*ast.NamespaceStatement)
	s.True(ok, "Expected first statement to be namespace")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseProgramWithExports tests parsing programs with exports
func (s *ProgramTestSuite) TestParseProgramWithExports() {
	input := `
namespace com/example
policy user {
	rule check = { yield true }
	export decision of check
}
export shape User
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for program with exports")
	s.NotNil(program, "Expected program for program with exports")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	namespaceStmt, ok := program.Statements[0].(*ast.NamespaceStatement)
	s.True(ok, "Expected first statement to be namespace")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseProgramWithComments tests parsing programs with comments
func (s *ProgramTestSuite) TestParseProgramWithComments() {
	input := `
-- This is a comment
namespace com/example
-- Another comment
policy user {
	rule check = { yield true } -- Inline comment
}
-- Final comment
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for program with comments")
	s.NotNil(program, "Expected program for program with comments")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	var namespaceStmt *ast.NamespaceStatement
	for _, stmt := range program.Statements {
		if ns, ok := stmt.(*ast.NamespaceStatement); ok {
			namespaceStmt = ns
			break
		}
	}
	s.NotNil(namespaceStmt, "Expected namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseProgramComplex tests parsing complex programs
func (s *ProgramTestSuite) TestParseProgramComplex() {
	input := `
-- Complex program example
namespace com/example

-- User shape
shape User { name:string age:int }

-- User policy
policy user {
	rule check = { yield true }
	fact name:string default "john"
	export decision of check
}
	
-- Admin policy
policy admin {
	rule validate = { yield x > 0 }
	fact role:string default "admin"
	export decision of validate
}

-- Export shape
export shape User
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for complex program")
	s.NotNil(program, "Expected program for complex program")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	var namespaceStmt *ast.NamespaceStatement
	for _, stmt := range program.Statements {
		if ns, ok := stmt.(*ast.NamespaceStatement); ok {
			namespaceStmt = ns
			break
		}
	}
	s.NotNil(namespaceStmt, "Expected namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseProgramEmpty tests parsing empty programs
func (s *ProgramTestSuite) TestParseProgramEmpty() {
	parser := NewParserFromString("", "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for empty program")
	s.Nil(program, "Expected nil program for empty input")
}

// TestParseProgramWhitespaceOnly tests parsing whitespace-only programs
func (s *ProgramTestSuite) TestParseProgramWhitespaceOnly() {
	parser := NewParserFromString("   \n\t   ", "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for whitespace-only program")
	s.Nil(program, "Expected nil program for whitespace-only input")
}

// TestParseProgramCommentOnly tests parsing comment-only programs
func (s *ProgramTestSuite) TestParseProgramCommentOnly() {
	parser := NewParserFromString("-- This is a comment", "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for comment-only program")
	s.NotNil(program, "Expected program for comment-only input")
}

// TestParseProgramInvalidNamespace tests parsing programs with invalid namespace
func (s *ProgramTestSuite) TestParseProgramInvalidNamespace() {
	testCases := []string{
		"policy user { }",     // Missing namespace
		"shape User { }",      // Missing namespace
		"rule check { true }", // Missing namespace
		"namespace",           // Incomplete namespace
		"namespace 123",       // Invalid namespace identifier
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		program, err := parser.ParseProgram(s.T().Context())
		s.Error(err, "Expected error for: %s", tc)
		s.Nil(program, "Expected nil program for: %s", tc)
	}
}

// TestParseProgramMultipleNamespaces tests parsing programs with multiple namespaces
func (s *ProgramTestSuite) TestParseProgramMultipleNamespaces() {
	input := `
namespace com/example
policy user { }
namespace com/other
`
	parser := NewParserFromString(input, "test.sentra")
	_, err := parser.ParseProgram(s.T().Context())
	s.Error(err, "Expected error for multiple namespaces")
}

// TestParseProgramEdgeCases tests parsing edge cases
func (s *ProgramTestSuite) TestParseProgramEdgeCases() {
	testCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		{"namespace com/example policy user { }", false, "Valid program with policy"},
		{"namespace com/example shape User { }", false, "Valid program with shape"},
		{"namespace com/example rule check { true }", true, "Rule at top level"},
		{"namespace com/example fact name:string", true, "Fact at top level"},
		{"namespace com/example use fn from @lib as alias", true, "Use at top level"},
		{"namespace com/example export rule check", true, "Export at top level"},
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

// TestProgramTestSuite runs the program test suite
func TestProgramTestSuite(t *testing.T) {
	suite.Run(t, new(ProgramTestSuite))
}
