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

package tokens

import "fmt"

// Pos represents a location within source code.
type Pos struct {
	// Line is the line number, starting from 1.
	Line int

	// Column is the column number, starting from 1.
	// This counts display characters, not bytes.
	Column int

	// Offset is the 0-based byte offset into the source file.
	// This points to the first byte of the UTF-8 sequence for the character.
	Offset int
}

// Range represents a contiguous region of source code.
type Range struct {
	// File is the source file name.
	File string

	// From is the start position (inclusive).
	From Pos

	// To is the end position (inclusive).
	To Pos
}

// String formats the span as "file:line:col-line:col" or "file:line:col-col" for single lines.
func (s Range) String() string {
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
