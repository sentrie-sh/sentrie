// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
)

// SourceError is a module load failure positioned in the module's original
// source, formatted compiler-style as `file:line:col: message`.
type SourceError struct {
	File    string // the module's source name, e.g. @local/lib/helpers.ts
	Line    int    // 1-based; 0 when the position is unknown
	Column  int    // 1-based; 0 when unknown
	Message string
	Err     error // the underlying error, if any
}

func (e *SourceError) Error() string {
	switch {
	case e.Line > 0 && e.Column > 0:
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	case e.Line > 0:
		return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
	default:
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
}

func (e *SourceError) Unwrap() error { return e.Err }

// compileError wraps a goja compile failure. A compile-phase
// CompilerSyntaxError carries its file, whose positions goja maps through the
// inline source map. A parse error carries only a message whose position is
// in the transpiled output, because goja attaches the source map after parsing.
func compileError(name string, err error) error {
	var se *goja.CompilerSyntaxError
	if !errors.As(err, &se) {
		return &SourceError{File: name, Message: err.Error(), Err: err}
	}
	if se.File != nil {
		pos := se.File.Position(se.Offset)
		return &SourceError{File: pos.Filename, Line: pos.Line, Column: pos.Column, Message: "SyntaxError: " + se.Message, Err: err}
	}
	return &SourceError{File: name, Message: "SyntaxError (position in transpiled output): " + se.Message, Err: err}
}
