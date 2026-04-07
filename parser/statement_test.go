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
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/stretchr/testify/require"
)

// TestParseNamespaceStatement tests parsing namespace statements
func (s *ParserTestSuite) TestParseNamespaceStatement() {
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
func (s *ParserTestSuite) TestParseNamespaceStatementInvalid() {
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
func (s *ParserTestSuite) TestParsePolicyStatement() {
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
func (s *ParserTestSuite) TestParsePolicyStatementInvalid() {
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
func (s *ParserTestSuite) TestParseRuleStatement() {
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
func (s *ParserTestSuite) TestParseRuleStatementInvalid() {
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
func (s *ParserTestSuite) TestParseFactStatement() {
	testCases := []struct {
		input       string
		expected    string
		optional    bool
		description string
	}{
		{"fact user:string", "user", false, "required fact"},
		{"fact user?:string", "user", true, "optional fact"},
		{"fact age:int default 25", "age", false, "required fact with default"},
		{"fact age?:int default 25", "age", true, "optional fact with default"},
		{"fact name:string default \"john\"", "name", false, "required fact with string default"},
		{"fact name?:string default \"john\"", "name", true, "optional fact with string default"},
		{"fact name:ShapeName default \"john\"", "name", false, "required fact with shape type"},
		{"fact name?:ShapeName default \"john\"", "name", true, "optional fact with shape type"},
		{"fact userId:string as id", "userId", false, "required fact with alias"},
		{"fact userId?:string as id", "userId", true, "optional fact with alias"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseFactStatement(s.T().Context(), parser)
		s.NoError(parser.err, "Expected no error for: %s (%s)", tc.input, tc.description)
		s.NotNil(stmt, "Expected statement for: %s (%s)", tc.input, tc.description)

		factStmt, ok := stmt.(*ast.FactStatement)
		s.True(ok, "Expected FactStatement for: %s (%s)", tc.input, tc.description)
		s.Equal(tc.expected, factStmt.Name, "Expected fact name: %s (%s)", tc.expected, tc.description)
		s.Equal(tc.optional, factStmt.Optional, "Expected optional=%v for: %s (%s)", tc.optional, tc.input, tc.description)
	}
}

// TestParseFactStatementInvalid tests parsing invalid fact statements
func (s *ParserTestSuite) TestParseFactStatementInvalid() {
	testCases := []struct {
		input       string
		description string
	}{
		{"fact", "missing identifier"},
		{"fact 123:string", "invalid identifier"},
		{"fact user", "missing type"},
		{"fact user:", "missing type after colon"},
		{"fact user!:string", "! operator not allowed (facts are always non-nullable)"},
		{"fact user!?:string", "! operator not allowed"},
		{"fact user?!:string", "! operator not allowed"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		stmt := parseFactStatement(s.T().Context(), parser)
		s.Error(parser.err, "Expected error for: %s (%s)", tc.input, tc.description)
		s.Nil(stmt, "Expected nil statement for: %s (%s)", tc.input, tc.description)
	}
}

// TestParseShapeStatement tests parsing shape statements
func (s *ParserTestSuite) TestParseShapeStatement() {
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
func (s *ParserTestSuite) TestParseShapeStatementEmpty() {
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
func (s *ParserTestSuite) TestParseShapeStatementInvalid() {
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
func (s *ParserTestSuite) TestParseUseStatement() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"use {fn1, fn2} from @lib/name as alias", "alias"},
		{"use {func} from @sentra/std as std", "std"},
		{"use {helper} from @local/utils as utils", "utils"},
		{"use {helper} from \"./fn.ts\"", "fn.ts"},
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
func (s *ParserTestSuite) TestParseUseStatementInvalid() {
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
func (s *ParserTestSuite) TestParseRuleExportStatement() {
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
func (s *ParserTestSuite) TestParseRuleExportStatementInvalid() {
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
func (s *ParserTestSuite) TestParseShapeExportStatement() {
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
func (s *ParserTestSuite) TestParseShapeExportStatementInvalid() {
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
func (s *ParserTestSuite) TestParseStatementComplexNested() {
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
func (s *ParserTestSuite) TestParseStatementEdgeCases() {
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

func (s *ParserTestSuite) TestParsePolicyMetadataKeywordOutsidePolicyBody() {
	src := `namespace com/example
title "x"`
	parser := NewParserFromString(src, "test.sentra")
	_, err := parser.ParseProgram(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "only allowed inside a policy")
}

func (s *ParserTestSuite) TestParsePolicyWithMetadataStatements() {
	src := `namespace com/example

policy p {
  title "Hi"
  description ""
  version "2.0.0"
  tag "k" = "v"
  tag "k" = ""
  fact user: string
  use { x } from @sentrie/std as std
  rule allow = default true { yield true }
  export decision of allow
}`
	parser := NewParserFromString(src, "test.sentra")
	prg, err := parser.ParseProgram(context.Background())
	s.NoError(err)
	s.NotNil(prg)
	var pol *ast.PolicyStatement
	for _, st := range prg.Statements {
		if p, ok := st.(*ast.PolicyStatement); ok {
			pol = p
			break
		}
	}
	require.NotNil(s.T(), pol)
	var titles, descs, vers, tags, facts int
	for _, st := range pol.Statements {
		switch st.(type) {
		case *ast.TitleStatement:
			titles++
		case *ast.DescriptionStatement:
			descs++
		case *ast.VersionStatement:
			vers++
		case *ast.TagStatement:
			tags++
		case *ast.FactStatement:
			facts++
		}
	}
	s.Equal(1, titles)
	s.Equal(1, descs)
	s.Equal(1, vers)
	s.Equal(2, tags)
	s.Equal(1, facts)
}

func (s *ParserTestSuite) TestParseTagStatementInvalid() {
	parser := NewParserFromString(`policy p { tag "a" "b" }`, "test.sentra")
	_, err := parser.ParseProgram(context.Background())
	s.Error(err)
}
