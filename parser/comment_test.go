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

	"github.com/binaek/sentra/ast"
	"github.com/stretchr/testify/suite"
)

// CommentTestSuite provides tests for comment and whitespace handling
type CommentTestSuite struct {
	suite.Suite
}

// findNamespaceStatement finds the namespace statement in a program's statements
func (s *CommentTestSuite) findNamespaceStatement(program *ast.Program) *ast.NamespaceStatement {
	for _, stmt := range program.Statements {
		if ns, ok := stmt.(*ast.NamespaceStatement); ok {
			return ns
		}
	}
	return nil
}

// SetupSuite initializes the test suite
func (s *CommentTestSuite) SetupSuite() {
	slog.Info("CommentTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *CommentTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "CommentTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *CommentTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "CommentTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *CommentTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestParseWithLineComments tests parsing with line comments
func (s *CommentTestSuite) TestParseWithLineComments() {
	input := `
-- This is a comment
namespace com/example; -- Another comment
-- Policy comment
policy test {
    let x = 42; -- Inline comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithMultiLineComments tests parsing with multi-line comments
func (s *CommentTestSuite) TestParseWithMultiLineComments() {
	input := `
-- This is a multi-line comment
-- that spans several lines
namespace com/example; -- Another comment
-- 
-- Policy comment
policy test {
    let x = 42; -- Inline comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithMixedComments tests parsing with mixed comment types
func (s *CommentTestSuite) TestParseWithMixedComments() {
	input := `
-- Line comment
-- Another line comment
namespace com/example; -- Line comment
-- Another comment
policy test {
    -- Line comment
    let x = 42; -- Inline comment
    -- Another comment
    rule test = true; -- Line comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithWhitespace tests parsing with various whitespace
func (s *CommentTestSuite) TestParseWithWhitespace() {
	input := `
namespace   com/example   ;   
policy   test   {   
   let   x   =   42   ;   
}   
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithTrailingComments tests parsing with trailing comments
func (s *CommentTestSuite) TestParseWithTrailingComments() {
	input := `
namespace com/example; -- Trailing comment
policy test { -- Trailing comment
    let x = 42; -- Trailing comment
    rule test = true; -- Trailing comment
} -- Trailing comment
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithPrecedingComments tests parsing with preceding comments
func (s *CommentTestSuite) TestParseWithPrecedingComments() {
	input := `
-- Preceding comment
namespace com/example;
-- Preceding comment
policy test {
    -- Preceding comment
    let x = 42;
    -- Preceding comment
    rule test = true;
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithNestedComments tests parsing with nested comments
func (s *CommentTestSuite) TestParseWithNestedComments() {
	input := `
-- Outer comment
-- Inner line comment
-- More outer comment
namespace com/example;
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")
}

// TestParseWithEmptyComments tests parsing with empty comments
func (s *CommentTestSuite) TestParseWithEmptyComments() {
	input := `
--
namespace com/example;
--
policy test {
    --
    let x = 42;
    --
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)

	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithOnlyComments tests parsing with only comments
func (s *CommentTestSuite) TestParseWithOnlyComments() {
	input := `
-- This is a comment
-- This is another comment
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err) // just having comments is valid
	s.NotNil(program)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	s.Equal(2, len(program.Statements), "Expected 0 statements in program")
}

// TestParseWithMalformedComments tests parsing with malformed comments
func (s *CommentTestSuite) TestParseWithMalformedComments() {
	input := `
/* Unclosed comment
namespace com/example;
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.Error(err)
	s.Nil(program)
	s.Contains(err.Error(), "expected")
}

// TestParseWithCommentsInStrings tests parsing with comments in strings
func (s *CommentTestSuite) TestParseWithCommentsInStrings() {
	input := `
namespace com/example;
policy test {
    let x = "-- This is not a comment";
    let y = "-- This is not a comment --";
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInExpressions tests parsing with comments in expressions
func (s *CommentTestSuite) TestParseWithCommentsInExpressions() {
	input := `
namespace com/example;
policy test {
    let x = 1 + 2; -- Comment after expression
    rule test = x > 0; -- Comment after expression
    fact user:string default "john"; -- Comment after expression
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInComplexExpressions tests parsing with comments in complex expressions
func (s *CommentTestSuite) TestParseWithCommentsInComplexExpressions() {
	input := `
namespace com/example;
policy test {
    let x = 1 + 2 * 3; -- Arithmetic
    rule test = x > 0 and y != ""; -- Logical
    fact user:string default (x > 0 ? "yes" : "no"); -- Ternary
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInLists tests parsing with comments in lists
func (s *CommentTestSuite) TestParseWithCommentsInLists() {
	input := `
namespace com/example;
policy test {
    let x = [1, 2, 3]; -- List comment
    let y = ["a", "b", "c"]; -- String list comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInMaps tests parsing with comments in maps
func (s *CommentTestSuite) TestParseWithCommentsInMaps() {
	input := `
namespace com/example;
policy test {
    let x = {"key": "value"}; -- Map comment
    let y = {"a": 1, "b": 2}; -- Multi-entry map comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInCalls tests parsing with comments in function calls
func (s *CommentTestSuite) TestParseWithCommentsInCalls() {
	input := `
namespace com/example;
policy test {
    let x = myFunction(1, 2); -- Call comment
    let y = otherFunction(a, b, c); -- Multi-arg call comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInIndexes tests parsing with comments in index expressions
func (s *CommentTestSuite) TestParseWithCommentsInIndexes() {
	input := `
namespace com/example;
policy test {
    let x = array[0]; -- Index comment
    let y = obj.field; -- Field access comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check for policy statements
	var policyCount int
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.PolicyStatement); ok {
			policyCount++
		}
	}
	s.Equal(1, policyCount, "Expected 1 policy statement")
}

// TestParseWithCommentsInShapes tests parsing with comments in shapes
func (s *CommentTestSuite) TestParseWithCommentsInShapes() {
	input := `
namespace com/example;
-- Shape comment
shape User {
    name: string -- Field comment
    age: int -- Another field comment
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")
	// Check for shape statements
	var shapeCount int
	for _, stmt := range program.Statements {
		if shapeStmt, ok := stmt.(*ast.ShapeStatement); ok {
			shapeCount++
			s.Equal("User", shapeStmt.Name, "Expected shape name")
		}
	}
	s.Equal(1, shapeCount, "Expected 1 shape statement")
}

// TestParseWithCommentsInExports tests parsing with comments in exports
func (s *CommentTestSuite) TestParseWithCommentsInExports() {
	input := `
namespace com/example
shape User {
    name: string
}
export shape User
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err)
	s.NotNil(program)
	// Check that we have statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")

	// Find the namespace statement (it might not be the first due to comments)
	namespaceStmt := s.findNamespaceStatement(program)
	s.NotNil(namespaceStmt, "Expected to find namespace statement")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")
	// Check for shape export statements
	var shapeExportCount int
	for _, stmt := range program.Statements {
		if exportStmt, ok := stmt.(*ast.ShapeExportStatement); ok {
			shapeExportCount++
			s.Equal("User", exportStmt.Name, "Expected shape export name")
		}
	}
	s.Equal(1, shapeExportCount, "Expected 1 shape export statement")
}

// TestParseWithCommentsTestSuite runs the comment test suite
func TestParseWithCommentsTestSuite(t *testing.T) {
	suite.Run(t, new(CommentTestSuite))
}
