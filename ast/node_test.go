package ast

import (
	"github.com/binaek/sentra/tokens"
)

// TestNodeInterface tests the Node interface implementation
func (s *AstTestSuite) TestNodeInterface() {
	// Test that all node types implement the Node interface
	pos := tokens.Position{Line: 1, Column: 1}

	// Test Identifier implements Node
	ident := &Identifier{Pos: pos, Value: "test"}
	s.Implements((*Node)(nil), ident)
	s.Equal("test", ident.String())
	s.Equal(pos, ident.Position())

	// Test StringLiteral implements Node
	str := &StringLiteral{Pos: pos, Value: "hello"}
	s.Implements((*Node)(nil), str)
	s.Equal(`"hello"`, str.String())
	s.Equal(pos, str.Position())
}

// TestCodeableInterface tests the Codeable interface implementation
func (s *AstTestSuite) TestCodeableInterface() {
	// TODO: Implement tests for Codeable interface
	// This will test nodes that implement both Node and Codeable interfaces
}

// TestStatementInterface tests the Statement interface implementation
func (s *AstTestSuite) TestStatementInterface() {
	// Test that all statement types implement the Statement interface
	pos := tokens.Position{Line: 1, Column: 1}

	// Test PolicyStatement implements Statement
	policy := &PolicyStatement{Pos: pos, Name: "testPolicy"}
	s.Implements((*Statement)(nil), policy)
	s.Implements((*Node)(nil), policy)
	s.Equal("testPolicy", policy.String())
	s.Equal(pos, policy.Position())

	// Test RuleStatement implements Statement
	rule := &RuleStatement{Pos: pos, RuleName: "testRule"}
	s.Implements((*Statement)(nil), rule)
	s.Implements((*Node)(nil), rule)
	s.Equal("testRule", rule.String())
	s.Equal(pos, rule.Position())
}

// TestExpressionInterface tests the Expression interface implementation
func (s *AstTestSuite) TestExpressionInterface() {
	// Test that all expression types implement the Expression interface
	pos := tokens.Position{Line: 1, Column: 1}

	// Test Identifier implements Expression
	ident := &Identifier{Pos: pos, Value: "test"}
	s.Implements((*Expression)(nil), ident)
	s.Implements((*Node)(nil), ident)

	// Test StringLiteral implements Expression
	str := &StringLiteral{Pos: pos, Value: "hello"}
	s.Implements((*Expression)(nil), str)
	s.Implements((*Node)(nil), str)
}

// TestNodePositioning tests position handling across different node types
func (s *AstTestSuite) TestNodePositioning() {
	// Test various position values
	testPositions := []tokens.Position{
		{Line: 1, Column: 1},
		{Line: 10, Column: 5},
		{Line: 100, Column: 50},
		{Line: 0, Column: 0}, // Edge case
	}

	for _, pos := range testPositions {
		ident := &Identifier{Pos: pos, Value: "test"}
		s.Equal(pos, ident.Position())

		str := &StringLiteral{Pos: pos, Value: "test"}
		s.Equal(pos, str.Position())
	}
}

// TestNodeStringRepresentation tests string representation of nodes
func (s *AstTestSuite) TestNodeStringRepresentation() {
	pos := tokens.Position{Line: 1, Column: 1}

	// Test Identifier string representation
	ident := &Identifier{Pos: pos, Value: "myVariable"}
	s.Equal("myVariable", ident.String())

	// Test StringLiteral string representation
	str := &StringLiteral{Pos: pos, Value: "hello world"}
	s.Equal(`"hello world"`, str.String())

	// Test empty values
	emptyIdent := &Identifier{Pos: pos, Value: ""}
	s.Equal("", emptyIdent.String())

	emptyStr := &StringLiteral{Pos: pos, Value: ""}
	s.Equal(`""`, emptyStr.String())
}
