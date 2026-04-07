// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
	"strings"

	"github.com/Masterminds/semver/v3"
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

// PolicyTagPair is one key/value from policy `tag` statements (order preserved in Policy.TagPairs).
type PolicyTagPair struct {
	Key   string
	Value string
}

// Policy holds the AST statements and exports.
type Policy struct {
	Statement  *ast.PolicyStatement
	Namespace  *Namespace
	Name       string
	FQN        ast.FQN
	FilePath   string
	Statements []ast.Statement

	Title          *string
	Description    *string
	VersionLiteral string
	Version        *semver.Version
	TagPairs       []PolicyTagPair
	// TagsByKey is derived from TagPairs for query ergonomics; map iteration order is not stable.
	TagsByKey map[string][]string

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

// latePolicyHeaderErr reports metadata, fact, or use after the policy body has started.
func latePolicyHeaderErr(keyword, at string) error {
	return errors.Wrapf(xerr.ErrIndex, "'%s' must appear before rules, exports, lets, and shapes at %s", keyword, at)
}

type policyPhase int

const (
	policyPhaseMeta policyPhase = iota
	policyPhaseFacts
	policyPhaseUses
	policyPhaseBody
)

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
		seenIdentifiers: make(map[string]ast.Positionable),
	}

	phase := policyPhaseMeta
	var titleAt, descriptionAt, versionAt ast.Positionable

	for _, stmt := range policy.Statements {
		if policyStmtKindOf(stmt) == policyStmtComment {
			continue
		}

		switch stmt := stmt.(type) {
		case *ast.TitleStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("title", stmt.Span().String())
				}
				return nil, errors.Wrapf(xerr.ErrPolicyMetadataContiguous, "at %s", stmt.Span())
			}
			if titleAt != nil {
				return nil, xerr.ErrConflict("policy title", stmt.Span(), titleAt.Span())
			}
			trimmed := strings.TrimSpace(stmt.Value)
			if trimmed == "" {
				return nil, errors.Wrapf(xerr.ErrPolicyEmptyTitle, "at %s", stmt.Span())
			}
			t := trimmed
			p.Title = &t
			titleAt = stmt

		case *ast.DescriptionStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("description", stmt.Span().String())
				}
				return nil, errors.Wrapf(xerr.ErrPolicyMetadataContiguous, "at %s", stmt.Span())
			}
			if descriptionAt != nil {
				return nil, xerr.ErrConflict("policy description", stmt.Span(), descriptionAt.Span())
			}
			d := strings.TrimSpace(stmt.Value)
			p.Description = &d
			descriptionAt = stmt

		case *ast.VersionStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("version", stmt.Span().String())
				}
				return nil, errors.Wrapf(xerr.ErrPolicyMetadataContiguous, "at %s", stmt.Span())
			}
			if versionAt != nil {
				return nil, xerr.ErrConflict("policy version", stmt.Span(), versionAt.Span())
			}
			p.VersionLiteral = stmt.Literal
			// SemVer is validated on TrimSpace(literal); VersionLiteral stays verbatim for display/diagnostics.
			ver, err := semver.NewVersion(strings.TrimSpace(stmt.Literal))
			if err != nil {
				return nil, errors.Wrapf(xerr.ErrPolicyInvalidVersion, "at %s", stmt.Span())
			}
			p.Version = ver
			versionAt = stmt

		case *ast.TagStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("tag", stmt.Span().String())
				}
				return nil, errors.Wrapf(xerr.ErrPolicyMetadataContiguous, "at %s", stmt.Span())
			}
			key := strings.TrimSpace(stmt.Key)
			if key == "" {
				return nil, errors.Wrapf(xerr.ErrPolicyEmptyTagKey, "at %s", stmt.Span())
			}
			p.TagPairs = append(p.TagPairs, PolicyTagPair{Key: key, Value: stmt.Value})

		case *ast.FactStatement:
			switch phase {
			case policyPhaseMeta:
				phase = policyPhaseFacts
			case policyPhaseFacts:
			case policyPhaseUses:
				return nil, errors.Wrapf(xerr.ErrPolicyFactAfterUse, "at %s", stmt.Span())
			case policyPhaseBody:
				return nil, latePolicyHeaderErr("fact", stmt.Span().String())
			}
			if err := p.AddFact(stmt); err != nil {
				return nil, err
			}

		case *ast.UseStatement:
			switch phase {
			case policyPhaseMeta:
				phase = policyPhaseUses
			case policyPhaseFacts:
				phase = policyPhaseUses
			case policyPhaseUses:
			case policyPhaseBody:
				return nil, latePolicyHeaderErr("use", stmt.Span().String())
			}
			if _, ok := p.Uses[stmt.As]; ok {
				return nil, errors.Wrapf(xerr.ErrIndex, "cannot rebind to existing alias '%s' at %s", stmt.As, stmt.Span())
			}
			p.Uses[stmt.As] = stmt

		case *ast.VarDeclaration:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if err := p.AddLet(stmt); err != nil {
				return nil, err
			}

		case *ast.RuleStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if err := p.AddRule(stmt); err != nil {
				return nil, err
			}

		case *ast.RuleExportStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if _, ok := p.Rules[stmt.Of]; !ok {
				return nil, errors.Wrapf(xerr.ErrIndex, "cannot export unknown rule: '%s' at %s", stmt.Of, stmt.Span())
			}

			if _, ok := p.RuleExports[stmt.Of]; ok {
				return nil, xerr.ErrConflict("rule export", stmt.Span(), stmt.Span())
			}

			att := []*RuleExportAttachment{}
			for _, a := range stmt.Attachments {
				exists := slices.IndexFunc(att, func(t *RuleExportAttachment) bool {
					return t.Name == a.What
				})

				if exists != -1 {
					return nil, xerr.ErrConflict("rule export attachment", a.Span(), att[exists].Value.Span())
				}

				att = append(att, &RuleExportAttachment{Name: a.What, Value: a.As})
			}

			p.RuleExports[stmt.Of] = &ExportedRule{RuleName: stmt.Of, Attachments: att}

		case *ast.ShapeStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if err := p.AddShape(stmt); err != nil {
				return nil, err
			}

		default:
			return nil, errors.Wrapf(xerr.ErrIndex, "unsupported statement in policy at %s", stmt.Span())
		}
	}

	p.TagsByKey = buildTagsByKey(p.TagPairs)

	if len(p.RuleExports) == 0 {
		return nil, errors.Wrapf(xerr.ErrIndex, "policy '%s' at '%s' does not export any rules", policy.Name, policy.Span())
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
		return errors.Wrapf(xerr.ErrIndex, "failed to create shape: %s at %s", shape.Name, shape.Span())
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
