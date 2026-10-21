// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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

		fromPackage, found := advanceModulePathSegment(p)
		if !found {
			return nil
		}
		from := []string{fromPackage.Value}
		for p.canExpect(tokens.TokenDiv) {
			p.advance() // consume '/'
			fromModule, found := advanceModulePathSegment(p)
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
	} else if relativeFrom == "" {
		// A keyword segment such as `string` in `@sentrie/string` is a valid
		// path segment but not a valid identifier, so it cannot be the alias
		// that call sites use. Ask for an explicit one here rather than let
		// `string.lower(x)` fail later with an unrelated expression error.
		if _, isKeyword := tokens.IsKeyword(alias); isKeyword {
			p.errorf("module @%s: default alias %q is a keyword; add 'as <alias>'", strings.Join(libFrom, "/"), alias)
			return nil
		}
	}

	return ast.NewUseStatement(modules, relativeFrom, libFrom, alias, rnge)
}

// advanceModulePathSegment consumes one segment of an `@scope/name` module
// path. A segment is an identifier or a keyword, so modules can be named after
// keywords (`@sentrie/string`). Keywords are accepted only here: everywhere else
// they keep their keyword meaning.
func advanceModulePathSegment(p *Parser) (tokens.Instance, bool) {
	if kind, isKeyword := tokens.IsKeyword(p.head().Value); isKeyword && p.head().IsOfKind(kind) {
		return p.advance(), true
	}
	return p.advanceExpected(tokens.Ident)
}
