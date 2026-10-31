// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// parsePolicyBody parses a one-policy program and returns the policy's statements.
func (s *ParserTestSuite) parsePolicyBody(body string) []ast.Statement {
	src := "namespace com/ex\npolicy p {\n" + body + "\n}\n"
	prog, err := NewParserFromString(src, "contribute.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
	for _, stmt := range prog.Statements {
		if p, ok := stmt.(*ast.PolicyStatement); ok {
			return p.Statements
		}
	}
	s.FailNow("no policy statement")
	return nil
}

func (s *ParserTestSuite) TestParseRuleDeclaredType() {
	stmts := s.parsePolicyBody(`rule deny: list[string] = []
rule plain = true`)
	s.Require().Len(stmts, 2)

	typed, ok := stmts[0].(*ast.RuleStatement)
	s.Require().True(ok)
	s.Equal("deny", typed.RuleName)
	s.Require().NotNil(typed.Type)
	s.Equal("list[string]", typed.Type.String())

	plain, ok := stmts[1].(*ast.RuleStatement)
	s.Require().True(ok)
	s.Nil(plain.Type)
}

func (s *ParserTestSuite) TestParseContributeStatement() {
	stmts := s.parsePolicyBody(`rule deny: list[string] = []
contribute deny when user.disabled { yield "user disabled" }
contribute deny { yield "always" }`)
	s.Require().Len(stmts, 3)

	guarded, ok := stmts[1].(*ast.ContributeStatement)
	s.Require().True(ok)
	s.Equal("deny", guarded.RuleName)
	s.Require().NotNil(guarded.When)
	s.IsType(&ast.BlockExpression{}, guarded.Body)
	s.Equal("contribute_statement", guarded.Kind())

	unguarded, ok := stmts[2].(*ast.ContributeStatement)
	s.Require().True(ok)
	s.Nil(unguarded.When)
	s.IsType(&ast.BlockExpression{}, unguarded.Body)
}

func (s *ParserTestSuite) TestParseContributeErrors() {
	for name, body := range map[string]string{
		"missing rule name": `contribute when x { yield 1 }`,
		"missing body":      `contribute deny when x`,
		"bad rule type":     `rule deny: = []`,
	} {
		s.Run(name, func() {
			src := "namespace com/ex\npolicy p {\n" + body + "\n}\n"
			_, err := NewParserFromString(src, "contribute.sentrie").ParseProgram(s.T().Context())
			s.Error(err)
		})
	}
}

func (s *ParserTestSuite) TestContributeIsPolicyScopeOnly() {
	_, err := NewParserFromString("namespace com/ex\ncontribute deny { yield 1 }\n", "contribute.sentrie").ParseProgram(s.T().Context())
	s.Error(err)
}
