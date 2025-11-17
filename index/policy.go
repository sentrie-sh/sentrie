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
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
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
	RuleExports map[string]*ExportedRule
	Uses        map[string]*ast.UseStatement // alias -> use statement
	Shapes      map[string]*Shape            // policy-local shapes

	seenIdentifiers map[string]ast.Positionable
}

func (p *Policy) String() string {
	return p.FQN.String()
}

func createPolicy(ns *Namespace, policy *ast.PolicyStatement, program *ast.Program) (*Policy, error) {
	p := &Policy{
		Statement:       policy,
		Namespace:       ns,
		Name:            policy.Name,
		FQN:             ast.CreateFQN(ns.FQN, policy.Name),
		FilePath:        program.Reference,
		Statements:      policy.Statements,
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Rules:           make(map[string]*Rule),
		RuleExports:     make(map[string]*ExportedRule),
		Uses:            make(map[string]*ast.UseStatement),
		Shapes:          make(map[string]*Shape),
		seenIdentifiers: make(map[string]ast.Positionable), // a map of seen identifiers
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
			// nothing should precede a use statement except comments and facts
			if idx > 0 {
				_, isPrecedingComment := policy.Statements[idx-1].(*ast.CommentStatement)
				_, isPrecedingFact := policy.Statements[idx-1].(*ast.FactStatement)
				_, isPrecedingUse := policy.Statements[idx-1].(*ast.UseStatement)

				if !isPrecedingComment && !isPrecedingFact && !isPrecedingUse {
					return nil, errors.Wrapf(ErrIndex, "'use' statement must be declared immediately after facts have been declared in a policy at %s", stmt.Span())
				}
			}
			if _, ok := p.Uses[stmt.As]; ok {
				return nil, errors.Wrapf(ErrIndex, "cannot rebind to existing alias '%s' at %s", stmt.As, stmt.Span())
			}
			p.Uses[stmt.As] = stmt

		case *ast.VarDeclaration:
			if err := p.AddLet(stmt); err != nil {
				return nil, err
			}

		case *ast.FactStatement:
			// nothing should precede a fact statement except comments and other facts
			if idx > 0 {
				_, isPrecedingComment := policy.Statements[idx-1].(*ast.CommentStatement)
				_, isPrecedingFact := policy.Statements[idx-1].(*ast.FactStatement)

				if !isPrecedingComment && !isPrecedingFact {
					return nil, errors.Wrapf(ErrIndex, "fact statement must be the first statement in a policy at %s", stmt.Span())
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
				return nil, errors.Wrapf(ErrIndex, "cannot export unknown rule: '%s' at %s", stmt.Of, stmt.Span())
			}

			if _, ok := p.RuleExports[stmt.Of]; ok {
				return nil, xerr.ErrConflict("rule export", stmt.Span(), stmt.Span())
			}

			att := []*RuleExportAttachment{}
			for _, a := range stmt.Attachments {
				// check if this attachment is already added
				exists := slices.IndexFunc(att, func(t *RuleExportAttachment) bool {
					return t.Name == a.What
				})

				if exists != -1 {
					return nil, xerr.ErrConflict("rule export attachment", a.Span(), att[exists].Value.Span())
				}

				att = append(att, &RuleExportAttachment{Name: a.What, Value: a.As})
			}

			p.RuleExports[stmt.Of] = &ExportedRule{RuleName: stmt.Of, Attachments: att}
		default:
			// ignore other statements
			_ = stmt
		}
	}

	if len(p.RuleExports) == 0 {
		return nil, errors.Wrapf(ErrIndex, "policy '%s' at '%s' does not export any rules", policy.Name, policy.Span())
	}

	return p, nil
}

func (p *Policy) AddLet(let *ast.VarDeclaration) error {
	if seen, ok := p.seenIdentifiers[let.Name]; ok {
		return xerr.ErrConflict("let declaration", let.Span(), seen.Span())
	}

	p.Lets[let.Name] = let
	p.seenIdentifiers[let.Name] = let
	return nil
}

func (p *Policy) AddRule(rule *ast.RuleStatement) error {
	r, err := createRule(p, rule)
	if err != nil {
		return err
	}

	if seen, ok := p.seenIdentifiers[rule.RuleName]; ok {
		return xerr.ErrConflict("rule declaration", rule.Span(), seen.Span())
	}

	p.Rules[rule.RuleName] = r
	p.seenIdentifiers[rule.RuleName] = r

	return nil
}

func (p *Policy) AddShape(shape *ast.ShapeStatement) error {
	if seen, ok := p.Shapes[shape.Name]; ok {
		return xerr.ErrConflict("shape declaration", shape.Span(), seen.Span())
	}

	s, err := createShape(p.Namespace, p, shape)
	if err != nil {
		return errors.Wrapf(ErrIndex, "failed to create shape: %s at %s", shape.Name, shape.Span())
	}

	p.Shapes[shape.Name] = s
	return nil
}

func (p *Policy) AddFact(fact *ast.FactStatement) error {
	if seen, ok := p.seenIdentifiers[fact.Alias]; ok {
		return xerr.ErrConflict("fact declaration", fact.Span(), seen.Span())
	}

	// Required facts (not optional) cannot have default values
	if !fact.Optional && fact.Default != nil {
		return xerr.ErrInvalidInvocation(fmt.Sprintf("required fact '%s' at %s cannot have a default value", fact.Alias, fact.Span()))
	}

	p.Facts[fact.Alias] = fact
	p.seenIdentifiers[fact.Alias] = fact
	return nil
}
