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
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sentrie-sh/sentrie/lexer"
	"github.com/sentrie-sh/sentrie/tokens"
)

type Parser struct {
	lexer     *lexer.Lexer
	reference string
	current   tokens.Instance
	next      tokens.Instance

	atEof bool // Indicates if the parser has reached the end of the file

	err error

	// Pratt parser function maps
	prefixHandlers          map[tokens.Kind]prefixParser
	infixHandlers           map[tokens.Kind]infixParser
	statementHandlers       map[tokens.Kind]statementParser
	policyStatementHandlers map[tokens.Kind]statementParser
}

// NewParser creates a new parser
func NewParser(input io.Reader, filename string) *Parser {
	lexer := lexer.NewLexer(input, filename)
	parser := &Parser{
		lexer:     lexer,
		reference: filename,
	}

	parser.registerParseFns()

	parser.advance()
	parser.advance()
	return parser
}

// NewParserFromString creates a new parser from a string (convenience function)
func NewParserFromString(input, filename string) *Parser {
	return NewParser(strings.NewReader(input), filename)
}

func (p *Parser) head() tokens.Instance {
	return p.current
}

// advance moves the HEAD to the next token
func (p *Parser) advance() tokens.Instance {
	if p.atEof {
		return tokens.Err(p.current.Range, "cannot advance, already at EOF")
	}
	if p.current.IsOfKind(tokens.Error) {
		p.errorf(p.current.Value)
		return p.current
	}
	current := p.current
	p.current = p.next
	if p.current.Kind == tokens.EOF {
		p.atEof = true
		return current
	}
	p.next = p.lexer.NextToken()
	return current
}

func (p *Parser) advanceExpected(kind tokens.Kind) (tokens.Instance, bool) {
	token := p.current
	if !token.IsOfKind(kind) {
		p.errorf("expected %s, got %s at %s", kind, p.current.Kind, p.current.Range)
		return tokens.Err(p.current.Range, fmt.Sprintf("expected %s, got %s", kind, p.current.Kind)), false
	}
	return p.advance(), true
}

func (p *Parser) expect(kind tokens.Kind) bool {
	if p.current.Kind != kind {
		p.errorf("expected '%s', got %s at %s", kind, p.current.Kind, p.current.Range)
		return false
	}
	_ = p.advance()
	return true
}

func (p *Parser) canExpect(kind tokens.Kind) bool {
	return p.current.Kind == kind
}

func (p *Parser) canExpectAnyOf(kinds ...tokens.Kind) bool {
	for _, kind := range kinds {
		if p.current.Kind == kind {
			return true
		}
	}
	return false
}

func (p *Parser) hasTokens() bool {
	return !p.atEof
}

func (p *Parser) peek() tokens.Instance {
	if p.atEof {
		return tokens.Instance{Kind: tokens.EOF} // Return EOF if at end of file
	}
	return p.next
}

// errorf adds a formatted error
func (p *Parser) errorf(format string, args ...interface{}) {
	format = "parsing error at %s: " + format
	args = append([]any{p.current.Range.String()}, args...)

	p.err = errors.Join(
		p.err,
		fmt.Errorf(format, args...),
	)
}

func (p *Parser) registerPrefix(tokenType tokens.Kind, fn prefixParser) {
	p.prefixHandlers[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType tokens.Kind, fn infixParser) {
	p.infixHandlers[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t tokens.Kind) {
	p.errorf("no prefix parse function found for '%s'", t)
}
