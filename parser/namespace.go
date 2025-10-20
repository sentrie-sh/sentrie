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

func parseNamespaceStatement(ctx context.Context, p *Parser) ast.Statement {
	slog.DebugContext(ctx, "PARSE_NS", "current", p.current)
	defer slog.DebugContext(ctx, "PARSE_NS_DONE", "current", p.current)

	head := p.head()
	if !p.expect(tokens.KeywordNamespace) {
		return nil // Error in parsing the namespace statement
	}
	name, nameRange := parseFQN(ctx, p)
	if len(name) == 0 {
		return nil
	}
	return &ast.NamespaceStatement{
		Range: tokens.Range{
			File: head.Range.File,
			From: tokens.Pos{
				Line:   head.Range.From.Line,
				Column: head.Range.From.Column,
				Offset: head.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   nameRange.To.Line,
				Column: nameRange.To.Column,
				Offset: nameRange.To.Offset,
			},
		},
		Name: name,
	}
}
