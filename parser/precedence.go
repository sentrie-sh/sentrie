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

import "github.com/sentrie-sh/sentrie/tokens"

// Precedence levels for Pratt parser
type Precedence uint8

const (
	LOWEST     Precedence = iota
	TERNARY               // ? :
	OR                    // or
	XOR                   // xor
	AND                   // and
	EQUALITY              // == != is
	COMPARISON            // > < >= <= matches contains in
	SUM                   // + -
	PRODUCT               // * / %
	UNARY                 // !x -x +x not
	CALL                  // myFunction(X)
	INDEX                 // array[index], obj.field
	PRIMARY               // base precedence for primary expressions
)

var precedences = map[tokens.Kind]Precedence{
	tokens.TokenQuestion:        TERNARY,
	tokens.KeywordOr:            OR,
	tokens.KeywordXor:           XOR,
	tokens.KeywordAnd:           AND,
	tokens.TokenEq:              EQUALITY,
	tokens.TokenNeq:             EQUALITY,
	tokens.KeywordIs:            EQUALITY,
	tokens.TokenLt:              COMPARISON,
	tokens.TokenGt:              COMPARISON,
	tokens.TokenLte:             COMPARISON,
	tokens.TokenGte:             COMPARISON,
	tokens.KeywordIn:            COMPARISON,
	tokens.KeywordMatches:       COMPARISON,
	tokens.KeywordContains:      COMPARISON,
	tokens.KeywordNot:           UNARY,
	tokens.TokenBang:            UNARY,
	tokens.TokenPlus:            SUM,
	tokens.TokenMinus:           SUM,
	tokens.TokenDiv:             PRODUCT,
	tokens.TokenMul:             PRODUCT,
	tokens.TokenMod:             PRODUCT,
	tokens.PunctLeftParentheses: CALL,
	tokens.KeywordCast:          CALL,
	tokens.TokenDot:             INDEX,
	tokens.PunctLeftBracket:     INDEX,
}
