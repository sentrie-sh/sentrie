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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/tokens"
	"github.com/pkg/errors"
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
					Pos:     comment.Position,
				})
			}
			continue
		}
		firstStmt = stmt
		break
	}

	if firstStmt == nil {
		err := errors.Wrapf(ErrParse, "no namespace in program at %s", p.reference)
		p.err = err
		return nil, err
	}

	// Check if first non-comment statement is namespace
	_, ok := firstStmt.(*ast.NamespaceStatement)
	if !ok {
		err := fmt.Errorf("program must start with namespace, got %T at %s", firstStmt, firstStmt.Position())
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
			Pos:     comment.Position,
		})
	}

	// Parse remaining statements
	for p.hasTokens() {
		stmt := parseStatement(ctx, p)
		if p.err != nil {
			return nil, p.err
		}
		if stmt == nil {
			err := fmt.Errorf("failed to parse statement at %s", p.current.Position)
			p.err = err
			return nil, err
		}

		// this MUST not be a namespace statement
		_, ok := stmt.(*ast.NamespaceStatement)
		if ok {
			err := fmt.Errorf("namespace cannot be declared after namespace declaration at %s", stmt.Position())
			p.err = err
			return nil, err
		}

		prg.Statements = append(prg.Statements, stmt)

		if p.canExpect(tokens.TrailingComment) {
			comment := p.advance()
			prg.Statements = append(prg.Statements, &ast.CommentStatement{
				Content: comment.Value,
				Pos:     comment.Position,
			})
		}

		// consume the optional semicolon
		if p.canExpect(tokens.PunctSemicolon) {
			p.advance()
		}
	}

	return prg, nil
}
