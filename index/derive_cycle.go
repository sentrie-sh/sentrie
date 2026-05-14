// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
	var walk func(ast.Expression)
	walk = func(e ast.Expression) {
		if e == nil {
			return
		}
		switch n := e.(type) {
		case *ast.CallExpression:
			if id, ok := n.Callee.(*ast.Identifier); ok {
				if t := d.DefineShort[id.Value]; t != nil {
					out = append(out, t)
				}
			}
			if fqn := ast.SlashCalleeFQNS(n.Callee); fqn != "" {
				if t := d.DefineFQN[fqn]; t != nil {
					out = append(out, t)
				} else if t2, ok := idx.DerivesByFQN[fqn]; ok {
					out = append(out, t2)
				}
			}
			for _, a := range n.Arguments {
				walk(a)
			}
		case *ast.InfixExpression:
			walk(n.Left)
			walk(n.Right)
		case *ast.UnaryExpression:
			walk(n.Right)
		case *ast.TernaryExpression:
			if n.Elvis {
				walk(n.Condition)
				walk(n.ElseBranch)
			} else {
				walk(n.Condition)
				walk(n.ThenBranch)
				walk(n.ElseBranch)
			}
		case *ast.BlockExpression:
			for _, st := range n.Statements {
				if vd, ok := st.(*ast.VarDeclaration); ok {
					walk(vd.Value)
				}
			}
			walk(n.Yield)
		case *ast.LambdaExpression:
			walk(n.Body)
		case *ast.FieldAccessExpression:
			walk(n.Left)
		case *ast.IndexAccessExpression:
			walk(n.Left)
			walk(n.Index)
		case *ast.ListLiteral:
			for _, v := range n.Values {
				walk(v)
			}
		case *ast.MapLiteral:
			for _, kv := range n.Entries {
				walk(kv.Key)
				walk(kv.Value)
			}
		case *ast.CastExpression:
			walk(n.Expr)
		case *ast.TransformExpression:
			walk(n.Argument)
		default:
		}
	}
	walk(d.Lambda.Body)
	return out
}
