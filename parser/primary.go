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
	"fmt"
	"strconv"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func parseNullLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	if token.Kind != tokens.KeywordNull {
		p.err = fmt.Errorf("expected `null` literal, got %s at line %d, column %d", token.Kind, token.Range.From.Line, token.Range.From.Column)
		return nil
	}
	return &ast.NullLiteral{Range: tokens.Range{
		File: token.Range.File,
		From: tokens.Pos{
			Line:   token.Range.From.Line,
			Column: token.Range.From.Column,
			Offset: token.Range.From.Offset,
		},
		To: tokens.Pos{
			Line:   token.Range.From.Line,
			Column: token.Range.From.Column,
			Offset: token.Range.From.Offset,
		},
	}}
}

func parseTrinaryLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	tristateValue := trinary.FromToken(token)
	return &ast.TrinaryLiteral{
		Range: tokens.Range{
			File: token.Range.File,
			From: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
		},
		Value: tristateValue,
	}
}

func parseIdentifier(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	return &ast.Identifier{
		Range: tokens.Range{
			File: token.Range.File,
			From: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
		},
		Value: token.Value,
	}
}

func parseIntegerLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	value, err := strconv.ParseInt(token.Value, 10, 64)
	if err != nil {
		p.errorf("invalid integer literal %q at %s: %w", token.Value, token.Range.From, err)
		return nil
	}
	return &ast.IntegerLiteral{
		Range: tokens.Range{
			File: token.Range.File,
			From: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
		},
		Value: value,
	}
}

func parseStringLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	return &ast.StringLiteral{
		Range: tokens.Range{
			File: token.Range.File,
			From: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
		},
		Value: token.Value,
	}
}

func parseFloatLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance()
	value, err := strconv.ParseFloat(token.Value, 64)
	if err != nil {
		p.errorf("invalid float literal %q at %s: %w", token.Value, token.Range.From, err)
		return nil
	}
	return &ast.FloatLiteral{
		Range: tokens.Range{
			File: token.Range.File,
			From: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   token.Range.From.Line,
				Column: token.Range.From.Column,
				Offset: token.Range.From.Offset,
			},
		},
		Value: value,
	}
}
