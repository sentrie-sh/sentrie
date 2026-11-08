// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) parseRetention(input string) *ast.Program {
	program, err := NewParserFromString(input, "retention.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
	s.Require().NotNil(program)
	return program
}

// comments returns the content and Trailing flag of each comment statement.
func comments(stmts []ast.Statement) (contents []string, trailing []bool) {
	for _, st := range stmts {
		if c, ok := st.(*ast.CommentStatement); ok {
			contents = append(contents, c.Content)
			trailing = append(trailing, c.Trailing)
		}
	}
	return contents, trailing
}

func (s *ParserTestSuite) TestTopLevelCommentsRecordTrailing() {
	program := s.parseRetention("-- head\nnamespace a -- after ns\n-- own line\nshape S string -- after shape\n")
	contents, trailing := comments(program.Statements)
	s.Equal([]string{"head", "after ns", "own line", "after shape"}, contents)
	s.Equal([]bool{false, true, false, true}, trailing)
}

func (s *ParserTestSuite) TestPolicyBodyKeepsComments() {
	program := s.parseRetention(`namespace a
policy p { -- open
  -- lead
  fact f: string -- after fact
  -- between
  export decision of r -- after export
  -- last
}`)
	policy := program.Statements[1].(*ast.PolicyStatement)
	contents, trailing := comments(policy.Statements)
	s.Equal([]string{"open", "lead", "after fact", "between", "after export", "last"}, contents)
	s.Equal([]bool{true, false, true, false, true, false}, trailing)
	s.Len(policy.Statements, 8)
}

func (s *ParserTestSuite) TestBlockBodyAcceptsTrailingComment() {
	program := s.parseRetention("namespace a\npolicy p {\n  rule r = { -- open\n    -- lead\n    let x = 1\n    yield x\n  }\n}")
	rule := program.Statements[1].(*ast.PolicyStatement).Statements[0].(*ast.RuleStatement)
	block := rule.Body.(*ast.BlockExpression)
	contents, trailing := comments(block.Statements)
	s.Equal([]string{"open", "lead"}, contents)
	s.Equal([]bool{true, false}, trailing)
}

func (s *ParserTestSuite) TestShapeBodyKeepsComments() {
	program := s.parseRetention("namespace a\nshape S { -- open\n  -- lead\n  a: string -- after a\n  b: number\n}")
	shape := program.Statements[1].(*ast.ShapeStatement)
	s.Len(shape.Complex.Fields, 2)
	s.Require().Len(shape.Complex.Comments, 3)
	got := []string{}
	for _, c := range shape.Complex.Comments {
		got = append(got, c.Content)
	}
	s.Equal([]string{"open", "lead", "after a"}, got)
	s.True(shape.Complex.Comments[0].Trailing)
	s.False(shape.Complex.Comments[1].Trailing)
	s.True(shape.Complex.Comments[2].Trailing)
}

func (s *ParserTestSuite) TestPipelineKeepsSourceForm() {
	program := s.parseRetention("namespace a\npolicy p {\n  let x = v |> f(#, 1) |> g()\n}")
	let := program.Statements[1].(*ast.PolicyStatement).Statements[0].(*ast.VarDeclaration)
	outer := let.Value.(*ast.CallExpression)
	s.Equal("g", outer.Callee.(*ast.Identifier).Value)
	s.Empty(outer.PipeArgs)
	s.Len(outer.Arguments, 1, "evaluation still sees the desugared arguments")

	inner, ok := outer.Piped.(*ast.CallExpression)
	s.Require().True(ok)
	s.Equal("v", inner.Piped.(*ast.Identifier).Value)
	s.Require().Len(inner.PipeArgs, 2)
	s.IsType(&ast.PipelineHoleExpression{}, inner.PipeArgs[0])
	s.Equal("v", inner.Arguments[0].(*ast.Identifier).Value, "the hole is still substituted")
}

func (s *ParserTestSuite) TestPlainCallHasNoPipeSource() {
	program := s.parseRetention("namespace a\npolicy p {\n  let x = f(1)\n}")
	call := program.Statements[1].(*ast.PolicyStatement).Statements[0].(*ast.VarDeclaration).Value.(*ast.CallExpression)
	s.Nil(call.Piped)
	s.Nil(call.PipeArgs)
}

func (s *ParserTestSuite) TestBooleanSpellingIsKept() {
	program := s.parseRetention("namespace a\npolicy p {\n  let x: boolean = true\n  let y: trinary = true\n}")
	stmts := program.Statements[1].(*ast.PolicyStatement).Statements
	s.True(stmts[0].(*ast.VarDeclaration).Type.(*ast.TrinaryTypeRef).Boolean)
	s.False(stmts[1].(*ast.VarDeclaration).Type.(*ast.TrinaryTypeRef).Boolean)
}

func (s *ParserTestSuite) TestConsecutiveAndFileEndComments() {
	program := s.parseRetention("-- one\n-- two\n\n\n-- three\nnamespace a\npolicy p {\n  -- x\n  -- y\n}\n-- after policy\n-- end\n")
	contents, trailing := comments(program.Statements)
	s.Equal([]string{"one", "two", "three", "after policy", "end"}, contents)
	s.Equal([]bool{false, false, false, false, false}, trailing)
	policy := program.Statements[4].(*ast.PolicyStatement)
	inner, _ := comments(policy.Statements)
	s.Equal([]string{"x", "y"}, inner)
}

func (s *ParserTestSuite) TestCommentsDoNotChangeStatementCounts() {
	plain := s.parseRetention("namespace a\npolicy p {\n  fact f: string\n  let x = 1\n}")
	noisy := s.parseRetention("-- h\nnamespace a -- n\npolicy p { -- o\n  -- a\n  fact f: string -- f\n  -- b\n  let x = 1 -- x\n  -- c\n}\n-- e")
	count := func(stmts []ast.Statement) int {
		n := 0
		for _, st := range stmts {
			if _, ok := st.(*ast.CommentStatement); !ok {
				n++
			}
		}
		return n
	}
	policyOf := func(prg *ast.Program) *ast.PolicyStatement {
		for _, st := range prg.Statements {
			if p, ok := st.(*ast.PolicyStatement); ok {
				return p
			}
		}
		return nil
	}
	s.Equal(count(plain.Statements), count(noisy.Statements))
	s.Require().NotNil(policyOf(noisy))
	s.Equal(count(policyOf(plain).Statements), count(policyOf(noisy).Statements))
}
