// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

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
	var lastSegment tokens.Instance

	if p.canExpect(tokens.String) {
		fromModule, _ := p.advanceExpected(tokens.String)
		relativeFrom = fromModule.Value
		rnge.To = tokenEnd(fromModule)
	} else {
		if !p.expect(tokens.TokenAt) {
			return nil
		}
		fromPackage, found := advanceModulePathSegment(p)
		if !found {
			return nil
		}
		lastSegment = fromPackage
		rnge.To = tokenEnd(fromPackage)
		from := []string{fromPackage.Value}
		for p.canExpect(tokens.TokenDiv) {
			p.advance() // consume '/'
			fromModule, found := advanceModulePathSegment(p)
			if !found {
				return nil
			}
			lastSegment = fromModule
			rnge.To = tokenEnd(fromModule)
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
		rnge.To = tokenEnd(asAlias)
	} else if relativeFrom == "" {
		// A keyword segment such as `string` in `@sentrie/string` is a valid
		// path segment but not a valid identifier, so it cannot be the alias
		// that call sites use. Ask for an explicit one here rather than let
		// `string.lower(x)` fail later with an unrelated expression error.
		if _, isKeyword := tokens.IsKeyword(alias); isKeyword {
			// Reported at the segment itself, not at the token after the statement.
			p.err = errors.Join(p.err, fmt.Errorf("parsing error at %s: module @%s: default alias %q is a keyword; add 'as <alias>'",
				lastSegment.Range.String(), strings.Join(libFrom, "/"), alias))
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

// tokenEnd returns the position just past t on t's own line, the end a
// statement span should carry when t is its last token.
//
// The lexer's end position is the position of the rune after the token, which
// is right when that rune is on the same line but not otherwise: before a
// newline it is reported as column 0 of the next line (#216), and at EOF it is
// one column short. Offsets are right in every case, so the end is rebuilt
// from the token's start and its source width. The width in runes is the byte
// width less the extra bytes of multi-byte runes, which a string's escapes
// carry over unchanged from source to value. A value containing a newline (a
// heredoc) is not on one line, so it keeps the lexer's end.
func tokenEnd(t tokens.Instance) tokens.Pos {
	width := t.Range.To.Offset - t.Range.From.Offset
	if width < 0 || strings.ContainsRune(t.Value, '\n') {
		return t.Range.To
	}
	runes := width - (len(t.Value) - utf8.RuneCountInString(t.Value))
	return tokens.Pos{
		Line:   t.Range.From.Line,
		Column: t.Range.From.Column + runes,
		Offset: t.Range.To.Offset,
	}
}
