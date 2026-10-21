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

// Parse decodes and checks a capabilities file. Unknown fields, a missing or
// unsupported version, trailing data, empty names, and module entries without
// the "@sentrie/" prefix are errors. Names are not checked against the
// builtins and modules this build ships, so a file written for a newer build
// still loads: a name this build does not know grants nothing.
func Parse(data []byte) (*Capabilities, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var c Capabilities
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("capabilities: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("capabilities: unexpected data after the top-level object")
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

// Load reads and parses the capabilities file at path. An empty path means no
// file was given and returns nil, which grants everything.
func Load(path string) (*Capabilities, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("capabilities: %w", err)
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
