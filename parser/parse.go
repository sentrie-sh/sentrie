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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (p *Parser) ParseProgram(ctx context.Context) (*ast.Program, error) {
	prg := &ast.Program{
		Reference: p.reference,
	}

	// First non-comment statement must be namespace
	if !p.hasTokens() {
		// we can have empty files
		return nil, nil
	}

	prg.Statements = make([]ast.Statement, 0)

	// lets start parsing
	// Skip comment statements to find the first non-comment statement
	var firstStmt ast.Statement
	for p.hasTokens() {
		stmt := parseStatement(ctx, p)
		if p.err != nil {
			return nil, p.err
		}

		// Check if it's a comment statement
		if _, isComment := stmt.(*ast.CommentStatement); isComment {
			prg.Statements = append(prg.Statements, stmt)
			// consume trailing comments
			if p.canExpect(tokens.TrailingComment) {
				comment := p.advance()
				prg.Statements = append(prg.Statements, &ast.CommentStatement{
					Content: comment.Value,
					Range: tokens.Range{
						File: comment.Range.File,
						From: tokens.Pos{
							Line:   comment.Range.From.Line,
							Column: comment.Range.From.Column,
							Offset: comment.Range.From.Offset,
						},
						To: tokens.Pos{
							Line:   comment.Range.From.Line,
							Column: comment.Range.From.Column,
							Offset: comment.Range.From.Offset,
						},
					},
				})
			}
			continue
		}
		firstStmt = stmt
		break
	}

	if firstStmt == nil {
		// nothing to do here - a file with a bunch of comments is valid
		return prg, nil
	}

	// Check if first non-comment statement is namespace
	_, ok := firstStmt.(*ast.NamespaceStatement)
	if !ok {
		err := fmt.Errorf("program must start with namespace, got %T at %s", firstStmt, firstStmt.Span())
		p.err = err
		return nil, err
	}
	prg.Statements = append(prg.Statements, firstStmt)

	// consume the optional semicolon after namespace
	if p.canExpect(tokens.PunctSemicolon) {
		p.advance()
	}

	// consume trailing comments after namespace
	if p.canExpect(tokens.TrailingComment) {
		comment := p.advance()
		prg.Statements = append(prg.Statements, &ast.CommentStatement{
			Content: comment.Value,
			Range: tokens.Range{
				File: comment.Range.File,
				From: tokens.Pos{
					Line:   comment.Range.From.Line,
					Column: comment.Range.From.Column,
					Offset: comment.Range.From.Offset,
				},
				To: tokens.Pos{
					Line:   comment.Range.From.Line,
					Column: comment.Range.From.Column,
					Offset: comment.Range.From.Offset,
				},
			},
		})
	}

	// Parse remaining statements
	for p.hasTokens() {
		stmt := parseStatement(ctx, p)
		if p.err != nil {
			return nil, p.err
		}
		if stmt == nil {
			err := fmt.Errorf("failed to parse statement at line %d, column %d", p.current.Range.From.Line, p.current.Range.From.Column)
			p.err = err
			return nil, err
		}

		// this MUST not be a namespace statement
		_, ok := stmt.(*ast.NamespaceStatement)
		if ok {
			err := fmt.Errorf("namespace cannot be declared after namespace declaration at %s", stmt.Span())
			p.err = err
			return nil, err
		}

		prg.Statements = append(prg.Statements, stmt)

		if p.canExpect(tokens.TrailingComment) {
			comment := p.advance()
			prg.Statements = append(prg.Statements, &ast.CommentStatement{
				Content: comment.Value,
				Range: tokens.Range{
					File: comment.Range.File,
					From: tokens.Pos{
						Line:   comment.Range.From.Line,
						Column: comment.Range.From.Column,
						Offset: comment.Range.From.Offset,
					},
					To: tokens.Pos{
						Line:   comment.Range.From.Line,
						Column: comment.Range.From.Column,
						Offset: comment.Range.From.Offset,
					},
				},
			})
		}

		// consume the optional semicolon
		if p.canExpect(tokens.PunctSemicolon) {
			p.advance()
		}
	}

	return prg, nil
}
