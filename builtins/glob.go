// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/sentrie-sh/sentrie/box"
)

// Limits keep glob_match at O(pattern * value) with a small, fixed ceiling.
const (
	maxGlobPatternRunes = 1024
	maxGlobValueRunes   = 8192
)

var declGlobMatch = &Decl{
	Name: "glob_match",
	Description: "Reports whether a value matches a glob pattern. `*` matches any run of characters that are not delimiters, " +
		"`**` matches any run of characters including delimiters, `?` matches one non-delimiter character, " +
		"`[a-z]` and `[!a-z]` match a character class, and `\\` escapes the next character. " +
		"The delimiters list must not be empty (OPA defaults an empty list to [\".\"], so it is rejected rather than guessed).",
	DeriveSafe: true,
	Sig: Sig{
		Params: []ParamSig{
			{Name: "pattern", Kinds: kindString, KindError: "glob_match: pattern must be a string"},
			{Name: "delimiters", Kinds: kindList, KindError: "glob_match: delimiters must be a list of single-character strings"},
			{Name: "value", Kinds: kindString, KindError: "glob_match: value must be a string"},
		},
		TooFewError:  "glob_match requires 3 arguments",
		TooManyError: "glob_match requires 3 arguments",
		Result:       ParamSig{Name: "result", Kinds: kindBool},
	},
	Impl: implGlobMatch,
}

type globTokKind uint8

const (
	globLiteral    globTokKind = iota
	globAny                    // ?
	globClass                  // [...]
	globStar                   // *
	globDoubleStar             // **
)

type globRange struct{ lo, hi rune }

type globTok struct {
	kind   globTokKind
	r      rune
	ranges []globRange
	negate bool
}

func implGlobMatch(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	// An undefined argument propagates; null falls through to the kind errors.
	for _, a := range args {
		if a.IsUndefined() {
			return box.Undefined(), nil
		}
	}
	pattern, ok := args[0].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("glob_match: pattern must be a string")
	}
	delims, err := globDelimiters(args[1])
	if err != nil {
		return box.Undefined(), err
	}
	value, ok := args[2].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("glob_match: value must be a string")
	}
	if utf8.RuneCountInString(value) > maxGlobValueRunes {
		return box.Undefined(), fmt.Errorf("glob_match: value is longer than %d characters", maxGlobValueRunes)
	}
	toks, err := parseGlob(pattern)
	if err != nil {
		return box.Undefined(), err
	}
	return box.Bool(globMatches(toks, []rune(value), delims)), nil
}

func globDelimiters(v box.Value) (map[rune]struct{}, error) {
	list, ok := v.ListValue()
	if !ok {
		return nil, fmt.Errorf("glob_match: delimiters must be a list of single-character strings")
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("glob_match: delimiters must not be empty (OPA treats an empty list as [\".\"], so pass the delimiter you mean, for example [\"/\"])")
	}
	out := make(map[rune]struct{}, len(list))
	for _, d := range list {
		s, ok := d.StringValue()
		if !ok || utf8.RuneCountInString(s) != 1 {
			return nil, fmt.Errorf("glob_match: delimiters must be a list of single-character strings")
		}
		r, _ := utf8.DecodeRuneInString(s)
		out[r] = struct{}{}
	}
	return out, nil
}

func parseGlob(pattern string) ([]globTok, error) {
	rs := []rune(pattern)
	if len(rs) > maxGlobPatternRunes {
		return nil, fmt.Errorf("glob_match: pattern is longer than %d characters", maxGlobPatternRunes)
	}
	var toks []globTok
	for i := 0; i < len(rs); {
		switch c := rs[i]; c {
		case '*':
			j := i
			for j < len(rs) && rs[j] == '*' {
				j++
			}
			if j-i >= 2 {
				toks = append(toks, globTok{kind: globDoubleStar})
			} else {
				toks = append(toks, globTok{kind: globStar})
			}
			i = j
		case '?':
			toks = append(toks, globTok{kind: globAny})
			i++
		case '\\':
			if i+1 >= len(rs) {
				return nil, fmt.Errorf("glob_match: pattern ends with an unescaped backslash")
			}
			toks = append(toks, globTok{kind: globLiteral, r: rs[i+1]})
			i += 2
		case '[':
			tok, next, err := parseGlobClass(rs, i)
			if err != nil {
				return nil, err
			}
			toks = append(toks, tok)
			i = next
		case '{', '}':
			return nil, fmt.Errorf("glob_match: unescaped %q is not supported (alternation is not available; escape it with a backslash)", c)
		default:
			toks = append(toks, globTok{kind: globLiteral, r: c})
			i++
		}
	}
	return toks, nil
}

// parseGlobClass parses the class that opens at rs[start] and returns it with
// the index just past the closing bracket.
func parseGlobClass(rs []rune, start int) (globTok, int, error) {
	tok := globTok{kind: globClass}
	i := start + 1
	if i < len(rs) && (rs[i] == '!' || rs[i] == '^') {
		tok.negate = true
		i++
	}
	read := func() (rune, error) {
		if rs[i] == '\\' {
			if i+1 >= len(rs) {
				return 0, fmt.Errorf("glob_match: unterminated character class")
			}
			i++
		}
		r := rs[i]
		i++
		return r, nil
	}
	for {
		if i >= len(rs) {
			return globTok{}, 0, fmt.Errorf("glob_match: unterminated character class")
		}
		if rs[i] == ']' {
			break
		}
		lo, err := read()
		if err != nil {
			return globTok{}, 0, err
		}
		hi := lo
		if i+1 < len(rs) && rs[i] == '-' && rs[i+1] != ']' {
			i++
			if hi, err = read(); err != nil {
				return globTok{}, 0, err
			}
			if hi < lo {
				return globTok{}, 0, fmt.Errorf("glob_match: invalid range %q-%q in character class", lo, hi)
			}
		}
		tok.ranges = append(tok.ranges, globRange{lo, hi})
	}
	if len(tok.ranges) == 0 {
		return globTok{}, 0, fmt.Errorf("glob_match: empty character class")
	}
	return tok, i + 1, nil
}

func (t globTok) classMatches(r rune) bool {
	in := false
	for _, rg := range t.ranges {
		if r >= rg.lo && r <= rg.hi {
			in = true
			break
		}
	}
	return in != t.negate
}

// globMatches reports whether toks match all of runes. match[j] means the
// tokens from the current one onward match runes[j:], so the work is
// O(len(toks) * len(runes)) with no backtracking.
func globMatches(toks []globTok, runes []rune, delims map[rune]struct{}) bool {
	n := len(runes)
	isDelim := func(r rune) bool { _, ok := delims[r]; return ok }
	next := make([]bool, n+1)
	next[n] = true
	cur := make([]bool, n+1)
	for i := len(toks) - 1; i >= 0; i-- {
		t := toks[i]
		for j := n; j >= 0; j-- {
			var m bool
			switch t.kind {
			case globLiteral:
				m = j < n && runes[j] == t.r && next[j+1]
			case globAny:
				m = j < n && !isDelim(runes[j]) && next[j+1]
			case globClass:
				m = j < n && !isDelim(runes[j]) && t.classMatches(runes[j]) && next[j+1]
			case globStar:
				m = next[j] || (j < n && !isDelim(runes[j]) && cur[j+1])
			case globDoubleStar:
				m = next[j] || (j < n && cur[j+1])
			}
			cur[j] = m
		}
		next, cur = cur, next
	}
	return next[0]
}
