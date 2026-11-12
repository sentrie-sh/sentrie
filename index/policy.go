// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
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
	Data        map[string]*ast.DataStatement // alias -> pack data declaration
	Rules       map[string]*Rule
	Derives     map[string]*Derive
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
	return fmt.Errorf("'%s' must appear before rules, exports, lets, derives, and shapes at %s: %w", keyword, at, xerr.ErrIndex)
}

type policyPhase int

const (
	policyPhaseMeta policyPhase = iota
	policyPhaseFacts
	policyPhaseUses
	policyPhaseBody
)

func createPolicy(ns *Namespace, policy *ast.PolicyStatement, program *ast.Program, idx *Index, nsDerives map[string]*Derive) (*Policy, error) {
	if nsDerives == nil {
		nsDerives = map[string]*Derive{}
	}
	p := &Policy{
		Statement:       policy,
		Namespace:       ns,
		Name:            policy.Name,
		FQN:             ast.CreateFQN(ns.FQN, policy.Name),
		FilePath:        program.Reference,
		Statements:      policy.Statements,
		Lets:            make(map[string]*ast.VarDeclaration),
		Facts:           make(map[string]*ast.FactStatement),
		Data:            make(map[string]*ast.DataStatement),
		Rules:           make(map[string]*Rule),
		Derives:         make(map[string]*Derive),
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
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyMetadataContiguous)
			}
			if titleAt != nil {
				return nil, xerr.ErrConflict("policy title", stmt.Span(), titleAt.Span())
			}
			trimmed := strings.TrimSpace(stmt.Value)
			if trimmed == "" {
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyEmptyTitle)
			}
			t := trimmed
			p.Title = &t
			titleAt = stmt

		case *ast.DescriptionStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("description", stmt.Span().String())
				}
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyMetadataContiguous)
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
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyMetadataContiguous)
			}
			if versionAt != nil {
				return nil, xerr.ErrConflict("policy version", stmt.Span(), versionAt.Span())
			}
			p.VersionLiteral = stmt.Literal
			// SemVer is validated on TrimSpace(literal); VersionLiteral stays verbatim for display/diagnostics.
			ver, err := semver.NewVersion(strings.TrimSpace(stmt.Literal))
			if err != nil {
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyInvalidVersion)
			}
			p.Version = ver
			versionAt = stmt

		case *ast.TagStatement:
			if phase != policyPhaseMeta {
				if phase == policyPhaseBody {
					return nil, latePolicyHeaderErr("tag", stmt.Span().String())
				}
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyMetadataContiguous)
			}
			key := strings.TrimSpace(stmt.Key)
			if key == "" {
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyEmptyTagKey)
			}
			p.TagPairs = append(p.TagPairs, PolicyTagPair{Key: key, Value: stmt.Value})

		case *ast.FactStatement:
			switch phase {
			case policyPhaseMeta:
				phase = policyPhaseFacts
			case policyPhaseFacts:
			case policyPhaseUses:
				return nil, fmt.Errorf("at %s: %w", stmt.Span(), xerr.ErrPolicyFactAfterUse)
			case policyPhaseBody:
				return nil, latePolicyHeaderErr("fact", stmt.Span().String())
			}
			if err := p.AddFact(stmt); err != nil {
				return nil, err
			}

		case *ast.DataStatement:
			// A data declaration is a policy input like a fact, but it may sit
			// anywhere in the header: among facts, or among uses.
			switch phase {
			case policyPhaseMeta:
				phase = policyPhaseFacts
			case policyPhaseFacts, policyPhaseUses:
			case policyPhaseBody:
				return nil, latePolicyHeaderErr("data", stmt.Span().String())
			}
			if err := p.AddData(stmt); err != nil {
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
				return nil, fmt.Errorf("cannot rebind to existing alias '%s' at %s: %w", stmt.As, stmt.Span(), xerr.ErrIndex)
			}
			if d, ok := p.Data[stmt.As]; ok {
				return nil, xerr.ErrConflict("use alias", stmt.Span(), d.Span())
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

		case *ast.ContributeStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if err := p.addContribution(stmt); err != nil {
				return nil, err
			}

		case *ast.RuleExportStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if _, ok := p.Rules[stmt.Of]; !ok {
				return nil, fmt.Errorf("cannot export unknown rule: '%s' at %s: %w", stmt.Of, stmt.Span(), xerr.ErrIndex)
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

		case *ast.DeriveStatement:
			if phase != policyPhaseBody {
				phase = policyPhaseBody
			}
			if _, err := p.addDerive(idx, stmt, nsDerives, p.Derives); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unsupported statement in policy at %s: %w", stmt.Span(), xerr.ErrIndex)
		}
	}

	p.TagsByKey = buildTagsByKey(p.TagPairs)

	if len(p.RuleExports) == 0 {
		return nil, fmt.Errorf("policy '%s' at '%s' does not export any rules: %w", policy.Name, policy.Span(), xerr.ErrIndex)
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
		return fmt.Errorf("failed to create shape: %s at %s: %w", shape.Name, shape.Span(), xerr.ErrIndex)
	}

	p.Shapes[shape.Name] = s
	return nil
}

func (p *Policy) AddFact(fact *ast.FactStatement) error {
	if seen, ok := p.seenIdentifiers[fact.Alias]; ok {
		return xerr.ErrConflict("fact declaration", fact.Span(), seen.Span())
	}

	// The declared name is the key a caller supplies the fact under, so two
	// facts may not share it even when their aliases differ (#147).
	if seen := p.FactByName(fact.Name); seen != nil {
		return xerr.ErrConflict(fmt.Sprintf("fact name '%s'", fact.Name), fact.Span(), seen.Span())
	}

	// Required facts (not optional) cannot have default values
	if !fact.Optional && fact.Default != nil {
		return xerr.ErrInvalidInvocation(fmt.Sprintf("required fact '%s' at %s cannot have a default value", fact.Name, fact.Span()))
	}

	if err := checkFactDefault(fact); err != nil {
		return err
	}

	p.Facts[fact.Alias] = fact
	p.seenIdentifiers[fact.Alias] = fact
	return nil
}

// AddData registers a `data` declaration under its alias. The alias shares the
// policy's identifier space with facts, lets and rules, and may not reuse a
// `use` alias.
func (p *Policy) AddData(data *ast.DataStatement) error {
	if seen, ok := p.seenIdentifiers[data.Alias]; ok {
		return xerr.ErrConflict("data declaration", data.Span(), seen.Span())
	}
	if use, ok := p.Uses[data.Alias]; ok {
		return xerr.ErrConflict("data declaration", data.Span(), use.Span())
	}

	p.Data[data.Alias] = data
	p.seenIdentifiers[data.Alias] = data
	return nil
}

// FactByName returns the fact declared under name, the key a caller supplies
// it under (an HTTP request, `exec --facts`, an import's `with` clause), or
// nil. p.Facts is keyed by alias, the fact's local binding, so this is a scan;
// a policy declares few facts (#147).
func (p *Policy) FactByName(name string) *ast.FactStatement {
	for _, fact := range p.Facts {
		if fact.Name == name {
			return fact
		}
	}
	return nil
}

// FactAliasesByName maps each fact's declared name, the key a caller and an
// `import decision ... with <name> as ...` clause use, to its alias.
func (p *Policy) FactAliasesByName() map[string]string {
	out := make(map[string]string, len(p.Facts))
	for alias, fact := range p.Facts {
		out[fact.Name] = alias
	}
	return out
}

// checkFactDefault rejects a default whose type is statically wrong for the
// fact: a null default for a type that does not accept null, or a literal of
// the wrong kind (`fact n?: number default "x"`). It runs while the index is
// built, so `sentrie validate` reports it and no request evaluates it (#147).
// A default that is not a literal is checked against the type at runtime.
func checkFactDefault(fact *ast.FactStatement) error {
	if fact.Default == nil || fact.Type == nil {
		return nil
	}
	if _, ok := fact.Default.(*ast.NullLiteral); ok && !ast.AcceptsNull(fact.Type) {
		return fmt.Errorf("fact '%s' at %s cannot have a null default: its type %s does not accept null: %w", fact.Name, fact.Default.Span(), fact.Type.String(), xerr.ErrIndex)
	}
	if lit, litKind, ok := literalYieldMismatch(fact.Default, fact.Type); ok {
		return fmt.Errorf("default value of fact '%s' at %s is a %s, but the fact is declared %s: %w", fact.Name, lit.Span(), litKind, fact.Type.String(), xerr.ErrIndex)
	}
	return nil
}
