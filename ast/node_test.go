package ast

import (
	"github.com/sentrie-sh/sentrie/tokens"
)

// TestNodeInterface tests the Node interface implementation
func (s *AstTestSuite) TestNodeInterface() {
	// Test that all node types implement the Node interface
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 4, Offset: 3},
	}

	// Test Identifier implements Node
	ident := &Identifier{Range: r, Value: "test"}
	s.Implements((*Node)(nil), ident)
	s.Equal("test", ident.String())
	s.Equal(r, ident.Span())

	// Test StringLiteral implements Node
	str := &StringLiteral{Range: r, Value: "hello"}
	s.Implements((*Node)(nil), str)
	s.Equal(`"hello"`, str.String())
	s.Equal(r, str.Span())
}

// TestCodeableInterface tests the Codeable interface implementation
func (s *AstTestSuite) TestCodeableInterface() {
	// TODO: Implement tests for Codeable interface
	// This will test nodes that implement both Node and Codeable interfaces
}

// TestStatementInterface tests the Statement interface implementation
func (s *AstTestSuite) TestStatementInterface() {
	// Test that all statement types implement the Statement interface
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 4, Offset: 3},
	}

	// Test PolicyStatement implements Statement
	policy := &PolicyStatement{Range: r, Name: "testPolicy"}
	s.Implements((*Statement)(nil), policy)
	s.Implements((*Node)(nil), policy)
	s.Equal("testPolicy", policy.String())
	s.Equal(r, policy.Span())

	// Test RuleStatement implements Statement
	rule := &RuleStatement{Range: r, RuleName: "testRule"}
	s.Implements((*Statement)(nil), rule)
	s.Implements((*Node)(nil), rule)
	s.Equal("testRule", rule.String())
	s.Equal(r, rule.Span())
}

// TestExpressionInterface tests the Expression interface implementation
func (s *AstTestSuite) TestExpressionInterface() {
	// Test that all expression types implement the Expression interface
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 4, Offset: 3},
	}

	// Test Identifier implements Expression
	ident := &Identifier{Range: r, Value: "test"}
	s.Implements((*Expression)(nil), ident)
	s.Implements((*Node)(nil), ident)

	// Test StringLiteral implements Expression
	str := &StringLiteral{Range: r, Value: "hello"}
	s.Implements((*Expression)(nil), str)
	s.Implements((*Node)(nil), str)
}

// TestNodePositioning tests position handling across different node types
func (s *AstTestSuite) TestNodePositioning() {
	// Test various range values
	testRanges := []tokens.Range{
		{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 4, Offset: 3}},
		{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 5, Offset: 0}, To: tokens.Pos{Line: 10, Column: 8, Offset: 3}},
		{File: "test.sentra", From: tokens.Pos{Line: 100, Column: 50, Offset: 0}, To: tokens.Pos{Line: 100, Column: 53, Offset: 3}},
		{File: "test.sentra", From: tokens.Pos{Line: 0, Column: 0, Offset: 0}, To: tokens.Pos{Line: 0, Column: 3, Offset: 3}}, // Edge case
	}

	for _, r := range testRanges {
		ident := &Identifier{Range: r, Value: "test"}
		s.Equal(r, ident.Span())

		str := &StringLiteral{Range: r, Value: "test"}
		s.Equal(r, str.Span())
	}
}

// TestNodeStringRepresentation tests string representation of nodes
func (s *AstTestSuite) TestNodeStringRepresentation() {
	r := tokens.Range{
		File: "test.sentra",
		From: tokens.Pos{Line: 1, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: 1, Column: 4, Offset: 3},
	}

	// Test Identifier string representation
	ident := &Identifier{Range: r, Value: "myVariable"}
	s.Equal("myVariable", ident.String())

	// Test StringLiteral string representation
	str := &StringLiteral{Range: r, Value: "hello world"}
	s.Equal(`"hello world"`, str.String())

	// Test empty values
	emptyIdent := &Identifier{Range: r, Value: ""}
	s.Equal("", emptyIdent.String())

	emptyStr := &StringLiteral{Range: r, Value: ""}
	s.Equal(`""`, emptyStr.String())
}
