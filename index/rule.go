package index

import (
	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/tokens"
)

type Rule struct {
	Node    *ast.RuleStatement
	Policy  *Policy
	Name    string
	FQN     ast.FQN
	Default ast.Expression
	When    ast.Expression
	Body    ast.Expression
}

func (r *Rule) String() string {
	return r.FQN.String()
}

func (r *Rule) Position() tokens.Position {
	return r.Node.Position()
}

func createRule(p *Policy, stmt *ast.RuleStatement) (*Rule, error) {
	return &Rule{
		Node:    stmt,
		Policy:  p,
		Name:    stmt.RuleName,
		FQN:     ast.CreateFQN(p.FQN, stmt.RuleName),
		Default: stmt.Default,
		When:    stmt.When,
		Body:    stmt.Body,
	}, nil
}
