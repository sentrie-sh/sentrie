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

package lexer

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/tokens"
)

type Lexer struct {
	reader   *bufio.Reader
	filename string

	line   int
	column int

	current     rune
	currentLine []rune // buffer for lookback

	offset       int
	currentWidth int
	atEOF        bool

	identRegex *regexp.Regexp
}

func NewLexer(reader io.Reader, filename string) *Lexer {
	l := &Lexer{
		reader:      bufio.NewReader(reader),
		filename:    filename,
		currentLine: []rune{},
		identRegex:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
	}
	l.readRune() // Initialize the first rune
	return l
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() tokens.Instance {
	for {
		l.skipWhitespace()

		if l.current == 0 {
			return tokens.New(tokens.EOF, "", l.currentPosition())
		}

		pos := l.currentPosition()

		switch l.current {
		case '-':
			if l.peekAhead() == '-' {
				l.readRune() // consume second '-'
				commentKind, value := l.readComment()
				return tokens.New(commentKind, value, pos)
			}
			l.readRune()
			return tokens.New(tokens.TokenMinus, "-", pos)

		case '=':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				return tokens.New(tokens.TokenEq, "==", pos)
			}
			l.readRune()
			return tokens.New(tokens.TokenAssign, "=", pos)

		case '!':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				return tokens.New(tokens.TokenNeq, "!=", pos)
			}
			l.readRune()
			return tokens.New(tokens.TokenBang, "!", pos)

		case '<':
			if l.peekString(2) == "<<" {
				value, err := l.readHereDoc()
				if err != nil {
					return tokens.New(tokens.Error, err.Error(), pos)
				}
				return tokens.New(tokens.String, value, pos)
			}
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				return tokens.New(tokens.TokenLte, "<=", pos)
			}
			l.readRune()
			return tokens.New(tokens.TokenLt, "<", pos)

		case '>':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				return tokens.New(tokens.TokenGte, ">=", pos)
			}
			l.readRune()
			return tokens.New(tokens.TokenGt, ">", pos)

		case '+':
			l.readRune()
			return tokens.New(tokens.TokenPlus, "+", pos)
		case '*':
			l.readRune()
			return tokens.New(tokens.TokenMul, "*", pos)
		case '/':
			l.readRune()
			return tokens.New(tokens.TokenDiv, "/", pos)
		case '%':
			l.readRune()
			return tokens.New(tokens.TokenMod, "%", pos)
		case '?':
			l.readRune()
			return tokens.New(tokens.TokenQuestion, "?", pos)
		case ':':
			l.readRune()
			return tokens.New(tokens.PunctColon, ":", pos)
		case '.':
			l.readRune()
			return tokens.New(tokens.TokenDot, ".", pos)
		case '@':
			l.readRune()
			return tokens.New(tokens.TokenAt, "@", pos)
		case ',':
			l.readRune()
			return tokens.New(tokens.PunctComma, ",", pos)
		case ';':
			l.readRune()
			return tokens.New(tokens.PunctSemicolon, ";", pos)
		case '(':
			l.readRune()
			return tokens.New(tokens.PunctLeftParentheses, "(", pos)
		case ')':
			l.readRune()
			return tokens.New(tokens.PunctRightParentheses, ")", pos)
		case '{':
			l.readRune()
			return tokens.New(tokens.PunctLeftCurly, "{", pos)
		case '}':
			l.readRune()
			return tokens.New(tokens.PunctRightCurly, "}", pos)
		case '[':
			l.readRune()
			return tokens.New(tokens.PunctLeftBracket, "[", pos)
		case ']':
			l.readRune()
			return tokens.New(tokens.PunctRightBracket, "]", pos)

		case '"':
			value, err := l.readString()
			if err != nil {
				return tokens.New(tokens.Error, err.Error(), pos)
			}
			return tokens.New(tokens.String, value, pos)

		default:
			if unicode.IsLetter(l.current) || l.current == '_' {
				value := l.readIdentifier()
				if !l.identRegex.MatchString(value) {
					return tokens.Err(pos, "invalid identifier: "+value)
				}

				// is this a known keyword?
				if kind, isKeyword := tokens.IsKeyword(value); isKeyword {
					return tokens.New(kind, value, pos)
				}
				return tokens.New(tokens.Ident, value, pos)
			}

			if unicode.IsDigit(l.current) {
				value, kind := l.readNumber()
				return tokens.New(kind, value, pos)
			}

			// Unknown character
			char := string(l.current)
			l.readRune()
			return tokens.New(tokens.Error, "unexpected character: "+char, pos)
		}
	}
}

// readRune reads the next rune from input
func (l *Lexer) readRune() {
	if l.atEOF {
		l.current = 0
		l.currentWidth = 0
		return
	}

	r, size, err := l.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.atEOF = true
			l.current = 0
			l.currentWidth = 0
			return
		}
		// For other errors, we could set current to an error rune
		// For now, treat as EOF
		l.atEOF = true
		l.current = 0
		l.currentWidth = 0
		return
	}

	l.current = r
	l.currentWidth = size
	l.offset += size

	// Update the current line buffer
	l.currentLine = append(l.currentLine, r)

	if r == '\n' {
		l.line++
		l.currentLine = []rune{}
		l.column = 1
	} else {
		l.column++
	}
}

// peekAhead returns the next rune without advancing position
func (l *Lexer) peekAhead() rune {
	if l.atEOF {
		return 0
	}

	// Peek at least 4 bytes to handle any UTF-8 character
	bytes, err := l.reader.Peek(4)
	if err != nil && err != io.EOF {
		return 0
	}

	if len(bytes) == 0 {
		return 0
	}

	r, _ := utf8.DecodeRune(bytes)
	return r
}

// currentPosition returns the current position
func (l *Lexer) currentPosition() tokens.Position {
	return tokens.Position{
		Filename: l.filename,
		Offset:   l.offset - l.currentWidth,
		Line:     l.line,
		Column:   l.column - 1,
	}
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.current) {
		l.readRune()
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	var result strings.Builder

	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		result.WriteRune(l.current)
		l.readRune()
	}

	return result.String()
}

// readNumber reads an integer or float
func (l *Lexer) readNumber() (string, tokens.Kind) {
	result := bytes.NewBufferString("")
	kind := tokens.Int

	// consume digits
	for unicode.IsDigit(l.current) {
		result.WriteRune(l.current)
		l.readRune()
	}

	if l.current == '.' && unicode.IsDigit(l.peekAhead()) {
		kind = tokens.Float
		result.WriteRune(l.current)
		l.readRune() // consume '.'
		// consume the rest of the digits
		for unicode.IsDigit(l.current) {
			result.WriteRune(l.current)
			l.readRune()
		}
	}

	return result.String(), kind
}

// readComment reads a line comment starting with --
func (l *Lexer) readComment() (tokens.Kind, string) {
	kind := tokens.LineComment
	result := bytes.NewBufferString("")

	// if the line buffer has anything but \s*\-\-, then this is a trailing comment
	lineUptoDashes := l.currentLine[:len(l.currentLine)-2]

	idxOfNotWhitespace := slices.IndexFunc(lineUptoDashes, func(r rune) bool {
		return !unicode.IsSpace(r)
	})

	if idxOfNotWhitespace != -1 {
		// the presence of non-whitespace characters before the dashes indicates a trailing comment
		kind = tokens.TrailingComment
	}

	l.readRune() // consume second '-'

	for l.current != '\n' && l.current != 0 {
		result.WriteRune(l.current)
		l.readRune()
	}

	return kind, strings.TrimSpace(result.String())
}

// readString reads a quoted string literal
func (l *Lexer) readString() (string, error) {
	l.readRune() // skip opening quote

	var result strings.Builder
	for l.current != '"' && l.current != 0 {
		if l.current == '\\' {
			l.readRune()
			switch l.current {
			case '"', '\\', '/':
				result.WriteRune(l.current)
			case 'n':
				result.WriteRune('\n')
			case 't':
				result.WriteRune('\t')
			case 'r':
				result.WriteRune('\r')
			case 'b':
				result.WriteRune('\b')
			case 'f':
				result.WriteRune('\f')
			default:
				result.WriteRune(l.current)
			}
		} else {
			result.WriteRune(l.current)
		}
		l.readRune()
	}

	if l.current != '"' {
		return "", UnterminatedStringError(l.currentPosition())
	}
	l.readRune() // skip closing quote

	return result.String(), nil
}

// peekString peeks the next n bytes (ASCII use only, does not advance).
func (l *Lexer) peekString(n int) string {
	if l.atEOF || n <= 0 {
		return ""
	}
	b, err := l.reader.Peek(n)
	if err != nil && err != io.EOF {
		return ""
	}
	return string(b)
}

// readHereDoc reads a heredoc starting at the first '<' of '<<<'.
func (l *Lexer) readHereDoc() (string, error) {
	// We are currently on the first '<'. Consume the 3 '<'.
	l.readRune() // 1st '<'
	if l.current != '<' {
		return "", io.ErrUnexpectedEOF
	}
	l.readRune() // 2nd '<'
	if l.current != '<' {
		return "", io.ErrUnexpectedEOF
	}
	l.readRune() // 3rd '<'

	// Disallow spaces before tag to keep syntax tight.
	// Require TAG immediately.
	if !(unicode.IsLetter(l.current) || l.current == '_') {
		return "", errors.Wrap(InvalidHereDocSyntaxError(l.currentPosition()), "heredoc requires identifier tag after <<<")
	}

	// Read TAG (identifier)
	var tagBuilder strings.Builder
	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		tagBuilder.WriteRune(l.current)
		l.readRune()
	}
	tag := tagBuilder.String()
	if tag == "" || !l.identRegex.MatchString(tag) {
		return "", errors.Wrap(InvalidHereDocSyntaxError(l.currentPosition()), "invalid heredoc tag")
	}

	// Read to end of the tag line
	for l.current != '\n' && l.current != 0 {
		// No trailing junk allowed (only whitespace)
		if !unicode.IsSpace(l.current) && l.current != '\r' {
			return "", errors.Wrapf(InvalidHereDocSyntaxError(l.currentPosition()), "unexpected characters after heredoc tag %q", tag)
		}
		l.readRune()
	}
	// Consume the '\n' if present
	if l.current == '\n' {
		l.readRune()
	}

	// Now collect lines until a line that is exactly == tag
	var sb strings.Builder
	for {
		// Capture the current line (without the trailing '\n')
		var lineBuf []rune
		for l.current != '\n' && l.current != 0 {
			lineBuf = append(lineBuf, l.current)
			l.readRune()
		}
		line := string(lineBuf)

		// If line equals the tag exactly, stop. Do not include this line.
		if line == tag {
			// Consume the newline after the terminator line if present.
			if l.current == '\n' {
				l.readRune()
			}
			break
		}

		// Otherwise, append the line and restore newline if we had one.
		sb.WriteString(line)
		switch l.current {
		case '\n':
			sb.WriteByte('\n')
			l.readRune()
		case 0:
			// EOF before terminator
			return "", UnterminatedStringError(l.currentPosition())
		}
	}
	return sb.String(), nil
}
