// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parsePipelineExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	operator, found := p.advanceExpected(tokens.TokenPipeForward)
	if !found {
		return nil
	}
	if left == nil {
		p.errorf("invalid pipeline target: missing left-hand side expression")
		return nil
	}

	startedAsGrouped := p.head().IsOfKind(tokens.PunctLeftParentheses)
	right := p.parseExpression(ctx, precedence)
	if right == nil {
		return nil
	}
	if startedAsGrouped {
		p.errorf("invalid pipeline target: grouped expressions are not allowed on the right-hand side")
		return nil
	}

	pipelineRange := tokens.Range{
		File: operator.Range.File,
		From: left.Span().From,
		To:   right.Span().To,
	}

	const invalidPipelineRHS = "invalid pipeline target: right-hand side must be a call on an identifier or module-qualified field access"

	rhs, ok := right.(*ast.CallExpression)
	if !ok {
		p.errorf(invalidPipelineRHS)
		return nil
	}
	if !hasIdentifierRoot(rhs.Callee) {
		p.errorf(invalidPipelineRHS)
		return nil
	}

	var args []ast.Expression
	if containsPipelineHoleInExprs(rhs.Arguments) {
		args = make([]ast.Expression, len(rhs.Arguments))
		for i := range rhs.Arguments {
			args[i] = substitutePipelineHoles(rhs.Arguments[i], left)
		}
	} else {
		args = make([]ast.Expression, 0, len(rhs.Arguments)+1)
		args = append(args, left)
		args = append(args, rhs.Arguments...)
	}
	return ast.NewCallExpression(rhs.Callee, args, rhs.Memoized, rhs.MemoizeTTL, pipelineRange)
}

func hasIdentifierRoot(expr ast.Expression) bool {
	switch t := expr.(type) {
	case *ast.Identifier:
		return true
	case *ast.FieldAccessExpression:
		return hasIdentifierRoot(t.Left)
	default:
		return false
	}
}

func containsPipelineHoleInExprs(exprs []ast.Expression) bool {
	for i := range exprs {
		if containsPipelineHole(exprs[i]) {
			return true
		}
	}
	return false
}

func containsPipelineHole(expr ast.Expression) bool {
	switch t := expr.(type) {
	case *ast.PipelineHoleExpression:
		return true
	case *ast.CallExpression:
		if containsPipelineHole(t.Callee) {
			return true
		}
		return containsPipelineHoleInExprs(t.Arguments)
	case *ast.FieldAccessExpression:
		return containsPipelineHole(t.Left)
	case *ast.IndexAccessExpression:
		return containsPipelineHole(t.Left) || containsPipelineHole(t.Index)
	case *ast.ListLiteral:
		return containsPipelineHoleInExprs(t.Values)
	case *ast.MapLiteral:
		for i := range t.Entries {
			if containsPipelineHole(t.Entries[i].Key) || containsPipelineHole(t.Entries[i].Value) {
				return true
			}
		}
	case *ast.InfixExpression:
		return containsPipelineHole(t.Left) || containsPipelineHole(t.Right)
	case *ast.UnaryExpression:
		return containsPipelineHole(t.Right)
	case *ast.TernaryExpression:
		return containsPipelineHole(t.Condition) || containsPipelineHole(t.ThenBranch) || containsPipelineHole(t.ElseBranch)
	case *ast.CastExpression:
		return containsPipelineHole(t.Expr)
	case *ast.IsDefinedExpression:
		return containsPipelineHole(t.Left)
	case *ast.IsEmptyExpression:
		return containsPipelineHole(t.Left)
	case *ast.TransformExpression:
		return containsPipelineHole(t.Argument)
	case *ast.PrecedingCommentExpression:
		return containsPipelineHole(t.Wrap)
	case *ast.TrailingCommentExpression:
		return containsPipelineHole(t.Wrap)
	}
	return false
}

func substitutePipelineHoles(expr ast.Expression, replacement ast.Expression) ast.Expression {
	switch t := expr.(type) {
	case *ast.PipelineHoleExpression:
		return replacement
	case *ast.CallExpression:
		args := make([]ast.Expression, len(t.Arguments))
		for i := range t.Arguments {
			args[i] = substitutePipelineHoles(t.Arguments[i], replacement)
		}
		return ast.NewCallExpression(substitutePipelineHoles(t.Callee, replacement), args, t.Memoized, t.MemoizeTTL, t.Span())
	case *ast.FieldAccessExpression:
		return ast.NewFieldAccessExpression(substitutePipelineHoles(t.Left, replacement), t.Field, t.Span())
	case *ast.IndexAccessExpression:
		return ast.NewIndexAccessExpression(
			substitutePipelineHoles(t.Left, replacement),
			substitutePipelineHoles(t.Index, replacement),
			t.Span(),
		)
	case *ast.ListLiteral:
		values := make([]ast.Expression, len(t.Values))
		for i := range t.Values {
			values[i] = substitutePipelineHoles(t.Values[i], replacement)
		}
		return ast.NewListLiteral(values, t.Span())
	case *ast.MapLiteral:
		entries := make([]ast.MapEntry, len(t.Entries))
		for i := range t.Entries {
			entries[i] = ast.MapEntry{
				Key:   substitutePipelineHoles(t.Entries[i].Key, replacement),
				Value: substitutePipelineHoles(t.Entries[i].Value, replacement),
			}
		}
		return ast.NewMapLiteral(entries, t.Span())
	case *ast.InfixExpression:
		return ast.NewInfixExpression(
			substitutePipelineHoles(t.Left, replacement),
			substitutePipelineHoles(t.Right, replacement),
			t.Operator,
			t.Span(),
		)
	case *ast.UnaryExpression:
		return ast.NewUnaryExpression(t.Operator, substitutePipelineHoles(t.Right, replacement), t.Span())
	case *ast.TernaryExpression:
		return ast.NewTernaryExpression(
			substitutePipelineHoles(t.Condition, replacement),
			substitutePipelineHoles(t.ThenBranch, replacement),
			substitutePipelineHoles(t.ElseBranch, replacement),
			t.Span(),
		)
	case *ast.CastExpression:
		return ast.NewCastExpression(substitutePipelineHoles(t.Expr, replacement), t.TargetType, t.Span())
	case *ast.IsDefinedExpression:
		return ast.NewIsDefinedExpression(substitutePipelineHoles(t.Left, replacement), t.Span())
	case *ast.IsEmptyExpression:
		return ast.NewIsEmptyExpression(substitutePipelineHoles(t.Left, replacement), t.Span())
	case *ast.TransformExpression:
		return ast.NewTransformExpression(substitutePipelineHoles(t.Argument, replacement), t.Transformer, t.Span())
	case *ast.PrecedingCommentExpression:
		return ast.NewPrecedingCommentExpression(t.CommentContent, substitutePipelineHoles(t.Wrap, replacement), t.Span())
	case *ast.TrailingCommentExpression:
		return ast.NewTrailingCommentExpression(t.CommentContent, substitutePipelineHoles(t.Wrap, replacement), t.Span())
	default:
		return expr
	}
}
