// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sentrie-sh/sentrie/tokens"
)

type Lexer struct {
	reader   *bufio.Reader
	filename string

	// line, column and offset locate the current rune: line and column are
	// 1-based on every line, offset is the 0-based byte offset.
	line   int
	column int

	current     rune
	currentLine []rune // buffer for lookback

	offset       int
	currentWidth int
	atEOF        bool

	// end is the position just past the last consumed rune, on that rune's
	// own line. A token's end is read from here, so a token before a newline
	// or at EOF ends where its last rune does.
	end tokens.Pos

	identRegex *regexp.Regexp

	// pushBack is a LIFO stack: NextToken pops from here before lexing more input.
	pushBack []tokens.Instance
}

func NewLexer(reader io.Reader, filename string) *Lexer {
	l := &Lexer{
		reader:      bufio.NewReader(reader),
		filename:    filename,
		currentLine: []rune{},
		identRegex:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
		line:        1,
		column:      1,
		end:         tokens.Pos{Line: 1, Column: 1},
	}
	l.readRune() // Initialize the first rune
	return l
}

// PushBack returns a token to the stream so the next NextToken yields it.
// Used for speculative lookahead in the parser (e.g. grouped vs lambda).
func (l *Lexer) PushBack(t tokens.Instance) {
	l.pushBack = append(l.pushBack, t)
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() tokens.Instance {
	if len(l.pushBack) > 0 {
		i := len(l.pushBack) - 1
		t := l.pushBack[i]
		l.pushBack = l.pushBack[:i]
		return t
	}
	for {
		l.skipWhitespace()

		if l.current == 0 {
			return tokens.EofInstance(l.filename, l.currentPosition())
		}

		startPos := l.currentPosition()

		switch l.current {
		case '-':
			if l.peekAhead() == '-' {
				l.readRune() // consume second '-'
				commentKind, value := l.readComment()
				endPos := l.endPosition()
				return tokens.New(commentKind, value, tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenMinus, "-", tokens.NewRange(l.filename, startPos, endPos))

		case '=':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenEq, "==", tokens.NewRange(l.filename, startPos, endPos))
			}
			if l.peekAhead() == '>' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenFatArrow, "=>", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenAssign, "=", tokens.NewRange(l.filename, startPos, endPos))

		case '!':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenNeq, "!=", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenBang, "!", tokens.NewRange(l.filename, startPos, endPos))

		case '<':
			if l.peekString(2) == "<<" {
				value, endPos, err := l.readHereDoc()
				if err != nil {
					endPos := l.endPosition()
					return tokens.New(tokens.Error, err.Error(), tokens.NewRange(l.filename, startPos, endPos))
				}
				return tokens.New(tokens.String, value, tokens.NewRange(l.filename, startPos, endPos))
			}
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenLte, "<=", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenLt, "<", tokens.NewRange(l.filename, startPos, endPos))

		case '>':
			if l.peekAhead() == '=' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenGte, ">=", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenGt, ">", tokens.NewRange(l.filename, startPos, endPos))

		case '+':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenPlus, "+", tokens.NewRange(l.filename, startPos, endPos))
		case '*':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenMul, "*", tokens.NewRange(l.filename, startPos, endPos))
		case '/':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenDiv, "/", tokens.NewRange(l.filename, startPos, endPos))
		case '%':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenMod, "%", tokens.NewRange(l.filename, startPos, endPos))
		case '?':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenQuestion, "?", tokens.NewRange(l.filename, startPos, endPos))
		case ':':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctColon, ":", tokens.NewRange(l.filename, startPos, endPos))
		case '.':
			if l.peekString(2) == ".." {
				l.readRune()
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenDotDotDot, "...", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenDot, ".", tokens.NewRange(l.filename, startPos, endPos))
		case '@':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenAt, "@", tokens.NewRange(l.filename, startPos, endPos))
		case '#':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.TokenPipelineHole, "#", tokens.NewRange(l.filename, startPos, endPos))
		case '|':
			if l.peekAhead() == '>' {
				l.readRune()
				l.readRune()
				endPos := l.endPosition()
				return tokens.New(tokens.TokenPipeForward, "|>", tokens.NewRange(l.filename, startPos, endPos))
			}
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.Error, "unexpected '|': only '|>' is supported", tokens.NewRange(l.filename, startPos, endPos))
		case ',':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctComma, ",", tokens.NewRange(l.filename, startPos, endPos))
		case ';':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctSemicolon, ";", tokens.NewRange(l.filename, startPos, endPos))
		case '(':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctLeftParentheses, "(", tokens.NewRange(l.filename, startPos, endPos))
		case ')':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctRightParentheses, ")", tokens.NewRange(l.filename, startPos, endPos))
		case '{':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctLeftCurly, "{", tokens.NewRange(l.filename, startPos, endPos))
		case '}':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctRightCurly, "}", tokens.NewRange(l.filename, startPos, endPos))
		case '[':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctLeftBracket, "[", tokens.NewRange(l.filename, startPos, endPos))
		case ']':
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.PunctRightBracket, "]", tokens.NewRange(l.filename, startPos, endPos))

		case '"':
			value, err := l.readString()
			if err != nil {
				endPos := l.endPosition()
				return tokens.New(tokens.Error, err.Error(), tokens.NewRange(l.filename, startPos, endPos))
			}
			endPos := l.endPosition()
			return tokens.New(tokens.String, value, tokens.NewRange(l.filename, startPos, endPos))

		default:
			if unicode.IsLetter(l.current) || l.current == '_' {
				value := l.readIdentifier()
				if !l.identRegex.MatchString(value) {
					endPos := l.endPosition()
					return tokens.Err(tokens.NewRange(l.filename, startPos, endPos), "invalid identifier: "+value)
				}

				// is this a known keyword?
				if kind, isKeyword := tokens.IsKeyword(value); isKeyword {
					endPos := l.endPosition()
					return tokens.New(kind, value, tokens.NewRange(l.filename, startPos, endPos))
				}
				endPos := l.endPosition()
				return tokens.New(tokens.Ident, value, tokens.NewRange(l.filename, startPos, endPos))
			}

			if unicode.IsDigit(l.current) {
				value, kind := l.readNumber()
				endPos := l.endPosition()
				return tokens.New(kind, value, tokens.NewRange(l.filename, startPos, endPos))
			}

			// Unknown character
			char := string(l.current)
			l.readRune()
			endPos := l.endPosition()
			return tokens.New(tokens.Error, char, tokens.NewRange(l.filename, startPos, endPos))
		}
	}
}

// readRune consumes the current rune and reads the next one from input.
func (l *Lexer) readRune() {
	// Step past the rune being consumed: record where it ends, then move the
	// position to the rune after it.
	if l.currentWidth > 0 {
		l.end = tokens.Pos{
			Line:   l.line,
			Column: l.column + 1,
			Offset: l.offset + l.currentWidth,
		}
		l.offset += l.currentWidth
		if l.current == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
	}

	if l.atEOF {
		l.current = 0
		l.currentWidth = 0
		return
	}

	r, size, err := l.reader.ReadRune()
	if err != nil {
		// io.EOF ends the input; any other read error is treated as EOF too.
		l.atEOF = true
		l.current = 0
		l.currentWidth = 0
		return
	}

	l.current = r
	l.currentWidth = size

	// Update the current line buffer
	l.currentLine = append(l.currentLine, r)
	if r == '\n' {
		l.currentLine = []rune{}
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

// currentPosition returns the position of the current rune, which is where
// the next token starts. At EOF it is the position just past the input.
func (l *Lexer) currentPosition() tokens.Pos {
	return tokens.Pos{
		Offset: l.offset,
		Line:   l.line,
		Column: l.column,
	}
}

// endPosition returns the exclusive end of the token just read: the position
// just past its last rune, on that rune's line. It is not the position of the
// next rune, which is on the next line when the token ends a line.
func (l *Lexer) endPosition() tokens.Pos {
	return l.end
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
		return "", UnterminatedStringError(l.filename, l.currentPosition())
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

// readHereDoc reads a heredoc starting at the first '<' of '<<<'. It returns
// the value and the heredoc's end, just past the terminator tag.
//
// Line endings may be LF or CRLF: a trailing '\r' is dropped from every body
// line, so the value always has LF line endings, and the terminator line may
// carry trailing spaces, tabs and a '\r' after the tag (#270).
func (l *Lexer) readHereDoc() (string, tokens.Pos, error) {
	start := l.currentPosition()

	// We are currently on the first '<'. Consume the 3 '<'.
	l.readRune() // 1st '<'
	if l.current != '<' {
		return "", tokens.Pos{}, io.ErrUnexpectedEOF
	}
	l.readRune() // 2nd '<'
	if l.current != '<' {
		return "", tokens.Pos{}, io.ErrUnexpectedEOF
	}
	l.readRune() // 3rd '<'

	// Disallow spaces before tag to keep syntax tight.
	// Require TAG immediately.
	if !unicode.IsLetter(l.current) && l.current != '_' {
		return "", tokens.Pos{}, fmt.Errorf("heredoc requires identifier tag after <<<: %w", InvalidHereDocSyntaxError(l.filename, l.currentPosition()))
	}

	// Read TAG (identifier)
	var tagBuilder strings.Builder
	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		tagBuilder.WriteRune(l.current)
		l.readRune()
	}
	tag := tagBuilder.String()
	if tag == "" || !l.identRegex.MatchString(tag) {
		return "", tokens.Pos{}, fmt.Errorf("invalid heredoc tag: %w", InvalidHereDocSyntaxError(l.filename, l.currentPosition()))
	}

	// Read to end of the tag line
	for l.current != '\n' && l.current != 0 {
		// No trailing junk allowed (only whitespace)
		if !unicode.IsSpace(l.current) && l.current != '\r' {
			return "", tokens.Pos{}, fmt.Errorf("unexpected characters after heredoc tag %q: %w", tag, InvalidHereDocSyntaxError(l.filename, l.currentPosition()))
		}
		l.readRune()
	}
	// Consume the '\n' if present
	if l.current == '\n' {
		l.readRune()
	}

	// Now collect lines until the terminator: a line that is the tag, with no
	// indentation, optionally followed by spaces, tabs and a '\r'.
	var sb strings.Builder
	for {
		// Capture the current line (without the trailing '\n'). tagEnd is
		// just past the line's first len(tag) runes: the heredoc's end if
		// this line turns out to be the terminator.
		var lineBuf []rune
		tagEnd := l.endPosition()
		for l.current != '\n' && l.current != 0 {
			lineBuf = append(lineBuf, l.current)
			l.readRune()
			if len(lineBuf) == len(tag) {
				tagEnd = l.endPosition()
			}
		}
		line := string(lineBuf)

		// The terminator ends the heredoc and is not part of the value. Its
		// trailing whitespace and newline are left for skipWhitespace, and
		// the heredoc ends just past its tag.
		if strings.TrimRight(line, " \t\r") == tag {
			return sb.String(), tagEnd, nil
		}

		// Otherwise, append the line without a CRLF's '\r' and restore the
		// newline if we had one.
		sb.WriteString(strings.TrimSuffix(line, "\r"))
		switch l.current {
		case '\n':
			sb.WriteByte('\n')
			l.readRune()
		case 0:
			// EOF before the terminator. Point at the heredoc's start: the
			// end of the file says nothing about which heredoc is open.
			return "", tokens.Pos{}, fmt.Errorf("heredoc <<<%s has no terminator line %q: %w", tag, tag, UnterminatedStringError(l.filename, start))
		}
	}
}
