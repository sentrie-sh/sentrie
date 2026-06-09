// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
)

// walkDeriveExprDFS walks every nested expression under root in pre-order.
func walkDeriveExprDFS(root ast.Expression, visit func(ast.Expression) error) error {
	var walk func(ast.Expression) error
	walk = func(e ast.Expression) error {
		if e == nil {
			return nil
		}
		if err := visit(e); err != nil {
			return err
		}
		return forEachDeriveExprChild(e, walk)
	}
	return walk(root)
}

func forEachDeriveExprChild(e ast.Expression, yield func(ast.Expression) error) error {
	switch n := e.(type) {
	case *ast.Identifier, *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.NullLiteral, *ast.TrinaryLiteral, *ast.PipelineHoleExpression:
		return nil
	case *ast.CallExpression:
		if err := yield(n.Callee); err != nil {
			return err
		}
		for _, a := range n.Arguments {
			if err := yield(a); err != nil {
				return err
			}
		}
		return nil
	case *ast.InfixExpression:
		if err := yield(n.Left); err != nil {
			return err
		}
		return yield(n.Right)
	case *ast.UnaryExpression:
		return yield(n.Right)
	case *ast.TernaryExpression:
		if n.Elvis {
			if err := yield(n.Condition); err != nil {
				return err
			}
			return yield(n.ElseBranch)
		}
		if err := yield(n.Condition); err != nil {
			return err
		}
		if err := yield(n.ThenBranch); err != nil {
			return err
		}
		return yield(n.ElseBranch)
	case *ast.BlockExpression:
		for _, st := range n.Statements {
			vd, ok := st.(*ast.VarDeclaration)
			if !ok {
				return fmt.Errorf("derive walk: unsupported statement %T", st)
			}
			if err := yield(vd.Value); err != nil {
				return err
			}
		}
		return yield(n.Yield)
	case *ast.LambdaExpression:
		return yield(n.Body)
	case *ast.FieldAccessExpression:
		return yield(n.Left)
	case *ast.IndexAccessExpression:
		if err := yield(n.Left); err != nil {
			return err
		}
		return yield(n.Index)
	case *ast.ListLiteral:
		for _, v := range n.Values {
			if err := yield(v); err != nil {
				return err
			}
		}
		return nil
	case *ast.MapLiteral:
		for _, kv := range n.Entries {
			if err := yield(kv.Key); err != nil {
				return err
			}
			if err := yield(kv.Value); err != nil {
				return err
			}
		}
		return nil
	case *ast.CastExpression:
		return yield(n.Expr)
	case *ast.TransformExpression:
		return yield(n.Argument)
	case *ast.IsDefinedExpression:
		return yield(n.Left)
	case *ast.IsEmptyExpression:
		return yield(n.Left)
	case *ast.TrailingCommentExpression:
		return yield(n.Wrap)
	case *ast.PrecedingCommentExpression:
		return yield(n.Wrap)
	default:
		return fmt.Errorf("derive walk: unsupported expression %T", e)
	}
}
