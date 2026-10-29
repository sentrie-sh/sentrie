// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Package capabilities is the operator's grant: a deployment-level file that
// limits which builtins, which `@sentrie/*` modules, and whether local
// TypeScript modules a policy pack may use.
//
// A pack's `[permissions]` table (package pack) is the pack's request; a
// capabilities file is the operator's grant. A nil *Capabilities, the value
// when no file is given, grants everything, which is the behaviour without
// this package. Every method is safe on a nil receiver.
//
// The file format is documented in FORMAT.md in this directory.
package capabilities

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/sentrie-sh/sentrie/constants"
)

// CurrentVersion is the only file format version this build reads.
const CurrentVersion = 1

// Capabilities is a parsed capabilities file. A nil list or flag means the
// file did not restrict that dimension.
type Capabilities struct {
	// Version is the file format version and must equal CurrentVersion.
	Version int `json:"version"`

	// Builtins lists the native builtins a pack may call, by name ("count",
	// "now"). Absent grants every builtin; an empty list grants none.
	Builtins *[]string `json:"builtins,omitempty"`

	// Modules lists the builtin modules a pack may use, by full specifier
	// ("@sentrie/hash"). It applies to `use ... from @sentrie/...` and to
	// require() inside TypeScript. Absent grants every builtin module; an
	// empty list grants none.
	Modules *[]string `json:"modules,omitempty"`

	// LocalModules reports whether a pack may load its own TypeScript or
	// JavaScript files (`use ... from "./x"`, `@local/...`, relative require).
	// Absent grants them.
	LocalModules *bool `json:"local_modules,omitempty"`
}

var modulePrefix = "@" + constants.APPNAME + "/"

// Parse decodes and checks a capabilities file. An empty file, a top-level
// value that is not an object, a missing or unsupported version, unknown,
// differently-cased or duplicate keys, trailing data, empty names, and module
// entries without the "@sentrie/" prefix are errors. Syntax and type errors
// name the line and column. Names are not checked against the builtins and
// modules this build ships, so a file written for a newer build still loads:
// a name this build does not know grants nothing.
func Parse(data []byte) (*Capabilities, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("capabilities: empty file")
	}
	if err := checkTopLevelKeys(data); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var c Capabilities
	if err := dec.Decode(&c); err != nil {
		return nil, positioned(data, dec.InputOffset(), err)
	}

	if c.Version != CurrentVersion {
		return nil, fmt.Errorf("capabilities: unsupported version %d (this build reads version %d)", c.Version, CurrentVersion)
	}
	if c.Builtins != nil {
		for _, name := range *c.Builtins {
			if strings.TrimSpace(name) == "" {
				return nil, fmt.Errorf("capabilities: builtins: empty name")
			}
		}
	}
	if c.Modules != nil {
		for _, key := range *c.Modules {
			if !strings.HasPrefix(key, modulePrefix) || len(key) == len(modulePrefix) {
				return nil, fmt.Errorf("capabilities: modules: %q must be a full specifier such as %q", key, modulePrefix+"hash")
			}
		}
	}
	return &c, nil
}

// knownKeys are the exact JSON keys of Capabilities. encoding/json matches
// keys case-insensitively and lets a later duplicate win, which is unsafe for
// an allow-list, so checkTopLevelKeys holds the file to these spellings.
var knownKeys = func() map[string]bool {
	out := map[string]bool{}
	t := reflect.TypeFor[Capabilities]()
	for i := range t.NumField() {
		if name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ","); name != "" && name != "-" {
			out[name] = true
		}
	}
	return out
}()

// checkTopLevelKeys walks the top-level object once, rejecting a non-object
// value, unknown or differently-cased keys, duplicate keys, a missing
// "version" and trailing data, each with its position.
func checkTopLevelKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return positioned(data, dec.InputOffset(), err)
	}
	if tok != json.Delim('{') {
		return fmt.Errorf("capabilities: the top-level value must be an object")
	}

	seen := map[string]bool{}
	for dec.More() {
		at := skipSeparators(data, dec.InputOffset())
		tok, err := dec.Token()
		if err != nil {
			return positioned(data, dec.InputOffset(), err)
		}
		key, _ := tok.(string)
		switch {
		case !knownKeys[key]:
			return fmt.Errorf("capabilities: %s: unknown field %q", lineCol(data, at), key)
		case seen[key]:
			return fmt.Errorf("capabilities: %s: duplicate field %q", lineCol(data, at), key)
		}
		seen[key] = true
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return positioned(data, dec.InputOffset(), err)
		}
	}
	if _, err := dec.Token(); err != nil {
		return positioned(data, dec.InputOffset(), err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("capabilities: %s: unexpected data after the top-level object", lineCol(data, dec.InputOffset()))
	}
	if !seen["version"] {
		return errors.New(`capabilities: missing "version"`)
	}
	return nil
}

// positioned prefixes a decode error with its line and column. Syntax and
// type errors carry their own offset; anything else uses fallback.
func positioned(data []byte, fallback int64, err error) error {
	offset := fallback
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxErr):
		offset = syntaxErr.Offset
	case errors.As(err, &typeErr):
		offset = typeErr.Offset
	case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, io.EOF):
		offset = int64(len(data))
		err = errors.New("unexpected end of file")
	}
	return fmt.Errorf("capabilities: %s: %w", lineCol(data, offset), err)
}

// skipSeparators advances offset past whitespace and commas, to the start of
// the next key.
func skipSeparators(data []byte, offset int64) int64 {
	for offset < int64(len(data)) {
		switch data[offset] {
		case ' ', '\t', '\r', '\n', ',':
			offset++
		default:
			return offset
		}
	}
	return offset
}

// lineCol renders a byte offset into data as a 1-based "line:column".
func lineCol(data []byte, offset int64) string {
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	line, col := 1, 1
	for _, b := range data[:offset] {
		if b == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}
	return fmt.Sprintf("%d:%d", line, col)
}

// Load reads and parses the capabilities file at path. An empty path means no
// file was given and returns nil, which grants everything. Every error is
// prefixed with the path.
func Load(path string) (*Capabilities, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: capabilities: %w", path, err)
	}
	c, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

// AllowsBuiltin reports whether the grant lets a pack call the builtin name.
func (c *Capabilities) AllowsBuiltin(name string) bool {
	if c == nil || c.Builtins == nil {
		return true
	}
	return slices.Contains(*c.Builtins, name)
}

// AllowsModule reports whether the grant lets a pack load the builtin module
// key ("@sentrie/hash").
func (c *Capabilities) AllowsModule(key string) bool {
	if c == nil || c.Modules == nil {
		return true
	}
	return slices.Contains(*c.Modules, key)
}

// AllowsLocalModules reports whether the grant lets a pack load its own
// TypeScript or JavaScript files.
func (c *Capabilities) AllowsLocalModules() bool {
	if c == nil || c.LocalModules == nil {
		return true
	}
	return *c.LocalModules
}
