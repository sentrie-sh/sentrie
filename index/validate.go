package index

import (
	"context"
	"strings"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/dag"
	"github.com/pkg/errors"
)

// Validate the index for consistency and correctness.
// Checks for:
// - Cyclic dependencies
func (idx *Index) Validate(ctx context.Context) error {
	if err := idx.detectRuleCycle(ctx); err != nil {
		return err
	}
	if err := idx.detectShapeCycle(ctx); err != nil {
		return err
	}
	return nil
}

func (idx *Index) detectRuleCycle(ctx context.Context) error {
	ruleDag := dag.New[*Rule]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return errors.Wrapf(ErrIndex, "validation cancelled")
		default:
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			for _, rule := range policy.Rules {
				ruleDag.AddNode(rule)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges
	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return errors.Wrapf(ErrIndex, "validation cancelled")
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			// add the edges for the policy rules
			for _, rule := range policy.Rules {
				if ctx.Err() != nil {
					return errors.Wrapf(ErrIndex, "validation cancelled")
				}
				if importClause, ok := rule.Body.(*ast.ImportClause); ok {
					var ns, pol string
					if len(importClause.FromPolicyFQN) == 1 {
						// we only have a policy name - the namespace is the current policy's namespace
						ns = policy.Namespace.FQN.String()
						pol = importClause.FromPolicyFQN[0]
					} else {
						// we have a namespace and policy name
						ns = strings.Join(importClause.FromPolicyFQN[:len(importClause.FromPolicyFQN)-1], ast.FQNSeparator)
						pol = importClause.FromPolicyFQN[len(importClause.FromPolicyFQN)-1]
					}

					p, err := idx.ResolvePolicy(ns, pol)
					if err != nil {
						return errors.Wrapf(ErrIndex, "error resolving policy: %s", err)
					}
					if err := ruleDag.AddEdge(rule, p.Rules[importClause.RuleToImport]); err != nil {
						return errors.Wrapf(ErrIndex, "error adding edge: %s", err)
					}
				}
			}
		}
	}

	// check for cyclic dependencies
	if paths := ruleDag.DetectAllCycles(); len(paths) > 0 {
		pathStr := make([]string, 0, len(paths[0]))
		for _, node := range paths[0] {
			pathStr = append(pathStr, node.String())
		}
		return errors.Wrapf(ErrIndex, "detected cyclic dependency in rules: %s", strings.Join(pathStr, " -> "))
	}

	return nil
}

func (idx *Index) detectShapeCycle(ctx context.Context) error {
	shapeDag := dag.New[*Shape]()

	for _, ns := range idx.Namespaces {
		select {
		case <-ctx.Done():
			return errors.Wrapf(ErrIndex, "validation cancelled")
		default:
		}

		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			shapeDag.AddNode(shape)
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			for _, shape := range policy.Shapes {
				shapeDag.AddNode(shape)
			}
		}
	}

	// now that we added all the nodes, we need to add the edges

	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return errors.Wrapf(ErrIndex, "validation cancelled")
		}
		// add the edges for the namespace shapes
		for _, shape := range ns.Shapes {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			if shape.Complex != nil && len(shape.Complex.WithFQN) > 0 {
				// find the shape with the FQN
				withShape, ok := ns.Shapes[shape.Complex.WithFQN.String()]
				if !ok {
					return errors.Wrapf(ErrIndex, "shape not found: %s at %s", shape.Complex.WithFQN.String(), shape.Node.Pos)
				}
				if err := shapeDag.AddEdge(shape, withShape); err != nil {
					return errors.Wrapf(ErrIndex, "error adding edge: %s", err)
				}
			}
		}

		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return errors.Wrapf(ErrIndex, "validation cancelled")
			}
			// add the edges for the policy shapes
			for _, shape := range policy.Shapes {
				if shape.Complex != nil && len(shape.Complex.WithFQN) > 0 {
					// find the shape with the FQN
					withShape, ok := ns.Shapes[shape.Complex.WithFQN.String()]
					if !ok {
						return errors.Wrapf(ErrIndex, "shape not found: %s at %s", shape.Complex.WithFQN.String(), shape.Node.Pos)
					}
					if err := shapeDag.AddEdge(shape, withShape); err != nil {
						return errors.Wrapf(ErrIndex, "error adding edge: %s", err)
					}
				}
			}
		}
	}

	// check for cyclic dependencies
	if paths := shapeDag.DetectAllCycles(); len(paths) > 0 {
		pathStr := make([]string, 0, len(paths[0]))
		for _, node := range paths[0] {
			pathStr = append(pathStr, node.String())
		}
		return errors.Wrapf(ErrIndex, "detected cyclic dependencies in shapes: %s", strings.Join(pathStr, " -> "))
	}

	return nil
}
