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
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'use' '{' func (',' func)* '}' 'from' moduleName 'as' alias
func parseUseStatement(ctx context.Context, p *Parser) ast.Statement {
	head, found := p.advanceExpected(tokens.KeywordUse)
	if !found {
		return nil
	}
	rnge := head.Range

	fns := []string{}

	if !p.expect(tokens.PunctLeftCurly) {
		return nil
	}

	firstModuleName, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	fns = append(fns, firstModuleName.Value)

	for !p.head().IsOfKind(tokens.PunctRightCurly) {
		if !p.expect(tokens.PunctComma) {
			return nil
		}
		fn, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		fns = append(fns, fn.Value)
	}

	rightCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}
	rnge.To = rightCurly.Range.To

	modules := fns

	if !p.expect(tokens.KeywordFrom) {
		return nil
	}

	if !p.canExpectAnyOf(tokens.String, tokens.TokenAt) {
		p.errorf("expected string or '@' for module import")
		return nil
	}

	relativeFrom := ""
	libFrom := []string{}

	if p.canExpect(tokens.String) {
		fromModule, _ := p.advanceExpected(tokens.String)
		relativeFrom = fromModule.Value
		rnge.To = fromModule.Range.To
	} else {
		at, found := p.advanceExpected(tokens.TokenAt)
		if !found {
			return nil
		}
		rnge.To = at.Range.To

		fromPackage, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		from := []string{fromPackage.Value}
		for p.canExpect(tokens.TokenDiv) {
			p.advance() // consume '/'
			fromModule, found := p.advanceExpected(tokens.Ident)
			if !found {
				return nil
			}
			rnge.To = fromModule.Range.To
			from = append(from, fromModule.Value)
		}
		libFrom = from
	}

	// default alias to the module name
	// for @foo/bar, default alias is bar - the last part of the path
	// for quoted strings, it's the last part of the path
	alias := ""
	if relativeFrom != "" {
		parts := strings.Split(relativeFrom, "/")
		alias = parts[len(parts)-1]
	} else {
		alias = libFrom[len(libFrom)-1]
	}

	if p.canExpect(tokens.KeywordAs) {
		p.advance() // consume 'as'

		asAlias, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		alias = asAlias.Value
		rnge.To = asAlias.Range.To
	}

	return ast.NewUseStatement(modules, relativeFrom, libFrom, alias, rnge)
}
