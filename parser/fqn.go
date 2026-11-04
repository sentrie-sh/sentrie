// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"log/slog"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// parseFQN parses a slash-separated name (`a/b/c`) in a declaration or type
// position: a namespace declaration, the `from` of an imported decision, a
// shape's `with` base and a named type. The first segment must be an
// identifier, so a name in a type position can still be told apart from a type
// keyword (`string` vs `text/string`). Every later segment may be a keyword
// (`constraints/in`, `text/string`), since after `/` a keyword cannot mean
// anything else here (#297).
//
// Slash-qualified names in expression position (a derive callee
// `lib/in/double(2)`, the right of `is`) are division chains read by the
// expression parser, not by this production. There a keyword after `/` is also
// a segment (see parseInfixExpression), except the six keywords that start an
// expression (`not`, `true`, `false`, `unknown`, `null`, `transform`): `a/not b`
// and `a/true` stay division, so those six cannot be a later segment of a name
// used in an expression.
func parseFQN(ctx context.Context, p *Parser) *ast.FQN {
	slog.DebugContext(ctx, "PARSE_FQN", "current", p.current)
	defer slog.DebugContext(ctx, "PARSE_FQN_DONE", "current", p.current)

	var parts []string

	if kind, isKeyword := tokens.IsKeyword(p.head().Value); isKeyword && p.head().IsOfKind(kind) {
		p.errorf("expected %s, got keyword '%s': the first segment of a name cannot be a keyword", tokens.Ident, p.head().Value)
		return nil
	}
	firstIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	parts = append(parts, firstIdent.Value)
	rnge := firstIdent.Range

	for p.canExpect(tokens.TokenDiv) {
		p.advance() // consume the '/'

		segment, found := advanceNameOrKeyword(p, tokens.Ident.String())
		if !found {
			return nil
		}
		parts = append(parts, segment.Value)
		rnge.To = segment.Range.To
	}

	fqn := ast.NewFQN(parts, rnge)

	return &fqn
}
