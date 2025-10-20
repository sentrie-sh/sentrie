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

func parseShapeExportStatement(ctx context.Context, p *Parser) ast.Statement {
	start := p.head()

	p.advance() // consume 'export'

	if !p.expect(tokens.KeywordShape) {
		return nil
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	return &ast.ShapeExportStatement{
		Name: name.Value,
		Range: tokens.Range{
			File: start.Range.File,
			From: tokens.Pos{
				Line:   start.Range.From.Line,
				Column: start.Range.From.Column,
				Offset: start.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   name.Range.From.Line,
				Column: name.Range.From.Column,
				Offset: name.Range.From.Offset,
			},
		},
	}
}
