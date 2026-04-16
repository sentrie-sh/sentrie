// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

	switch rhs := right.(type) {
	case *ast.Identifier:
		call := ast.NewCallExpression(rhs, []ast.Expression{left}, false, nil, pipelineRange)
		return applyPipelineMemoizationSuffix(ctx, p, call, pipelineRange)
	case *ast.FieldAccessExpression:
		if !hasIdentifierRoot(rhs) {
			break
		}
		call := ast.NewCallExpression(rhs, []ast.Expression{left}, false, nil, pipelineRange)
		return applyPipelineMemoizationSuffix(ctx, p, call, pipelineRange)
	case *ast.CallExpression:
		if hasIdentifierRoot(rhs.Callee) {
			args := rhs.Arguments
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
	}

	p.errorf(
		"invalid pipeline target: right-hand side must be an identifier, a module-qualified field access, or a call on one of those targets",
	)
	return nil
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

func applyPipelineMemoizationSuffix(ctx context.Context, p *Parser, call *ast.CallExpression, baseRange tokens.Range) ast.Expression {
	hadBang := p.head().IsOfKind(tokens.TokenBang)
	suffix := parseMemoizationSuffix(ctx, p)
	if suffix == nil {
		if hadBang {
			return nil
		}
		return call
	}
	rnge := baseRange
	rnge.To = suffix.To
	return ast.NewCallExpression(call.Callee, call.Arguments, true, suffix.TTL, rnge)
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
