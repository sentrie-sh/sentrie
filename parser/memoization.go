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
