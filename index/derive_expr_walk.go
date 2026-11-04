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

// forEachDeriveExprChild yields the direct children of e for derive analysis.
// The shape of the AST is owned by forEachExprChild; this function delegates to
// it and keeps only what derive analysis needs on top:
//
//   - A derive body is a pure expression, so the kinds that only make sense in
//     a rule body (import and attachment clauses) are rejected rather than
//     descended into. This keeps derive purity failing closed.
//   - A block may hold only `let` declarations and line comments before its
//     yield. The let's value is yielded directly (the VarDeclaration node
//     itself is not an expression), a comment is skipped, and any other
//     statement is rejected, matching the rule walkDeriveBlock enforces.
//   - The errors name the derive walk, not the reference walk.
//
// A new expression kind needs a case in forEachExprChild only;
// TestForEachExprChildCoversEveryASTNode and
// TestForEachDeriveExprChildCoversEveryASTNode check this.
func forEachDeriveExprChild(e ast.Expression, yield func(ast.Expression) error) error {
	switch e.(type) {
	case *ast.ImportClause, *ast.WithClause, *ast.AttachmentClause:
		return fmt.Errorf("derive walk: unsupported expression %T", e)
	}
	var childErr error
	err := forEachExprChild(e, func(child ast.Node) error {
		switch c := child.(type) {
		case *ast.VarDeclaration:
			childErr = yield(c.Value)
		case *ast.CommentStatement:
			childErr = nil
		case ast.Expression:
			childErr = yield(c)
		default:
			childErr = fmt.Errorf("derive walk: unsupported statement %T", child)
		}
		return childErr
	})
	if childErr != nil {
		return childErr
	}
	if err != nil {
		// forEachExprChild reports only the kind of e itself when it has no case.
		return fmt.Errorf("derive walk: unsupported expression %T", e)
	}
	return nil
}
