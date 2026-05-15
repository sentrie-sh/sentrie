// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Derive cycle detection shares expression walking with derive purity validation;
// see derive_expr_walk.go — new expression kinds must update forEachDeriveExprChild.
package index

import (
	"context"
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/dag"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (idx *Index) detectDeriveCycle(ctx context.Context) error {
	g := dag.New[*Derive]()
	for _, d := range idx.DerivesByFQN {
		if ctx.Err() != nil {
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		g.AddNode(d)
	}
	for _, d := range idx.DerivesByFQN {
		if ctx.Err() != nil {
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		for _, callee := range deriveCallees(idx, d) {
			if callee == d {
				continue
			}
			if err := g.AddEdge(d, callee); err != nil {
				return fmt.Errorf("derive dependency: %w", err)
			}
		}
	}
	if paths := g.DetectFirstCycle(); len(paths) > 0 {
		ps := make([]string, 0, len(paths))
		for _, n := range paths {
			ps = append(ps, n.FQN.String())
		}
		return fmt.Errorf("cyclic derive dependency: %s: %w", strings.Join(ps, " -> "), xerr.ErrIndex)
	}
	return nil
}

func deriveCallees(idx *Index, d *Derive) []*Derive {
	var out []*Derive
	_ = walkDeriveExprDFS(d.Lambda.Body, func(e ast.Expression) error {
		ce, ok := e.(*ast.CallExpression)
		if !ok {
			return nil
		}
		if id, ok := ce.Callee.(*ast.Identifier); ok {
			if t := d.DefineShort[id.Value]; t != nil {
				out = append(out, t)
			}
		}
		if fqn := ast.SlashCalleeFQNS(ce.Callee); fqn != "" {
			if t := d.DefineFQN[fqn]; t != nil {
				out = append(out, t)
			} else if t2, ok := idx.DerivesByFQN[fqn]; ok {
				out = append(out, t2)
			}
		}
		return nil
	})
	return out
}
