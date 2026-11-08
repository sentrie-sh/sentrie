// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
)

// forEachExprChild yields every direct child node of n that can hold a
// reference to a rule or let, for every expression kind and for the statements
// that appear inside a block (let, rule, comment). Unlike
// forEachDeriveExprChild it is not derive-specific: it descends into import
// clauses and their `with` values, and it never rejects a statement.
//
// It is the single place that knows the shape of the AST for reference
// analysis. A node kind it has no case for is an error, never a silent skip, so
// a new expression kind cannot be missed by a check built on it;
// TestForEachExprChildCoversEveryASTNode fails when the ast package gains a kind
// this function does not handle.
//
// Children that are nil are not yielded. A TypeRef operand (cast target, `is`
// type) is not an expression and is not yielded.
func forEachExprChild(n ast.Node, yield func(ast.Node) error) error {
	each := func(children ...ast.Node) error {
		for _, c := range children {
			if c == nil {
				continue
			}
			if err := yield(c); err != nil {
				return err
			}
		}
		return nil
	}

	switch n := n.(type) {
	case *ast.Identifier, *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.NullLiteral, *ast.TrinaryLiteral, *ast.PipelineHoleExpression,
		*ast.CommentStatement:
		return nil
	case *ast.CallExpression:
		if err := each(n.Callee); err != nil {
			return err
		}
		for _, a := range n.Arguments {
			if err := each(a); err != nil {
				return err
			}
		}
		return nil
	case *ast.InfixExpression:
		return each(n.Left, n.Right)
	case *ast.UnaryExpression:
		return each(n.Right)
	case *ast.TernaryExpression:
		if n.Elvis {
			return each(n.Condition, n.ElseBranch)
		}
		return each(n.Condition, n.ThenBranch, n.ElseBranch)
	case *ast.BlockExpression:
		for _, st := range n.Statements {
			if err := each(st); err != nil {
				return err
			}
		}
		return each(n.Yield)
	case *ast.LambdaExpression:
		return each(n.Body)
	case *ast.FieldAccessExpression:
		return each(n.Left)
	case *ast.IndexAccessExpression:
		return each(n.Left, n.Index)
	case *ast.ListLiteral:
		for _, v := range n.Values {
			if err := each(v); err != nil {
				return err
			}
		}
		return nil
	case *ast.MapLiteral:
		for _, kv := range n.Entries {
			if err := each(kv.Key, kv.Value); err != nil {
				return err
			}
		}
		return nil
	case *ast.CastExpression:
		return each(n.Expr)
	case *ast.TransformExpression:
		return each(n.Argument)
	case *ast.IsDefinedExpression:
		return each(n.Left)
	case *ast.IsEmptyExpression:
		return each(n.Left)
	case *ast.IsExpression:
		// Right is set when the right-hand side is a value (or a name the
		// index has not resolved yet); a shape or type operand leaves it nil.
		return each(n.Left, n.Right)
	case *ast.TrailingCommentExpression:
		return each(n.Wrap)
	case *ast.PrecedingCommentExpression:
		return each(n.Wrap)
	case *ast.ImportClause:
		// The `with` values are evaluated in the importing policy, so they can
		// reference its rules and lets.
		for _, w := range n.Withs {
			if w == nil {
				continue
			}
			if err := each(w); err != nil {
				return err
			}
		}
		return nil
	case *ast.WithClause:
		return each(n.Expr)
	case *ast.AttachmentClause:
		return each(n.As)
	case *ast.VarDeclaration:
		return each(n.Value)
	case *ast.RuleStatement:
		return each(n.Default, n.When, n.Body)
	default:
		return fmt.Errorf("reference walk: unsupported node %T", n)
	}
}
