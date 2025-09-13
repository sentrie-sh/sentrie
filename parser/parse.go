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

	for p.hasTokens() {
		stmt := parseStatement(ctx, p)
		if p.err != nil {
			return nil, p.err
		}
		if stmt == nil {
			return nil, fmt.Errorf("failed to parse statement at %s", p.current.Position)
		}

		switch stmt := stmt.(type) {
		case *ast.NamespaceStatement:
			if prg.Namespace != nil {
				return nil, fmt.Errorf("multiple namespace statements at %s", stmt.Position())
			}
			prg.Namespace = stmt
		case *ast.PolicyStatement:
			prg.Policies = append(prg.Policies, stmt)
		case *ast.ShapeStatement:
			prg.Shapes = append(prg.Shapes, stmt)
		case *ast.ShapeExportStatement:
			prg.ShapeExports = append(prg.ShapeExports, stmt)
		}

		// consume the optional semicolon
		if p.canExpect(tokens.PunctSemicolon) {
			p.advance()
		}
	}

	if prg.Namespace == nil {
		return nil, errors.Wrapf(ErrParse, "no namespace in program at %s", p.reference)
	}

	return prg, nil
}
