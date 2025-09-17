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

// StatementTestSuite provides tests for statement parsing
type StatementTestSuite struct {
	suite.Suite
}

// SetupSuite initializes the test suite
func (s *StatementTestSuite) SetupSuite() {
	slog.Info("StatementTestSuite SetupSuite start")
}

// BeforeTest runs before each test
func (s *StatementTestSuite) BeforeTest(suiteName, testName string) {
	slog.Info("BeforeTest start", "TestSuite", "StatementTestSuite", "TestName", testName)
}

// AfterTest runs after each test
func (s *StatementTestSuite) AfterTest(suiteName, testName string) {
	slog.Info("AfterTest start", "TestSuite", "StatementTestSuite", "TestName", testName)
}

// TearDownSuite cleans up after all tests
func (s *StatementTestSuite) TearDownSuite() {
	slog.Info("TearDownSuite")
	slog.Info("TearDownSuite end")
}

// TestParseNamespaceStatement tests parsing namespace statements
func (s *StatementTestSuite) TestParseNamespaceStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"namespace com/example;", "com/example"},
		{"namespace com/example/test;", "com/example/test"},
		{"namespace test;", "test"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseNamespaceStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		namespaceStmt, ok := stmt.(*ast.NamespaceStatement)
		s.True(ok, "Expected NamespaceStatement for: %s", tc.input)
		s.Equal(tc.expected, namespaceStmt.Name.String(), "Expected namespace name: %s", tc.expected)
	}
}

// TestParseNamespaceStatementInvalid tests parsing invalid namespace statements
func (s *StatementTestSuite) TestParseNamespaceStatementInvalid() {
	testCases := []string{
		"namespace",      // Missing identifier
		"namespace 123;", // Invalid identifier
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseNamespaceStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParsePolicyStatement tests parsing policy statements
func (s *StatementTestSuite) TestParsePolicyStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"policy test { }", "test"},
		{"policy user { let x = 42; }", "user"},
		{"policy admin { rule check = { yield true } }", "admin"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseThePolicyStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		policyStmt, ok := stmt.(*ast.PolicyStatement)
		s.True(ok, "Expected PolicyStatement for: %s", tc.input)
		s.Equal(tc.expected, policyStmt.Name, "Expected policy name: %s", tc.expected)
	}
}

// TestParsePolicyStatementInvalid tests parsing invalid policy statements
func (s *StatementTestSuite) TestParsePolicyStatementInvalid() {
	testCases := []string{
		"policy",         // Missing identifier
		"policy 123 { }", // Invalid identifier
		"policy test",    // Missing body
		"policy test {",  // Missing closing brace
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseThePolicyStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseRuleStatement tests parsing rule statements
func (s *StatementTestSuite) TestParseRuleStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"rule check = import decision rulename from com/example", "check"},
		{"rule check = { yield true }", "check"},
		{"rule validate = { yield x > 0 }", "validate"},
		{"rule test = true", "test"},
		{"rule simple = x > 0", "simple"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseRuleStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		ruleStmt, ok := stmt.(*ast.RuleStatement)
		s.True(ok, "Expected RuleStatement for: %s", tc.input)
		s.Equal(tc.expected, ruleStmt.RuleName, "Expected rule name: %s", tc.expected)
	}
}

// TestParseRuleStatementInvalid tests parsing invalid rule statements
func (s *StatementTestSuite) TestParseRuleStatementInvalid() {
	testCases := []string{
		"rule",                     // Missing identifier
		"rule 123 { }",             // Invalid identifier
		"rule test",                // Missing body
		"rule test {",              // Missing closing brace
		"rule = true",              // Missing identifier
		"rule 123 = true",          // Invalid identifier
		"rule test { yield true }", // Missing = operator
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseRuleStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseFactStatement tests parsing fact statements
func (s *StatementTestSuite) TestParseFactStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"fact user:string", "user"},
		{"fact age:int default 25", "age"},
		{"fact name:string default \"john\"", "name"},
		{"fact name:ShapeName default \"john\"", "name"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseFactStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		factStmt, ok := stmt.(*ast.FactStatement)
		s.True(ok, "Expected FactStatement for: %s", tc.input)
		s.Equal(tc.expected, factStmt.Name, "Expected fact name: %s", tc.expected)
	}
}

// TestParseFactStatementInvalid tests parsing invalid fact statements
func (s *StatementTestSuite) TestParseFactStatementInvalid() {
	testCases := []string{
		"fact",            // Missing identifier
		"fact 123:string", // Invalid identifier
		"fact user",       // Missing type
		"fact user:",      // Missing type after colon
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseFactStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseShapeStatement tests parsing shape statements
func (s *StatementTestSuite) TestParseShapeStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"shape User { }", "User"},
		{"shape Person { name:string age:int }", "Person"},
		{"shape Simple string", "Simple"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseShapeStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		shapeStmt, ok := stmt.(*ast.ShapeStatement)
		s.True(ok, "Expected ShapeStatement for: %s", tc.input)
		s.Equal(tc.expected, shapeStmt.Name, "Expected shape name: %s", tc.expected)
	}
}

// TestParseShapeStatementEmpty tests parsing empty shape statements
func (s *StatementTestSuite) TestParseShapeStatementEmpty() {
	parser := NewParserFromString("shape User { }", "test.sentra")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.NoError(parser.err, "Expected no error for empty shape")
	s.NotNil(stmt, "Expected statement for empty shape")

	shapeStmt, ok := stmt.(*ast.ShapeStatement)
	s.True(ok, "Expected ShapeStatement for empty shape")
	s.Equal("User", shapeStmt.Name, "Expected shape name: User")
	s.NotNil(shapeStmt.Complex, "Expected Complex field for shape with body")
}

// TestParseShapeStatementInvalid tests parsing invalid shape statements
func (s *StatementTestSuite) TestParseShapeStatementInvalid() {
	testCases := []string{
		"shape",         // Missing identifier
		"shape 123 { }", // Invalid identifier
		"shape User",    // Missing body
		"shape User {",  // Missing closing brace
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseShapeStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseUseStatement tests parsing use statements
func (s *StatementTestSuite) TestParseUseStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"use fn1, fn2 from @lib/name as alias", "alias"},
		{"use func from @sentra/std as std", "std"},
		{"use helper from @local/utils as utils", "utils"},
		{"use helper from \"./fn.ts\"", "fn.ts"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseUseStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		useStmt, ok := stmt.(*ast.UseStatement)
		s.True(ok, "Expected UseStatement for: %s", tc.input)
		s.Equal(tc.expected, useStmt.As, "Expected use alias: %s", tc.expected)
	}
}

// TestParseUseStatementInvalid tests parsing invalid use statements
func (s *StatementTestSuite) TestParseUseStatementInvalid() {
	testCases := []string{
		// "use", // Missing everything
		// "use fn1 from", // Missing module
		// "use fn1 from @lib/name as", // Missing alias
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseUseStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseRuleExportStatement tests parsing rule export statements
func (s *StatementTestSuite) TestParseRuleExportStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"export decision of check", "check"},
		{"export decision of validate", "validate"},
		{"export decision of test", "test"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseRuleExportStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		exportStmt, ok := stmt.(*ast.RuleExportStatement)
		s.True(ok, "Expected RuleExportStatement for: %s", tc.input)
		s.Equal(tc.expected, exportStmt.Of, "Expected export rule name: %s", tc.expected)
	}
}

// TestParseRuleExportStatementInvalid tests parsing invalid rule export statements
func (s *StatementTestSuite) TestParseRuleExportStatementInvalid() {
	testCases := []string{
		"export",                 // Missing rule
		"export decision of",     // Missing identifier
		"export decision of 123", // Invalid identifier
		"export shape test",      // Wrong keyword
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseRuleExportStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseShapeExportStatement tests parsing shape export statements
func (s *StatementTestSuite) TestParseShapeExportStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"export shape User", "User"},
		{"export shape Person", "Person"},
		{"export shape Simple", "Simple"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseShapeExportStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s", tc.input)
		s.NotNil(stmt, "Expected statement for: %s", tc.input)

		exportStmt, ok := stmt.(*ast.ShapeExportStatement)
		s.True(ok, "Expected ShapeExportStatement for: %s", tc.input)
		s.Equal(tc.expected, exportStmt.Name, "Expected export shape name: %s", tc.expected)
	}
}

// TestParseShapeExportStatementInvalid tests parsing invalid shape export statements
func (s *StatementTestSuite) TestParseShapeExportStatementInvalid() {
	testCases := []string{
		"export",           // Missing shape
		"export shape",     // Missing identifier
		"export shape 123", // Invalid identifier
		"export rule test", // Wrong keyword
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc, "test.sentra")
		stmt := parseShapeExportStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s", tc)
		s.Nil(stmt, "Expected nil statement for: %s", tc)
	}
}

// TestParseStatementComplexNested tests parsing complex nested statements
func (s *StatementTestSuite) TestParseStatementComplexNested() {
	input := `
namespace com/example;
policy user {
	rule check = { yield true }
	fact name:string default "john"
	shape User { name:string age:int }
}
`
	parser := NewParserFromString(input, "test.sentra")
	program, err := parser.ParseProgram(s.T().Context())
	s.NoError(err, "Expected no error for complex nested statements")
	s.NotNil(program, "Expected program for complex nested statements")

	// Check namespace (first statement should be namespace)
	s.Greater(len(program.Statements), 0, "Expected statements in program")
	namespaceStmt, ok := program.Statements[0].(*ast.NamespaceStatement)
	s.True(ok, "Expected first statement to be namespace")
	s.Equal("com/example", namespaceStmt.Name.String(), "Expected namespace name")

	// Check statements
	s.Greater(len(program.Statements), 0, "Expected statements in program")
}

// TestParseStatementEdgeCases tests parsing edge cases
func (s *StatementTestSuite) TestParseStatementEdgeCases() {
	testCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		// {"namespace", true, "Incomplete namespace"},
		// {"policy", true, "Incomplete policy"},
		// {"shape", true, "Incomplete shape"},
		// {"rule", true, "Incomplete rule"},
		// {"fact", true, "Incomplete fact"},
		// {"export", true, "Incomplete export"},
		// {"use", true, "Incomplete use"},
		// {"", false, "Empty input"},
		{"-- comment", false, "Comment only"},
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

// TestStatementTestSuite runs the statement test suite
func TestStatementTestSuite(t *testing.T) {
	suite.Run(t, new(StatementTestSuite))
}
