// SPDX-License-Identifier: Apache-2.0

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
	"cmp"
	"context"
	"strings"

	"github.com/pkg/errors"
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
		idx.validationError = errors.Wrapf(idx.validationError, "validation error")
		idx.validated = 1

		if idx.validationError != nil {
			// we couldn't validate the index, so we can't commit
			return
		}

		if err := idx.Commit(ctx); err != nil {
			return
		}
	})
	return idx.validationError
}

func (idx *Index) IsValid(ctx context.Context) error {
	return idx.Validate(ctx)
}

func (idx *Index) validate(ctx context.Context) error {
	// Check for self-references in rules and shapes
	if err := idx.detectReferenceCycle(ctx); err != nil {
		return err
	}
	rg, err := idx.detectRuleCycle(ctx)
	if err != nil {
		return err
	}
	sg, err := idx.detectShapeCycle(ctx)
	if err != nil {
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
			return errors.Wrapf(ErrIndex, "validation cancelled")
		default:
		}

		for _, policy := range ns.Policies {
			g := dag.New[String]()
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}

			for _, rule := range policy.Rules {
				if ctx.Err() != nil {
					return errors.Wrapf(ErrIndex, "validation cancelled")
				}
				g.AddNode(String(rule.Name))
				addNodes(g, []ast.Node{rule.Default, rule.When, rule.Body}, String(rule.Name), policy)
			}

			for _, let := range policy.Lets {
				g.AddNode(String(let.Name))
				addNodes(g, []ast.Node{let.Value}, String(let.Name), policy)
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

func addNodes(g dag.G[String], nodes []ast.Node, referedBy String, policy *Policy) {
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
			addNodes(g, []ast.Node{n.Body}, String(n.RuleName), policy)
		case *ast.VarDeclaration:
			g.AddNode(String(n.Name))
			_ = g.AddEdge(String(referedBy.String()), String(n.Name))
			addNodes(g, []ast.Node{n.Value}, String(n.Name), policy)
		case *ast.CallExpression:
			addNodes(g, []ast.Node{n.Callee}, referedBy, policy)
			for _, arg := range n.Arguments {
				addNodes(g, []ast.Node{arg}, referedBy, policy)
			}
		case *ast.InfixExpression:
			addNodes(g, []ast.Node{n.Left, n.Right}, referedBy, policy)
		case *ast.UnaryExpression:
			addNodes(g, []ast.Node{n.Right}, referedBy, policy)
		case *ast.TernaryExpression:
			addNodes(g, []ast.Node{n.Condition, n.ThenBranch, n.ElseBranch}, referedBy, policy)
		case *ast.BlockExpression:
			for _, stmt := range n.Statements {
				addNodes(g, []ast.Node{stmt}, referedBy, policy)
			}
			// Also check the yield expression
			addNodes(g, []ast.Node{n.Yield}, referedBy, policy)
		case *ast.ListLiteral:
			for _, elem := range n.Values {
				addNodes(g, []ast.Node{elem}, referedBy, policy)
			}
		case *ast.MapLiteral:
			for _, entry := range n.Entries {
				addNodes(g, []ast.Node{entry.Value}, referedBy, policy)
			}
		case *ast.FieldAccessExpression:
			addNodes(g, []ast.Node{n.Left}, referedBy, policy)
		case *ast.ImportClause:
			// Import clauses don't contain self-references
		default:
			// For any other node types, we don't need to check them
		}
	}
}

// containsSelfReference recursively checks if an AST node contains a self-reference
// func containsSelfReference(node ast.Node, ident string) bool {
// 	switch n := node.(type) {
// 	case *ast.Identifier:
// 		return n.Value == ident
// 	case *ast.CallExpression:
// 		if containsSelfReference(n.Callee, ident) {
// 			return true
// 		}
// 		for _, arg := range n.Arguments {
// 			if containsSelfReference(arg, ident) {
// 				return true
// 			}
// 		}
// 	case *ast.InfixExpression:
// 		return containsSelfReference(n.Left, ident) || containsSelfReference(n.Right, ident)
// 	case *ast.UnaryExpression:
// 		return containsSelfReference(n.Right, ident)
// 	case *ast.TernaryExpression:
// 		return containsSelfReference(n.Condition, ident) ||
// 			containsSelfReference(n.ThenBranch, ident) ||
// 			containsSelfReference(n.ElseBranch, ident)
// 	case *ast.BlockExpression:
// 		for _, stmt := range n.Statements {
// 			if containsSelfReference(stmt, ident) {
// 				return true
// 			}
// 		}
// 		// Also check the yield expression
// 		if n.Yield != nil && containsSelfReference(n.Yield, ident) {
// 			return true
// 		}
// 	case *ast.ListLiteral:
// 		for _, elem := range n.Values {
// 			if containsSelfReference(elem, ident) {
// 				return true
// 			}
// 		}
// 	case *ast.MapLiteral:
// 		for _, entry := range n.Entries {
// 			if containsSelfReference(entry.Value, ident) {
// 				return true
// 			}
// 		}
// 	case *ast.FieldAccessExpression:
// 		return containsSelfReference(n.Left, ident)
// 	case *ast.ImportClause:
// 		// Import clauses don't contain self-references
// 		return false
// 	default:
// 		// For any other node types, we don't need to check them
// 		return false
// 	}
// 	return false
// }

func (idx *Index) detectRuleCycle(ctx context.Context) (dag.G[*Rule], error) {
	ruleDag := dag.New[*Rule]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return nil, errors.Wrapf(ErrIndex, "validation cancelled")
		default:
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
			}
			for _, rule := range policy.Rules {
				ruleDag.AddNode(rule)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges
	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return nil, errors.Wrapf(ErrIndex, "validation cancelled")
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
			}
			// add the edges for the policy rules
			for _, rule := range policy.Rules {
				if ctx.Err() != nil {
					return nil, errors.Wrapf(ErrIndex, "validation cancelled")
				}
				if importClause, ok := rule.Body.(*ast.ImportClause); ok {
					var ns, pol string
					if len(importClause.FromPolicyFQN.Parts) == 1 {
						// we only have a policy name - the namespace is the current policy's namespace
						ns = policy.Namespace.FQN.String()
						pol = importClause.FromPolicyFQN.Parts[0]
					} else {
						// we have a namespace and policy name
						ns = strings.Join(importClause.FromPolicyFQN.Parts[:len(importClause.FromPolicyFQN.Parts)-1], ast.FQNSeparator)
						pol = importClause.FromPolicyFQN.Parts[len(importClause.FromPolicyFQN.Parts)-1]
					}

					p, err := idx.ResolvePolicy(ns, pol)
					if err != nil {
						return nil, errors.Wrapf(ErrIndex, "error resolving policy: %s", err)
					}
					if err := ruleDag.AddEdge(rule, p.Rules[importClause.RuleToImport]); err != nil {
						return nil, errors.Wrapf(ErrIndex, "error adding edge: %s", err)
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
		return nil, errors.Wrapf(ErrIndex, "detected cyclic dependency in rules: %s", strings.Join(pathStr, " -> "))
	}

	return ruleDag, nil
}

func (idx *Index) detectShapeCycle(ctx context.Context) (dag.G[*Shape], error) {
	shapeDag := dag.New[*Shape]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return nil, errors.Wrapf(ErrIndex, "validation cancelled")
		default:
		}

		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
			}
			shapeDag.AddNode(shape)
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
			}
			for _, shape := range policy.Shapes {
				shapeDag.AddNode(shape)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges

	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return nil, errors.Wrapf(ErrIndex, "validation cancelled")
		}
		// add the edges for the namespace shapes
		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
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
				return nil, errors.Wrapf(ErrIndex, "error resolving shape: %s", err)
			}
			// find the shape with the FQN
			if err := shapeDag.AddEdge(shape, withShape); err != nil {
				return nil, errors.Wrapf(ErrIndex, "error adding edge: %s", err)
			}
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return nil, errors.Wrapf(ErrIndex, "validation cancelled")
			}
			// add the edges for the policy shapes
			for _, shape := range policy.Shapes {
				if shape.Model != nil && shape.Model.WithFQN != nil && !shape.Model.WithFQN.IsEmpty() {
					// find the shape with the FQN
					withShape, ok := ns.Shapes[shape.Model.WithFQN.String()]
					if !ok {
						return nil, errors.Wrapf(ErrIndex, "shape not found: %s at %s", shape.Model.WithFQN.String(), shape.Statement.Span().String())
					}
					if err := shapeDag.AddEdge(shape, withShape); err != nil {
						return nil, errors.Wrapf(ErrIndex, "error adding edge: %s", err)
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
		return nil, errors.Wrapf(ErrIndex, "detected cyclic dependencies in shapes: %s", strings.Join(pathStr, " -> "))
	}

	return shapeDag, nil
}
