// SPDX-License-Identifier: Apache-2.0
//
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

func parseFieldAccessExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	operatorToken := p.advance()
	if !operatorToken.IsOfKind(tokens.TokenDot) {
		return nil // Error in parsing field access
	}

	fieldName, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	return ast.NewFieldAccessExpression(left, fieldName.Value, tokens.Range{
		File: operatorToken.Range.File,
		From: operatorToken.Range.From,
		To:   fieldName.Range.To,
	})
}

func parseIndexAccessExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	lbracket, found := p.advanceExpected(tokens.PunctLeftBracket)
	if !found {
		return nil // Error in parsing index access
	}

	index := p.parseExpression(ctx, LOWEST)
	if index == nil {
		return nil // Error in parsing index expression
	}

	rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
	if !found {
		return nil // Error in parsing index access
	}

	return ast.NewIndexAccessExpression(left, index, tokens.Range{
		File: rBracket.Range.File,
		From: lbracket.Range.From,
		To:   rBracket.Range.To,
	})
}
