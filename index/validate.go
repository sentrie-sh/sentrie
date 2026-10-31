// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/dag"
	"github.com/sentrie-sh/sentrie/xerr"
)

// Validate the index for consistency and correctness.
// Checks for:
// - Cyclic dependencies
func (idx *Index) Validate(ctx context.Context) error {
	idx.validationOnce.Do(func() {
		idx.validationError = idx.validate(ctx)
		if idx.validationError != nil {
			idx.validationError = fmt.Errorf("validation error: %w", idx.validationError)
		}
		idx.validated = 1

		if idx.validationError != nil {
			// we couldn't validate the index, so we can't commit
			return
		}

		if err := idx.Commit(ctx); err != nil {
			idx.validationError = err
			return
		}
	})
	return idx.validationError
}

func (idx *Index) IsValid(ctx context.Context) error {
	return idx.Validate(ctx)
}

func (idx *Index) validate(ctx context.Context) error {
	// Settle `x is Name` first: the walks below must not read a shape name
	// as a reference to a value.
	if err := idx.resolveIsExpressions(ctx); err != nil {
		return err
	}
	// Check for self-references in rules and shapes
	if err := idx.detectReferenceCycle(ctx); err != nil {
		return err
	}
	rg, err := idx.detectRuleCycle(ctx)
	if err != nil {
		return err
	}
	if err := idx.checkImportWiths(ctx); err != nil {
		return err
	}
	sg, err := idx.detectShapeCycle(ctx)
	if err != nil {
		return err
	}
	if err := idx.detectDeriveCycle(ctx); err != nil {
		return err
	}
	if err := idx.validateDerivePurity(); err != nil {
		return err
	}
	if err := idx.checkBuiltinCalls(ctx); err != nil {
		return err
	}

	idx.ruleDag = rg
	idx.shapeDag = sg

	return nil
}

type String string

func (s String) String() string {
	return string(s)
}

func (idx *Index) detectReferenceCycle(ctx context.Context) error {
	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		default:
		}

		for _, policy := range ns.Policies {
			g := dag.New[String]()
			if ctx.Err() != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}

			for _, rule := range policy.Rules {
				if ctx.Err() != nil {
					return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
				}
				g.AddNode(String(rule.Name))
				if err := addNodes(g, []ast.Node{rule.Default, rule.When, rule.Body}, String(rule.Name), policy); err != nil {
					return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
				}
			}

			for _, let := range policy.Lets {
				g.AddNode(String(let.Name))
				if err := addNodes(g, []ast.Node{let.Value}, String(let.Name), policy); err != nil {
					return fmt.Errorf("%s: %w", err, xerr.ErrIndex)
				}
			}

			cycles := g.DetectFirstCycle()
			if len(cycles) > 0 {
				c := make([]string, 0, len(cycles))
				for _, node := range cycles {
					c = append(c, node.String())
				}
				return xerr.ErrInfiniteRecursion(c)
			}
		}
	}
	return nil
}

// addNodes adds to g an edge from referedBy to every rule or let that the given
// nodes reference. Identifiers add edges; a rule or let declared inside a body
// adds its own node and is walked with itself as the referrer. Every other node
// is descended into through forEachExprChild, so no expression kind is skipped.
// It fails when the walker meets a node kind it does not know.
func addNodes(g dag.G[String], nodes []ast.Node, referedBy String, policy *Policy) error {
	for _, node := range nodes {
		if node == nil {
			continue
		}

		switch n := node.(type) {
		case *ast.Identifier:
			_ = g.AddEdge(String(referedBy.String()), String(n.Value))
		case *ast.RuleStatement:
			g.AddNode(String(n.RuleName))
			_ = g.AddEdge(String(referedBy.String()), String(n.RuleName))
			if err := addNodes(g, []ast.Node{n.Default, n.When, n.Body}, String(n.RuleName), policy); err != nil {
				return err
			}
		case *ast.VarDeclaration:
			g.AddNode(String(n.Name))
			_ = g.AddEdge(String(referedBy.String()), String(n.Name))
			if err := addNodes(g, []ast.Node{n.Value}, String(n.Name), policy); err != nil {
				return err
			}
		default:
			err := forEachExprChild(node, func(child ast.Node) error {
				return addNodes(g, []ast.Node{child}, referedBy, policy)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (idx *Index) detectRuleCycle(ctx context.Context) (dag.G[*Rule], error) {
	ruleDag := dag.New[*Rule]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		default:
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			for _, rule := range policy.Rules {
				ruleDag.AddNode(rule)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges
	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			// add the edges for the policy rules
			for _, rule := range policy.Rules {
				if ctx.Err() != nil {
					return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
				}
				if importClause, ok := rule.Body.(*ast.ImportClause); ok {
					ns, pol := importTarget(policy, importClause)

					p, err := idx.ResolvePolicy(ns, pol)
					if err != nil {
						return nil, fmt.Errorf("error resolving policy: %s: %w", err, xerr.ErrIndex)
					}
					if err := ruleDag.AddEdge(rule, p.Rules[importClause.RuleToImport]); err != nil {
						return nil, fmt.Errorf("error adding edge: %s: %w", err, xerr.ErrIndex)
					}
				}
			}
		}
	}

	// check for cyclic dependencies
	if paths := ruleDag.DetectFirstCycle(); len(paths) > 0 {
		pathStr := make([]string, 0, len(paths))
		for _, node := range paths {
			pathStr = append(pathStr, node.String())
		}
		return nil, fmt.Errorf("detected cyclic dependency in rules: %s: %w", strings.Join(pathStr, " -> "), xerr.ErrIndex)
	}

	return ruleDag, nil
}

func (idx *Index) detectShapeCycle(ctx context.Context) (dag.G[*Shape], error) {
	shapeDag := dag.New[*Shape]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		default:
		}

		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			shapeDag.AddNode(shape)
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			for _, shape := range policy.Shapes {
				shapeDag.AddNode(shape)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges

	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		// add the edges for the namespace shapes
		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			if shape.Model == nil || shape.Model.WithFQN == nil || shape.Model.WithFQN.IsEmpty() {
				continue
			}

			withShape, err := idx.ResolveShape(
				cmp.Or( // if there's no parent namespace, use the namespace FQN
					shape.Model.WithFQN.Parent().String(),
					shape.Namespace.FQN.String(),
				),
				shape.Model.WithFQN.LastSegment())
			if err != nil {
				return nil, fmt.Errorf("error resolving shape: %s: %w", err, xerr.ErrIndex)
			}
			// find the shape with the FQN
			if err := shapeDag.AddEdge(shape, withShape); err != nil {
				return nil, fmt.Errorf("error adding edge: %s: %w", err, xerr.ErrIndex)
			}
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			// add the edges for the policy shapes
			for _, shape := range policy.Shapes {
				if shape.Model != nil && shape.Model.WithFQN != nil && !shape.Model.WithFQN.IsEmpty() {
					// find the shape with the FQN
					withShape, ok := ns.Shapes[shape.Model.WithFQN.String()]
					if !ok {
						return nil, fmt.Errorf("shape not found: %s at %s: %w", shape.Model.WithFQN.String(), shape.Statement.Span().String(), xerr.ErrIndex)
					}
					if err := shapeDag.AddEdge(shape, withShape); err != nil {
						return nil, fmt.Errorf("error adding edge: %s: %w", err, xerr.ErrIndex)
					}
				}
			}
		}
	}

	// check for cyclic dependencies
	if paths := shapeDag.DetectFirstCycle(); len(paths) > 0 {
		pathStr := make([]string, 0, len(paths))
		for _, node := range paths {
			pathStr = append(pathStr, node.String())
		}
		return nil, fmt.Errorf("detected cyclic dependencies in shapes: %s: %w", strings.Join(pathStr, " -> "), xerr.ErrIndex)
	}

	return shapeDag, nil
}

// importTarget returns the namespace and policy name an import clause refers
// to. A single-segment path names a policy in the importing policy's namespace.
func importTarget(from *Policy, imp *ast.ImportClause) (ns, pol string) {
	parts := imp.FromPolicyFQN.Parts
	if len(parts) == 1 {
		return from.Namespace.FQN.String(), parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ast.FQNSeparator), parts[len(parts)-1]
}
