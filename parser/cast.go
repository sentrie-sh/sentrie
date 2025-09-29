// Copyright 2025 Binaek Sarkar
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

// CAST expr as <type>
func parseCastExpression(ctx context.Context, p *Parser) ast.Expression {
	start := p.head()

	// consume the 'cast' token
	if !p.expect(tokens.KeywordCast) {
		return nil
	}

	// parse the type to cast to
	what := p.parseExpression(ctx, LOWEST)
	if what == nil {
		return nil
	}

	if !p.expect(tokens.KeywordAs) {
		return nil
	}

	// for now we allow allow casting to primitive types - not sure why we would need anything else
	if !p.canExpectAnyOf(PRIMITIVE_TYPES...) {
		p.errorf("expected primitive type after 'as' in 'cast', got %s", p.head().Kind)
		return nil
	}

	// parse the expression to cast to
	typeRef := parseTypeRef(ctx, p)
	if typeRef == nil {
		p.errorf("expected expression after type in cast, got %s", p.head().Kind)
		return nil
	}

	return &ast.CastExpression{
		Pos:        start.Position,
		Expr:       what,
		TargetType: typeRef,
	}
}
