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

func parseTernaryExpression(ctx context.Context, p *Parser, condition ast.Expression, precedence Precedence) ast.Expression {
	// Parse the '?' token
	if !p.expect(tokens.TokenQuestion) {
		return nil
	}

	// Parse the true branch
	trueBranch := p.parseExpression(ctx, precedence)

	// Parse the ':' token
	if !p.expect(tokens.PunctColon) {
		return nil
	}

	// Parse the false branch
	falseBranch := p.parseExpression(ctx, precedence)

	return &ast.TernaryExpression{
		Condition:  condition,
		ThenBranch: trueBranch,
		ElseBranch: falseBranch,
	}
}
