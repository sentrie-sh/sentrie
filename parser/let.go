// SPDX-License-Identifier: Apache-2.0

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

func parseLetsStatement(ctx context.Context, p *Parser) ast.Statement {
	start := p.head()
	rnge := start.Range

	p.advance() // consume 'let'

	nameIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	if nameIdent.Value == "tri" {
		str := ""
		_ = str
	}

	name := nameIdent.Value
	rnge.To = nameIdent.Range.To

	var typeRef ast.TypeRef
	if p.canExpect(tokens.PunctColon) {
		colon, found := p.advanceExpected(tokens.PunctColon)
		if !found {
			return nil
		}
		typeRef = parseTypeRef(ctx, p)
		if typeRef == nil {
			return nil
		}
		rnge.To = colon.Range.To
	}

	if !p.expect(tokens.TokenAssign) { // expect '='
		return nil
	}

	val := p.parseExpression(ctx, LOWEST)
	if val == nil {
		return nil
	}
	rnge.To = val.Span().To

	return ast.NewVarDeclaration(name, typeRef, val, rnge)
}
