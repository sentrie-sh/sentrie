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

func parseNotExpression(ctx context.Context, parser *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	notToken := parser.advance()

	opToken := parser.head()

	if !opToken.IsOfKind(tokens.KeywordNot, tokens.KeywordMatches, tokens.KeywordContains, tokens.KeywordIn) {
		parser.errorf("expected 'not', 'matches', 'contains', or 'in' after 'not', got %s", opToken.Kind)
		return nil
	}

	parser.advance()
	right := parser.parseExpression(ctx, precedence)
	if right == nil {
		return nil
	}

	// build the infix expression
	bin := &ast.InfixExpression{
		Range: tokens.Range{
			File: opToken.Range.File,
			From: tokens.Pos{
				Line:   opToken.Range.From.Line,
				Column: opToken.Range.From.Column,
				Offset: opToken.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   opToken.Range.From.Line,
				Column: opToken.Range.From.Column,
				Offset: opToken.Range.From.Offset,
			},
		},
		Left:     left,
		Operator: opToken.Value,
		Right:    right,
	}

	// wrap it in a unary expression
	return &ast.UnaryExpression{
		Range: tokens.Range{
			File: notToken.Range.File,
			From: tokens.Pos{
				Line:   notToken.Range.From.Line,
				Column: notToken.Range.From.Column,
				Offset: notToken.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   notToken.Range.From.Line,
				Column: notToken.Range.From.Column,
				Offset: notToken.Range.From.Offset,
			},
		},
		Operator: notToken.Value,
		Right:    bin,
	}
}
