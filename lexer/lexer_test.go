// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lexer

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
)

func collectKinds(input string) []tokens.Kind {
	l := NewLexer(strings.NewReader(input), "test.sent")
	kinds := []tokens.Kind{}
	for {
		tok := l.NextToken()
		kinds = append(kinds, tok.Kind)
		if tok.Kind == tokens.EOF || tok.Kind == tokens.Error {
			break
		}
	}
	return kinds
}

func mustNextToken(t *testing.T, l *Lexer) tokens.Instance {
	t.Helper()
	tok := l.NextToken()
	return tok
}

func TestLexerHereDocRejectsInvalidTagStart(t *testing.T) {
	l := NewLexer(strings.NewReader("<<<1TAG\nbody\nTAG\n"), "test.sent")

	tok := l.NextToken()
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s", tok.Kind)
	}
	if !strings.Contains(tok.Value, "heredoc requires identifier tag") {
		t.Fatalf("unexpected literal: %q", tok.Value)
	}
}

func TestLexerPipelineToken(t *testing.T) {
	l := NewLexer(strings.NewReader("value |> len"), "test.sent")

	tok := mustNextToken(t, l)
	if tok.Kind != tokens.Ident || tok.Value != "value" {
		t.Fatalf("expected first ident token, got %s(%q)", tok.Kind, tok.Value)
	}

	tok = mustNextToken(t, l)
	if tok.Kind != tokens.TokenPipeForward || tok.Value != "|>" {
		t.Fatalf("expected pipeline token, got %s(%q)", tok.Kind, tok.Value)
	}

	tok = mustNextToken(t, l)
	if tok.Kind != tokens.Ident || tok.Value != "len" {
		t.Fatalf("expected rhs ident token, got %s(%q)", tok.Kind, tok.Value)
	}
}

func TestLexerPipelineNoWhitespace(t *testing.T) {
	got := collectKinds("a|>b")
	want := []tokens.Kind{tokens.Ident, tokens.TokenPipeForward, tokens.Ident, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestLexerPipelineMultilineSequence(t *testing.T) {
	got := collectKinds("value\n|> string.trim\n|> len")
	want := []tokens.Kind{
		tokens.Ident,
		tokens.TokenPipeForward,
		tokens.KeywordString,
		tokens.TokenDot,
		tokens.Ident,
		tokens.TokenPipeForward,
		tokens.Ident,
		tokens.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

// A bare `|` separates union type members; the parser rejects it in an
// expression, where only `|>` is an operator.
func TestLexerBarePipe(t *testing.T) {
	cases := map[string][]tokens.Kind{
		"value | len":  {tokens.Ident, tokens.TokenPipe, tokens.Ident, tokens.EOF},
		"value |":      {tokens.Ident, tokens.TokenPipe, tokens.EOF},
		"value || len": {tokens.Ident, tokens.TokenPipe, tokens.TokenPipe, tokens.Ident, tokens.EOF},
		"a|b|>c":       {tokens.Ident, tokens.TokenPipe, tokens.Ident, tokens.TokenPipeForward, tokens.Ident, tokens.EOF},
	}
	for src, want := range cases {
		got := collectKinds(src)
		if len(got) != len(want) {
			t.Fatalf("%q: expected %d tokens, got %d: %v", src, len(want), len(got), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%q: token %d: expected %s, got %s", src, i, want[i], got[i])
			}
		}
	}
}

func TestLexerPipelineHoleToken(t *testing.T) {
	got := collectKinds("#")
	want := []tokens.Kind{tokens.TokenPipelineHole, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestLexerPipelineHoleInsideCall(t *testing.T) {
	got := collectKinds("f(a, #, b)")
	want := []tokens.Kind{
		tokens.Ident,
		tokens.PunctLeftParentheses,
		tokens.Ident,
		tokens.PunctComma,
		tokens.TokenPipelineHole,
		tokens.PunctComma,
		tokens.Ident,
		tokens.PunctRightParentheses,
		tokens.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

// TestLexerTokenPositions checks that lines and columns are 1-based on every
// line and that a token ends just past its last rune on its own line, whether
// a space, a newline or EOF follows it (#236).
func TestLexerTokenPositions(t *testing.T) {
	type span struct{ fromLine, fromCol, toLine, toCol, toOff int }
	testCases := []struct {
		input string
		want  []span
	}{
		{"ab cd", []span{{1, 1, 1, 3, 2}, {1, 4, 1, 6, 5}}},
		{"ab\ncd\n", []span{{1, 1, 1, 3, 2}, {2, 1, 2, 3, 5}}},
		{"x == y\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}, {1, 6, 1, 7, 6}}},
		{"\"é\"\n-- note\n", []span{{1, 1, 1, 4, 4}, {2, 1, 2, 8, 12}}},
		{"<<<T\na\nT\nx", []span{{1, 1, 3, 2, 8}, {4, 1, 4, 2, 10}}},
		// A CRLF heredoc and a terminator with trailing whitespace end just
		// past the tag (#270).
		{"<<<T\r\na\r\nT\r\nx", []span{{1, 1, 3, 2, 10}, {4, 1, 4, 2, 13}}},
		{"<<<T\na\nT \t\nx", []span{{1, 1, 3, 2, 8}, {4, 1, 4, 2, 12}}},
		// These operators took their end before their last read and ended one
		// column short; check them mid-line, at the end of a line and at EOF.
		{"a-b", []span{{1, 1, 1, 2, 1}, {1, 2, 1, 3, 2}, {1, 3, 1, 4, 3}}},
		{"x -\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 4, 3}}},
		{"x -", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 4, 3}}},
		{"x ==\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x ==", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x =>\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x =>", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x |>\n", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
		{"x |>", []span{{1, 1, 1, 2, 1}, {1, 3, 1, 5, 4}}},
	}
	for _, tc := range testCases {
		l := NewLexer(strings.NewReader(tc.input), "t.sentrie")
		for i, want := range tc.want {
			tok := l.NextToken()
			got := span{tok.Range.From.Line, tok.Range.From.Column, tok.Range.To.Line, tok.Range.To.Column, tok.Range.To.Offset}
			if got != want {
				t.Fatalf("input %q token %d (%s %q): got %+v, want %+v", tc.input, i, tok.Kind, tok.Value, got, want)
			}
		}
		if eof := l.NextToken(); eof.Kind != tokens.EOF {
			t.Fatalf("input %q: expected EOF, got %s", tc.input, eof.Kind)
		}
	}
}

// TestLexerCommentSpanIgnoresCRLF checks that a comment token covers the same
// line and columns whether its line ends in LF or CRLF, so the \r of a CRLF
// line break is not part of the span (#279). Offsets differ by the extra \r
// bytes, so the span length is compared instead.
func TestLexerCommentSpanIgnoresCRLF(t *testing.T) {
	type span struct {
		kind                                tokens.Kind
		value                               string
		fromLine, fromCol, toLine, toCol, n int
	}
	lex := func(input string) []span {
		l := NewLexer(strings.NewReader(input), "t.sentrie")
		var out []span
		for {
			tok := l.NextToken()
			if tok.Kind == tokens.EOF {
				return out
			}
			r := tok.Range
			out = append(out, span{tok.Kind, tok.Value, r.From.Line, r.From.Column, r.To.Line, r.To.Column, r.To.Offset - r.From.Offset})
		}
	}
	testCases := []struct {
		name  string
		input string
		kind  tokens.Kind
	}{
		{"line comment", "-- note\nx\n", tokens.LineComment},
		{"indented line comment", "  -- note  \nx\n", tokens.LineComment},
		{"trailing comment", "x -- note\ny\n", tokens.TrailingComment},
		{"empty comment", "x --\ny\n", tokens.TrailingComment},
		{"comment between lines", "a\n-- note\nb\n", tokens.LineComment},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lf := lex(tc.input)
			crlf := lex(strings.ReplaceAll(tc.input, "\n", "\r\n"))
			if len(lf) != len(crlf) {
				t.Fatalf("token count differs: LF %+v, CRLF %+v", lf, crlf)
			}
			seen := false
			for i := range lf {
				if lf[i] != crlf[i] {
					t.Fatalf("token %d differs: LF %+v, CRLF %+v", i, lf[i], crlf[i])
				}
				if lf[i].kind == tc.kind {
					seen = true
				}
			}
			if !seen {
				t.Fatalf("no %s token in %+v", tc.kind, lf)
			}
		})
	}

	// A \r that is not followed by \n is not a line break and stays in the
	// comment, in the middle of the text.
	got := lex("-- a\rb\n")
	if len(got) != 1 || got[0].value != "a\rb" {
		t.Fatalf("lone \\r inside a comment: got %+v", got)
	}
}

// TestLexerCommentEndOfFileAndFollowingTokensIgnoreCRLF checks the places a
// CRLF comment can end: at the end of the file with and without a final line
// break, and before more tokens, whose line and column must match the LF file
// (#279).
func TestLexerCommentEndOfFileAndFollowingTokensIgnoreCRLF(t *testing.T) {
	type tok struct {
		kind                                tokens.Kind
		value                               string
		fromLine, fromCol, toLine, toCol, n int
	}
	lex := func(input string) []tok {
		l := NewLexer(strings.NewReader(input), "t.sentrie")
		var out []tok
		for {
			tk := l.NextToken()
			if tk.Kind == tokens.EOF {
				return out
			}
			r := tk.Range
			out = append(out, tok{tk.Kind, tk.Value, r.From.Line, r.From.Column, r.To.Line, r.To.Column, r.To.Offset - r.From.Offset})
		}
	}
	testCases := []struct {
		name  string
		input string
		want  int // number of tokens
	}{
		{"last line, no final break", "x\n-- note", 2},
		{"last line, final break", "x\n-- note\n", 2},
		{"trailing comment on the last line, no final break", "x -- note", 2},
		{"only a comment, final break", "-- note\n", 1},
		{"only a comment, no final break", "-- note", 1},
		{"tokens on the next lines", "x -- one\ny -- two\n\n-- three\nz = 1\n", 8},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lf := lex(tc.input)
			crlf := lex(strings.ReplaceAll(tc.input, "\n", "\r\n"))
			if len(lf) != tc.want {
				t.Fatalf("LF token count: got %d, want %d: %+v", len(lf), tc.want, lf)
			}
			if len(lf) != len(crlf) {
				t.Fatalf("token count differs: LF %+v, CRLF %+v", lf, crlf)
			}
			for i := range lf {
				if lf[i] != crlf[i] {
					t.Fatalf("token %d differs: LF %+v, CRLF %+v", i, lf[i], crlf[i])
				}
			}
		})
	}

	// The next token's position is checked outright, not only against LF: the
	// token after a CRLF comment starts at line 2, column 1.
	got := lex("-- note\r\nx\r\n")
	if len(got) != 2 || got[1].kind != tokens.Ident || got[1].fromLine != 2 || got[1].fromCol != 1 {
		t.Fatalf("token after a CRLF comment: got %+v", got)
	}

	// A lone \r at the very end of the file (nothing follows it) is not a
	// line break either: it stays in the comment's span, trimmed from the text.
	got = lex("-- note\r")
	if len(got) != 1 || got[0].value != "note" || got[0].n != len("-- note\r") {
		t.Fatalf("lone \\r at the end of the comment: got %+v", got)
	}

	// Only the \r that directly precedes the \n is the line break: in \r\r\n
	// the first \r is a lone one and stays in the comment's span, and the
	// following token still lands on the next line.
	got = lex("-- note\r\r\nx")
	if len(got) != 2 || got[0].value != "note" || got[0].n != len("-- note\r") || got[1].fromLine != 2 || got[1].fromCol != 1 {
		t.Fatalf("CR CR LF after a comment: got %+v", got)
	}
}

// TestLexerStringLiteralLineBreaks checks that a line break inside a string
// literal is \n whether the file uses LF or CRLF, while an escaped \r and a
// lone raw \r are kept (#280).
func TestLexerStringLiteralLineBreaks(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{"LF", "\"a\nb\"", "a\nb"},
		{"CRLF", "\"a\r\nb\"", "a\nb"},
		{"two CRLF breaks", "\"a\r\n\r\nb\"", "a\n\nb"},
		{"escaped CR before an LF break", "\"a\\r\nb\"", "a\r\nb"},
		{"escaped CR before a CRLF break", "\"a\\r\r\nb\"", "a\r\nb"},
		{"escaped CR alone", "\"a\\rb\"", "a\rb"},
		{"lone raw CR", "\"a\rb\"", "a\rb"},
		{"lone raw CR before the closing quote", "\"a\r\"", "a\r"},
		{"escaped line break, LF", "\"a\\\nb\"", "a\nb"},
		{"escaped line break, CRLF", "\"a\\\r\nb\"", "a\nb"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok := NewLexer(strings.NewReader(tc.input), "t.sentrie").NextToken()
			if tok.Kind != tokens.String || tok.Value != tc.want {
				t.Fatalf("input %q: got %s %q, want String %q", tc.input, tok.Kind, tok.Value, tc.want)
			}
		})
	}

	// The token's span still covers the source text, including the \r bytes.
	tok := NewLexer(strings.NewReader("\"a\r\nb\" x"), "t.sentrie").NextToken()
	if tok.Range.From.Line != 1 || tok.Range.To.Line != 2 || tok.Range.To.Column != 3 || tok.Range.To.Offset != 6 {
		t.Fatalf("CRLF string span: got %+v", tok.Range)
	}
}

// TestLexerStringLiteralCRLFPositions checks that the tokens after a string
// literal that spans lines sit on the same line and column in a CRLF file as in
// the LF file, and that an unterminated multi-line string reports the same
// error position (#280).
func TestLexerStringLiteralCRLFPositions(t *testing.T) {
	type tok struct {
		kind                                tokens.Kind
		value                               string
		fromLine, fromCol, toLine, toCol, n int
	}
	lex := func(input string) []tok {
		l := NewLexer(strings.NewReader(input), "t.sentrie")
		var out []tok
		for {
			tk := l.NextToken()
			if tk.Kind == tokens.EOF {
				return out
			}
			r := tk.Range
			out = append(out, tok{tk.Kind, tk.Value, r.From.Line, r.From.Column, r.To.Line, r.To.Column, r.To.Offset - r.From.Offset})
			if tk.Kind == tokens.Error {
				return out
			}
		}
	}

	t.Run("tokens after a multi-line string", func(t *testing.T) {
		lf := lex("let s = \"a\nb\nc\" + x\ny\n")
		crlf := lex("let s = \"a\r\nb\r\nc\" + x\r\ny\r\n")
		if len(lf) != len(crlf) || len(lf) < 7 {
			t.Fatalf("token count: LF %+v, CRLF %+v", lf, crlf)
		}
		for i := range lf {
			a, b := lf[i], crlf[i]
			// The string's own span covers the dropped \r bytes, so only its
			// length differs from LF; every other token matches outright.
			if a.kind == tokens.String {
				if b.n != a.n+2 {
					t.Fatalf("string span length: LF %d, CRLF %d", a.n, b.n)
				}
				a.n, b.n = 0, 0
			}
			if a != b {
				t.Fatalf("token %d differs: LF %+v, CRLF %+v", i, lf[i], crlf[i])
			}
		}
		// The identifier after the string is on line 3, column 6 in both.
		var x *tok
		for i := range crlf {
			if crlf[i].value == "x" {
				x = &crlf[i]
			}
		}
		if x == nil || x.fromLine != 3 || x.fromCol != 6 {
			t.Fatalf("x after the string: got %+v", x)
		}
	})

	t.Run("unterminated multi-line string", func(t *testing.T) {
		for _, input := range []string{"x = \"a\nb", "x = \"a\nb\nc"} {
			lf := lex(input)
			crlf := lex(strings.ReplaceAll(input, "\n", "\r\n"))
			lastLF, lastCRLF := lf[len(lf)-1], crlf[len(crlf)-1]
			if lastLF.kind != tokens.Error || lastCRLF.kind != tokens.Error {
				t.Fatalf("input %q: want an error token, got LF %+v, CRLF %+v", input, lastLF, lastCRLF)
			}
			if !strings.Contains(lastCRLF.value, "unterminated string literal") || lastLF.value != lastCRLF.value {
				t.Fatalf("input %q: error differs: LF %q, CRLF %q", input, lastLF.value, lastCRLF.value)
			}
		}
	})

	t.Run("CR CR LF drops only the last CR", func(t *testing.T) {
		tok := NewLexer(strings.NewReader("\"a\r\r\nb\""), "t.sentrie").NextToken()
		if tok.Kind != tokens.String || tok.Value != "a\r\nb" {
			t.Fatalf("got %s %q, want String %q", tok.Kind, tok.Value, "a\r\nb")
		}
	})

	t.Run("CR at the end of the file inside an unterminated string", func(t *testing.T) {
		// A string cut off by the end of the file after a \r is still reported
		// as unterminated.
		for _, input := range []string{"\"a\r", "\"a\r\n\r"} {
			got := lex(input)
			last := got[len(got)-1]
			if last.kind != tokens.Error || !strings.Contains(last.value, "unterminated string literal") {
				t.Fatalf("input %q: got %+v", input, got)
			}
		}
	})
}

func TestLexerNumberExponent(t *testing.T) {
	cases := []struct {
		input string
		kind  tokens.Kind
		value string
	}{
		{"42", tokens.Int, "42"},
		{"3.14", tokens.Float, "3.14"},
		{"1e5", tokens.Float, "1e5"},
		{"1E5", tokens.Float, "1E5"},
		{"1e+5", tokens.Float, "1e+5"},
		{"1e-5", tokens.Float, "1e-5"},
		{"2.5e-3", tokens.Float, "2.5e-3"},
		{"2.5E+10", tokens.Float, "2.5E+10"},
		{"0e0", tokens.Float, "0e0"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			l := NewLexer(strings.NewReader(tc.input), "test.sent")
			tok := l.NextToken()
			if tok.Kind != tc.kind || tok.Value != tc.value {
				t.Fatalf("got %s %q, want %s %q", tok.Kind, tok.Value, tc.kind, tc.value)
			}
			if tok.Range.To.Column != len(tc.input)+1 {
				t.Fatalf("token ends at column %d, want %d", tok.Range.To.Column, len(tc.input)+1)
			}
			if next := l.NextToken(); next.Kind != tokens.EOF {
				t.Fatalf("expected EOF after %q, got %s %q", tc.input, next.Kind, next.Value)
			}
		})
	}
}

func TestLexerNumberExponentInExpression(t *testing.T) {
	got := collectKinds("1e5+2.5e-3*x")
	want := []tokens.Kind{tokens.Float, tokens.TokenPlus, tokens.Float, tokens.TokenMul, tokens.Ident, tokens.EOF}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %s, want %s (all: %v)", i, got[i], want[i], got)
		}
	}
}

func TestLexerRejectsMalformedExponent(t *testing.T) {
	cases := []struct {
		input   string
		literal string
		message string
	}{
		{"1e", `"1e"`, "the exponent needs at least one digit"},
		{"1E", `"1E"`, "the exponent needs at least one digit"},
		{"1e+", `"1e+"`, "the exponent needs at least one digit"},
		{"1e-", `"1e-"`, "the exponent needs at least one digit"},
		{"1ex", `"1e"`, "the exponent needs at least one digit"},
		{"2.5e", `"2.5e"`, "the exponent needs at least one digit"},
		{"1e+ 5", `"1e+"`, "the exponent needs at least one digit"},
		{"1.e5", `"1.e5"`, "a digit must follow the decimal point"},
		{"1.E-5", `"1.E-5"`, "a digit must follow the decimal point"},
		{"1.e", `"1.e"`, "a digit must follow the decimal point"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			l := NewLexer(strings.NewReader(tc.input), "test.sent")
			tok := l.NextToken()
			if tok.Kind != tokens.Error {
				t.Fatalf("expected error token, got %s %q", tok.Kind, tok.Value)
			}
			if !strings.Contains(tok.Value, "invalid number literal "+tc.literal) || !strings.Contains(tok.Value, tc.message) {
				t.Fatalf("unexpected message %q", tok.Value)
			}
			if tok.Range.To.Column != len(strings.Trim(tc.literal, `"`))+1 {
				t.Fatalf("error token ends at column %d, want it to cover %s", tok.Range.To.Column, tc.literal)
			}
		})
	}
}

// TestLexerExponentMarkerOutsideNumbers tests that the exponent rule applies
// only straight after digits: `x.e5` is a field access, and `1.foo` is still an
// integer, a dot and an identifier.
func TestLexerExponentMarkerOutsideNumbers(t *testing.T) {
	cases := []struct {
		input string
		want  []tokens.Kind
	}{
		{"x.e5", []tokens.Kind{tokens.Ident, tokens.TokenDot, tokens.Ident, tokens.EOF}},
		{"e5", []tokens.Kind{tokens.Ident, tokens.EOF}},
		{"1.foo", []tokens.Kind{tokens.Int, tokens.TokenDot, tokens.Ident, tokens.EOF}},
		{"1 e5", []tokens.Kind{tokens.Int, tokens.Ident, tokens.EOF}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := collectKinds(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("token %d: got %s, want %s (all: %v)", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestLexerUnknownCharacterNamesTheCharacter(t *testing.T) {
	l := NewLexer(strings.NewReader("$"), "test.sent")
	tok := l.NextToken()
	if tok.Kind != tokens.Error {
		t.Fatalf("expected error token, got %s %q", tok.Kind, tok.Value)
	}
	if tok.Value != `unexpected character "$"` {
		t.Fatalf("unexpected message %q", tok.Value)
	}
	if tok.Range.From.Column != 1 || tok.Range.To.Column != 2 {
		t.Fatalf("error token spans columns %d-%d, want 1-2", tok.Range.From.Column, tok.Range.To.Column)
	}
}
