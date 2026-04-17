// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

type memoizationSuffix struct {
	TTL *time.Duration
	To  tokens.Pos
}

// parseMemoizationSuffix parses a trailing memoization suffix.
//
// Returns nil when no suffix is present.
func parseMemoizationSuffix(ctx context.Context, p *Parser) *memoizationSuffix {
	if !p.head().IsOfKind(tokens.TokenBang) {
		return nil
	}
	bang, found := p.advanceExpected(tokens.TokenBang)
	if !found {
		return nil
	}
	suffix := &memoizationSuffix{
		TTL: nil,
		To:  bang.Range.To,
	}
	if p.head().IsOfKind(tokens.Int) {
		literal := parseIntegerLiteral(ctx, p)
		if literal == nil {
			return nil
		}
		ttl := time.Duration(literal.(*ast.IntegerLiteral).Value) * time.Second
		suffix.TTL = &ttl
		suffix.To = literal.Span().To
	}
	return suffix
}
