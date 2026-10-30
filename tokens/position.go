// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package tokens

import (
	"fmt"
)

// Pos represents a location within source code.
type Pos struct {
	// Line is the line number, starting from 1.
	Line int

	// Column is the column number, starting from 1.
	// This counts Unicode code points (runes), not bytes or display width.
	// LSP clients count UTF-16 code units, so editor integrations must convert.
	Column int

	// Offset is the 0-based byte offset into the source file.
	// This points to the first byte of the UTF-8 sequence for the character.
	Offset int

	//
	isBadPos bool
}

var BadPos = Pos{
	Line:     -1,
	Column:   -1,
	Offset:   -1,
	isBadPos: true,
}

func (p *Pos) IsBadPos() bool {
	return p.isBadPos
}

// Range represents a contiguous region of source code.
type Range struct {
	// File is the source file name.
	File string

	// From is the start position (inclusive).
	From Pos

	// To is the end position (exclusive): just past the last character, on
	// that character's line.
	To Pos
}

func NewRange(file string, from Pos, to Pos) Range {
	return Range{
		File: file,
		From: from,
		To:   to,
	}
}

func BadRange(file string) Range {
	return Range{
		File: file,
		From: BadPos,
		To:   BadPos,
	}
}

func NewRangeFromPos(file string, pos Pos) Range {
	return Range{
		File: file,
		From: pos,
		To:   pos,
	}
}

// String formats the span as "file:line:col-line:col", or "file:line:col-col"
// for a single line. Lines and columns are 1-based and printed as stored. A
// span with no width, or with an unset end, prints its start only.
func (s Range) String() string {
	if s.To.Line == 0 || (s.From.Line == s.To.Line && (s.From.Column == s.To.Column || s.To.Column == 0)) {
		return fmt.Sprintf("%s:%d:%d", s.File, s.From.Line, s.From.Column)
	}
	if s.From.Line == s.To.Line {
		return fmt.Sprintf("%s:%d:%d-%d", s.File, s.From.Line, s.From.Column, s.To.Column)
	}
	return fmt.Sprintf("%s:%d:%d-%d:%d", s.File, s.From.Line, s.From.Column, s.To.Line, s.To.Column)
}

// // Position represents a location within source code.
// type Position struct {
// 	// Filename is the source file name.
// 	Filename string

// 	// Line is the line number, starting from 1.
// 	Line int

// 	// Column is the column number, starting from 1.
// 	// This counts display characters, not bytes.
// 	Column int

// 	// Offset is the 0-based byte offset into the source file.
// 	// This points to the first byte of the UTF-8 sequence for the character.
// 	Offset int
// }

// func (p Position) String() string {
// 	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
// }
