// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/runtime/derivepure"
)

func (idx *Index) validateDerivePurity() error {
	for _, d := range idx.DerivesByFQN {
		if err := validateDerivePure(idx, d); err != nil {
			return fmt.Errorf("derive %s: %w", d.FQN.String(), err)
		}
	}
	return nil
}

func validateDerivePure(idx *Index, d *Derive) error {
	lam := d.Lambda
	if lam == nil || lam.Body == nil {
		return nil
	}
	scope := make(map[string]struct{}, len(lam.Params))
	for _, p := range lam.Params {
		scope[p] = struct{}{}
	}
	return walkDeriveBlock(idx, d, lam.Body, scope)
}

func walkDeriveBlock(idx *Index, d *Derive, b *ast.BlockExpression, scope map[string]struct{}) error {
	if b == nil {
		return nil
	}
	for _, st := range b.Statements {
		vd, ok := st.(*ast.VarDeclaration)
		if !ok {
			return fmt.Errorf("derive body may only contain let declarations before yield (got %T)", st)
		}
		if err := walkDeriveExpr(idx, d, vd.Value, scope); err != nil {
			return err
		}
		scope = cloneScope(scope)
		scope[vd.Name] = struct{}{}
	}
	if err := yieldHasNoLambda(b.Yield); err != nil {
		return err
	}
	return walkDeriveExpr(idx, d, b.Yield, scope)
}

func cloneScope(scope map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(scope)+1)
	for k := range scope {
		out[k] = struct{}{}
	}
	return out
}

func yieldHasNoLambda(e ast.Expression) error {
	return scanLambdasOutsideCalls(e, false)
}

func scanLambdasOutsideCalls(e ast.Expression, underCall bool) error {
	if e == nil {
		return nil
	}
	switch n := e.(type) {
	case *ast.LambdaExpression:
		if !underCall {
			return fmt.Errorf("derive cannot yield a lambda value")
		}
		return scanLambdasOutsideCalls(n.Body, underCall)
	case *ast.CallExpression:
		if err := scanLambdasOutsideCalls(n.Callee, underCall); err != nil {
			return err
		}
		for _, a := range n.Arguments {
			if err := scanLambdasOutsideCalls(a, true); err != nil {
				return err
			}
		}
		return nil
	case *ast.InfixExpression:
		if err := scanLambdasOutsideCalls(n.Left, underCall); err != nil {
			return err
		}
		return scanLambdasOutsideCalls(n.Right, underCall)
	case *ast.UnaryExpression:
		return scanLambdasOutsideCalls(n.Right, underCall)
	case *ast.TernaryExpression:
		if n.Elvis {
			if err := scanLambdasOutsideCalls(n.Condition, underCall); err != nil {
				return err
			}
			return scanLambdasOutsideCalls(n.ElseBranch, underCall)
		}
		if err := scanLambdasOutsideCalls(n.Condition, underCall); err != nil {
			return err
		}
		if err := scanLambdasOutsideCalls(n.ThenBranch, underCall); err != nil {
			return err
		}
		return scanLambdasOutsideCalls(n.ElseBranch, underCall)
	case *ast.BlockExpression:
		for _, st := range n.Statements {
			vd, ok := st.(*ast.VarDeclaration)
			if !ok {
				return fmt.Errorf("derive walk: unsupported statement %T", st)
			}
			if err := scanLambdasOutsideCalls(vd.Value, underCall); err != nil {
				return err
			}
		}
		return scanLambdasOutsideCalls(n.Yield, underCall)
	case *ast.FieldAccessExpression:
		return scanLambdasOutsideCalls(n.Left, underCall)
	case *ast.IndexAccessExpression:
		if err := scanLambdasOutsideCalls(n.Left, underCall); err != nil {
			return err
		}
		return scanLambdasOutsideCalls(n.Index, underCall)
	case *ast.ListLiteral:
		for _, v := range n.Values {
			if err := scanLambdasOutsideCalls(v, underCall); err != nil {
				return err
			}
		}
		return nil
	case *ast.MapLiteral:
		for _, kv := range n.Entries {
			if err := scanLambdasOutsideCalls(kv.Key, underCall); err != nil {
				return err
			}
			if err := scanLambdasOutsideCalls(kv.Value, underCall); err != nil {
				return err
			}
		}
		return nil
	case *ast.CastExpression:
		return scanLambdasOutsideCalls(n.Expr, underCall)
	case *ast.TransformExpression:
		return scanLambdasOutsideCalls(n.Argument, underCall)
	case *ast.IsDefinedExpression:
		return scanLambdasOutsideCalls(n.Left, underCall)
	case *ast.IsEmptyExpression:
		return scanLambdasOutsideCalls(n.Left, underCall)
	case *ast.TrailingCommentExpression:
		return scanLambdasOutsideCalls(n.Wrap, underCall)
	case *ast.PrecedingCommentExpression:
		return scanLambdasOutsideCalls(n.Wrap, underCall)
	case *ast.Identifier, *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.NullLiteral, *ast.TrinaryLiteral, *ast.PipelineHoleExpression:
		return nil
	default:
		return fmt.Errorf("derive purity: unsupported expression in yield scan %T", e)
	}
}

func walkDeriveExpr(idx *Index, d *Derive, e ast.Expression, scope map[string]struct{}) error {
	if e == nil {
		return nil
	}
	switch n := e.(type) {
	case *ast.Identifier:
		return checkDeriveIdentifier(d, n.Value, scope)
	case *ast.CallExpression:
		return checkDeriveCall(idx, d, n, scope)
	case *ast.LambdaExpression:
		inner := cloneScope(scope)
		for _, p := range n.Params {
			inner[p] = struct{}{}
		}
		return walkDeriveBlock(idx, d, n.Body, inner)
	case *ast.BlockExpression:
		return walkDeriveBlock(idx, d, n, cloneScope(scope))
	case *ast.InfixExpression, *ast.UnaryExpression, *ast.TernaryExpression,
		*ast.FieldAccessExpression, *ast.IndexAccessExpression, *ast.ListLiteral, *ast.MapLiteral,
		*ast.CastExpression, *ast.TransformExpression, *ast.IsDefinedExpression, *ast.IsEmptyExpression,
		*ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.NullLiteral, *ast.TrinaryLiteral,
		*ast.PipelineHoleExpression, *ast.TrailingCommentExpression, *ast.PrecedingCommentExpression:
		return forEachDeriveExprChild(e, func(child ast.Expression) error {
			return walkDeriveExpr(idx, d, child, scope)
		})
	default:
		return fmt.Errorf("derive purity: unsupported expression %T", e)
	}
}

func checkDeriveIdentifier(d *Derive, name string, scope map[string]struct{}) error {
	if _, ok := scope[name]; ok {
		return nil
	}
	if d.DefineShort != nil && d.DefineShort[name] != nil {
		return fmt.Errorf("derive %q must be called as %s(...), not used as a bare value", name, name)
	}
	if d.Policy != nil {
		if d.Policy.Facts != nil {
			if _, ok := d.Policy.Facts[name]; ok {
				return fmt.Errorf("facts are not available inside a derive (%q)", name)
			}
		}
		if d.Policy.Rules != nil {
			if _, ok := d.Policy.Rules[name]; ok {
				return fmt.Errorf("rules cannot be referenced inside a derive (%q)", name)
			}
		}
	}
	return fmt.Errorf("identifier %q is not available in this derive (params, lets, pure builtins, and visible derives only)", name)
}

func checkDeriveCall(idx *Index, d *Derive, c *ast.CallExpression, scope map[string]struct{}) error {
	if _, ok := c.Callee.(*ast.FieldAccessExpression); ok {
		return fmt.Errorf("TypeScript module calls are not permitted inside a derive")
	}
	if fqn := ast.SlashCalleeFQNS(c.Callee); fqn != "" {
		parts := strings.Split(fqn, ast.FQNSeparator)
		if len(parts) >= 3 {
			if d.DefineFQN[fqn] != nil {
				for _, a := range c.Arguments {
					if err := walkDeriveExpr(idx, d, a, scope); err != nil {
						return err
					}
				}
				return nil
			}
			if _, ok := idx.DerivesByFQN[fqn]; ok {
				for _, a := range c.Arguments {
					if err := walkDeriveExpr(idx, d, a, scope); err != nil {
						return err
					}
				}
				return nil
			}
			return fmt.Errorf("unknown derive %q", fqn)
		}
	}
	if id, ok := c.Callee.(*ast.Identifier); ok {
		name := id.Value
		if _, ok := scope[name]; ok {
			for _, a := range c.Arguments {
				if err := walkDeriveExpr(idx, d, a, scope); err != nil {
					return err
				}
			}
			return nil
		}
		if d.DefineShort != nil && d.DefineShort[name] != nil {
			for _, a := range c.Arguments {
				if err := walkDeriveExpr(idx, d, a, scope); err != nil {
					return err
				}
			}
			return nil
		}
		if derivepure.IsPureBuiltin(name) {
			for _, a := range c.Arguments {
				if err := walkDeriveExpr(idx, d, a, scope); err != nil {
					return err
				}
			}
			return nil
		}
	}
	if err := walkDeriveExpr(idx, d, c.Callee, scope); err != nil {
		return err
	}
	for _, a := range c.Arguments {
		if err := walkDeriveExpr(idx, d, a, scope); err != nil {
			return err
		}
	}
	return fmt.Errorf("call is not permitted inside a derive (only visible derives and pure builtins)")
}
