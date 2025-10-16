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

func parseFromLeftCurly(ctx context.Context, p *Parser) ast.Expression {
	// depending on what follows a left curly, we switch between parsing a map literal or a block expression
	if p.peek().IsOfKind(tokens.String) || p.peek().IsOfKind(tokens.PunctLeftBracket) || p.peek().IsOfKind(tokens.PunctRightCurly) {
		return parseMapLiteral(ctx, p)
	}
	return parseBlockExpression(ctx, p)
}
