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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'transform' <expression> 'with' @string // where the string is a JQ compatible transformer
func parseTransformExpression(ctx context.Context, p *Parser) ast.Expression {
	transformToken, found := p.advanceExpected(tokens.KeywordTransform)
	if !found {
		return nil
	}
	rnge := transformToken.Range

	// Parse the expression that follows the transform keyword
	argument := p.parseExpression(ctx, LOWEST)
	if argument == nil {
		return nil // Error in parsing the expression
	}
	rnge.To = argument.Span().To

	if !p.expect(tokens.KeywordWith) {
		return nil
	}

	transformer, found := p.advanceExpected(tokens.String)
	if !found {
		return nil
	}
	rnge.To = transformer.Range.To

	return ast.NewTransformExpression(argument, transformer.Value, rnge)
}
