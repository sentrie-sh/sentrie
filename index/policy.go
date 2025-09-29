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

package index

import (
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

type RuleExportAttachment struct {
	Name  string
	Value ast.Expression
}

// ExportedRule captures an exported rule's name and its attachment names.
type ExportedRule struct {
	RuleName    string
	Attachments []*RuleExportAttachment // names only; values computed at runtime
}

// Policy holds the AST statements and exports.
type Policy struct {
	Statement  *ast.PolicyStatement
	Namespace  *Namespace
	Name       string
	FQN        ast.FQN
	FilePath   string
	Statements []ast.Statement

	Lets        map[string]*ast.VarDeclaration
	Facts       map[string]*ast.FactStatement
	Rules       map[string]*Rule
	RuleExports map[string]ExportedRule
	Uses        []*ast.UseStatement
	Shapes      map[string]*Shape // policy-local shapes

	knownIdentifiers map[string]positionable
}

func (p *Policy) String() string {
	return p.FQN.String()
}

type positionable interface {
	Position() tokens.Position
}

func createPolicy(ns *Namespace, policy *ast.PolicyStatement, program *ast.Program) (*Policy, error) {
	p := &Policy{
		Statement:        policy,
		Namespace:        ns,
		Name:             policy.Name,
		FQN:              ast.CreateFQN(ns.FQN, policy.Name),
		FilePath:         program.Reference,
		Statements:       policy.Statements,
		Lets:             make(map[string]*ast.VarDeclaration),
		Facts:            make(map[string]*ast.FactStatement),
		Rules:            make(map[string]*Rule),
		RuleExports:      make(map[string]ExportedRule),
		Uses:             make([]*ast.UseStatement, 0),
		Shapes:           make(map[string]*Shape),
		knownIdentifiers: make(map[string]positionable),
	}

	for idx, stmt := range policy.Statements {
		if _, ok := stmt.(*ast.CommentStatement); ok {
			continue
		}

		switch stmt := stmt.(type) {
		case *ast.ShapeStatement:
			if err := p.AddShape(stmt); err != nil {
				return nil, err
			}
		case *ast.UseStatement:
			// nothing should precede a use statement expect comments and facts
			if idx > 0 {
				if _, ok := policy.Statements[idx-1].(*ast.CommentStatement); !ok {
					if _, ok := policy.Statements[idx-1].(*ast.FactStatement); !ok {
						return nil, errors.Wrapf(ErrIndex, "'use' statement must be immediately after facts have been declared in a policy at %s", stmt.Position())
					}
				}
			}
			p.Uses = append(p.Uses, stmt)

		case *ast.VarDeclaration:
			if err := p.AddLet(stmt); err != nil {
				return nil, err
			}

		case *ast.FactStatement:
			// nothing should precede a fact statement expect comments
			if idx > 0 {
				if _, ok := policy.Statements[idx-1].(*ast.CommentStatement); !ok {
					return nil, errors.Wrapf(ErrIndex, "fact statement must be the first statement in a policy at %s", stmt.Position())
				}
			}
			if err := p.AddFact(stmt); err != nil {
				return nil, err
			}

		case *ast.RuleStatement:
			if err := p.AddRule(stmt); err != nil {
				return nil, err
			}

		case *ast.RuleExportStatement:
			// get the rule
			if _, ok := p.Rules[stmt.Of]; !ok {
				return nil, errors.Wrapf(ErrIndex, "cannot export unknown rule: '%s' at %s", stmt.Of, stmt.Position())
			}

			if _, ok := p.RuleExports[stmt.Of]; ok {
				return nil, errors.Wrapf(ErrIndex, "rule export conflict: '%s' at %s", stmt.Of, stmt.Position())
			}

			att := []*RuleExportAttachment{}
			for _, a := range stmt.Attachments {
				if _, ok := p.RuleExports[a.What]; ok {
					return nil, errors.Wrapf(ErrIndex, "rule export attachment conflict: '%s' at %s", a.What, a.Pos)
				}

				att = append(att, &RuleExportAttachment{Name: a.What, Value: a.As})
			}
			p.RuleExports[stmt.Of] = ExportedRule{RuleName: stmt.Of, Attachments: att}
		default:
			// ignore other statements
			_ = stmt
		}
	}

	if len(p.RuleExports) == 0 {
		return nil, errors.Wrapf(ErrIndex, "Policy '%s' at '%s' does not export any rules", policy.Name, policy.Position())
	}

	return p, nil
}

func (p *Policy) AddLet(let *ast.VarDeclaration) error {
	if _, ok := p.knownIdentifiers[let.Name]; ok {
		return errors.Wrapf(ErrIndex, "let name conflict: '%s' at %s with %s", let.Name, let.Position(), p.knownIdentifiers[let.Name].Position())
	}

	p.Lets[let.Name] = let
	p.knownIdentifiers[let.Name] = let
	return nil
}

func (p *Policy) AddRule(rule *ast.RuleStatement) error {
	r, err := createRule(p, rule)
	if err != nil {
		return err
	}

	if _, ok := p.knownIdentifiers[rule.RuleName]; ok {
		return errors.Wrapf(ErrIndex, "rule name conflict: '%s' at %s with %s", rule.RuleName, rule.Position(), p.knownIdentifiers[rule.RuleName].Position())
	}

	p.Rules[rule.RuleName] = r
	p.knownIdentifiers[rule.RuleName] = r

	return nil
}

func (p *Policy) AddShape(shape *ast.ShapeStatement) error {
	if s, ok := p.Shapes[shape.Name]; ok {
		return errors.Wrapf(ErrIndex, "shape name conflict: '%s' at %s with %s", shape.Name, shape.Position(), s.Statement.Pos)
	}

	s, err := createShape(p.Namespace, p, shape)
	if err != nil {
		return errors.Wrapf(ErrIndex, "failed to create shape: %s at %s", shape.Name, shape.Position())
	}

	p.Shapes[shape.Name] = s
	return nil
}

func (p *Policy) AddFact(fact *ast.FactStatement) error {
	if _, ok := p.knownIdentifiers[fact.Alias]; ok {
		return errors.Wrapf(ErrIndex, "fact alias conflict: '%s' at %s with %s", fact.Alias, fact.Position(), p.knownIdentifiers[fact.Alias].Position())
	}

	p.Facts[fact.Alias] = fact
	p.knownIdentifiers[fact.Alias] = fact
	return nil
}
