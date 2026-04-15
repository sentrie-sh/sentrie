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
			args := make([]ast.Expression, 0, len(rhs.Arguments)+1)
			args = append(args, left)
			args = append(args, rhs.Arguments...)
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
