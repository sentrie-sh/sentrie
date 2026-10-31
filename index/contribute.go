// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

// Contribution is one `contribute <rule> [when <expr>] <expr>` statement,
// attached to the rule it adds to. A rule's contributions are kept in
// declaration order, which is the order the runtime combines them in.
type Contribution struct {
	Node *ast.ContributeStatement
	Rule *Rule
	When ast.Expression
	Body ast.Expression
}

func (c *Contribution) Span() tokens.Range {
	return c.Node.Span()
}

// ContributionKind is how a rule combines its contributions.
type ContributionKind int

const (
	// ContributionNone: the rule's declared type cannot take contributions.
	ContributionNone ContributionKind = iota
	// ContributionList: list[T]; each contribution appends one element.
	ContributionList
	// ContributionBool: boolean or trinary; contributions are OR-ed in.
	ContributionBool
)

// ContributionKindOf reports how a rule declared with typeRef combines
// contributions, and the type each contribution must yield: the element type
// for a list, the declared type itself for a boolean.
func ContributionKindOf(typeRef ast.TypeRef) (ContributionKind, ast.TypeRef) {
	if typeRef == nil {
		return ContributionNone, nil
	}
	switch t := ast.UnwrapNullableTypeRef(typeRef).(type) {
	case *ast.ListTypeRef:
		return ContributionList, t.ElemType
	case *ast.TrinaryTypeRef:
		return ContributionBool, t
	}
	return ContributionNone, nil
}

// addContribution attaches stmt to the rule it names. The rule must already be
// declared in this policy (as `export` requires), must declare a list[T],
// boolean or trinary type, and must not delegate to an imported decision.
func (p *Policy) addContribution(stmt *ast.ContributeStatement) error {
	r, ok := p.Rules[stmt.RuleName]
	if !ok {
		return fmt.Errorf("cannot contribute to unknown rule '%s' at %s: rules must be declared with 'rule' before they take contributions: %w", stmt.RuleName, stmt.Span(), xerr.ErrIndex)
	}
	kind, yieldType := ContributionKindOf(r.Node.Type)
	if kind == ContributionNone {
		declared := "no type"
		if r.Node.Type != nil {
			declared = "type " + r.Node.Type.String()
		}
		return fmt.Errorf("cannot contribute to rule '%s' at %s: the rule declares %s, but only a rule declared as list[T], boolean or trinary takes contributions (declared at %s): %w", stmt.RuleName, stmt.Span(), declared, r.Span(), xerr.ErrIndex)
	}
	if _, isImport := r.Body.(*ast.ImportClause); isImport {
		return fmt.Errorf("cannot contribute to rule '%s' at %s: the rule imports its decision from another policy: %w", stmt.RuleName, stmt.Span(), xerr.ErrIndex)
	}
	if lit, litKind, ok := literalYieldMismatch(stmt.Body, yieldType); ok {
		return fmt.Errorf("contribution to rule '%s' at %s yields a %s, but the rule takes %s: %w", stmt.RuleName, lit.Span(), litKind, yieldType.String(), xerr.ErrIndex)
	}

	r.Contributions = append(r.Contributions, &Contribution{
		Node: stmt,
		Rule: r,
		When: stmt.When,
		Body: stmt.Body,
	})
	return nil
}

// literalYieldMismatch catches the statically obvious type errors: a body that
// is, or is a block that yields, a literal whose kind cannot satisfy want.
// Anything else is checked against the declared type at runtime.
func literalYieldMismatch(body ast.Expression, want ast.TypeRef) (ast.Expression, string, bool) {
	yield := body
	if b, ok := body.(*ast.BlockExpression); ok {
		yield = b.Yield
	}
	var got string
	switch yield.(type) {
	case *ast.StringLiteral:
		got = "string"
	case *ast.IntegerLiteral, *ast.FloatLiteral:
		got = "number"
	case *ast.TrinaryLiteral:
		got = "boolean"
	case *ast.ListLiteral:
		got = "list"
	case *ast.MapLiteral:
		got = "dict"
	case *ast.NullLiteral:
		if ast.IsNullableTypeRef(want) {
			return nil, "", false
		}
		got = "null"
	default:
		return nil, "", false
	}
	var accepts string
	switch ast.UnwrapNullableTypeRef(want).(type) {
	case *ast.StringTypeRef:
		accepts = "string"
	case *ast.NumberTypeRef:
		accepts = "number"
	case *ast.TrinaryTypeRef:
		accepts = "boolean"
	case *ast.ListTypeRef:
		accepts = "list"
	case *ast.RecordTypeRef:
		// A record is a fixed-length list at runtime.
		accepts = "list"
	case *ast.DictTypeRef:
		accepts = "dict"
	default:
		// document, shape and anything else: leave it to the runtime.
		return nil, "", false
	}
	if got == accepts {
		return nil, "", false
	}
	return yield, got, true
}

// detectContributionSelfReference rejects a contribution whose guard or body
// reads the rule it contributes to. The reference-cycle graph drops self
// edges, so without this the loop would only be found when the rule is
// evaluated. A lambda parameter or block let of the same name shadows the
// rule and is not a reference.
func (idx *Index) detectContributionSelfReference(ctx context.Context) error {
	for _, ns := range idx.Namespaces {
		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			for _, rule := range policy.Rules {
				for _, c := range rule.Contributions {
					var found *ast.Identifier
					visit := func(k *kindCheckCtx, e ast.Expression) {
						if id, ok := e.(*ast.Identifier); ok && found == nil && id.Value == rule.Name && !k.isShadowed(id.Value) {
							found = id
						}
					}
					k := newRuleKindCheckCtx(idx, policy)
					k.walkScopedExprs(c.When, visit)
					k.walkScopedExprs(c.Body, visit)
					if found != nil {
						return fmt.Errorf("contribution to rule '%s' at %s reads the rule itself at %s: %w", rule.Name, c.Span(), found.Span(), xerr.ErrIndex)
					}
				}
			}
		}
	}
	return nil
}
