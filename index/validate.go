package index

import (
	"cmp"
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/dag"
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
	if paths := ruleDag.DetectAllCycles(); len(paths) > 0 {
		pathStr := make([]string, 0, len(paths[0]))
		for _, node := range paths[0] {
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
			if shape.Model == nil || len(shape.Model.WithFQN) == 0 {
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
				if shape.Model != nil && len(shape.Model.WithFQN) > 0 {
					// find the shape with the FQN
					withShape, ok := ns.Shapes[shape.Model.WithFQN.String()]
					if !ok {
						return nil, errors.Wrapf(ErrIndex, "shape not found: %s at %s", shape.Model.WithFQN.String(), shape.Statement.Pos)
					}
					if err := shapeDag.AddEdge(shape, withShape); err != nil {
						return nil, errors.Wrapf(ErrIndex, "error adding edge: %s", err)
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
		return nil, errors.Wrapf(ErrIndex, "detected cyclic dependencies in shapes: %s", strings.Join(pathStr, " -> "))
	}

	return shapeDag, nil
}
