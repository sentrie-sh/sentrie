// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import "github.com/sentrie-sh/sentrie/tokens"

// Precedence levels for Pratt parser
type Precedence uint8

const (
	LOWEST     Precedence = iota
	PIPELINE             // |>
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
	tokens.TokenPipeForward:     PIPELINE,
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
