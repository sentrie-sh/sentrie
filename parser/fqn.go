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
	"log/slog"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseFQN(ctx context.Context, p *Parser) *ast.FQN {
	slog.DebugContext(ctx, "PARSE_FQN", "current", p.current)
	defer slog.DebugContext(ctx, "PARSE_FQN_DONE", "current", p.current)

	var parts []string

	// consume the first firstIdent
	firstIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	parts = append(parts, firstIdent.Value)
	rnge := firstIdent.Range

	for p.canExpect(tokens.TokenDiv) {
		p.advance() // consume the '/'

		ident, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		parts = append(parts, ident.Value)
		rnge.To = ident.Range.To
	}

	fqn := ast.NewFQN(parts, rnge)

	return &fqn
}
