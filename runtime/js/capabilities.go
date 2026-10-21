// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/xerr"
)

// SetCapabilities installs the operator's grant. PrepareUse and LoadRequire
// then refuse a builtin module the grant does not list and, when local modules
// are not granted, every pack-local file, so require() inside TypeScript is
// held to the same grant as a `use` statement. A nil grant allows everything.
//
// Like RegisterGoBuiltin, it must be called during executor construction,
// before the registry is shared between goroutines.
func (r *Registry) SetCapabilities(grant *capabilities.Capabilities) {
	r.grant = grant
}

// checkGrant is the runtime half of index.CheckCapabilities for modules.
func (r *Registry) checkGrant(key string, builtin bool) error {
	if builtin {
		if !r.grant.AllowsModule(key) {
			return fmt.Errorf("module %q: %w", key, xerr.ErrNotGranted)
		}
		return nil
	}
	if !r.grant.AllowsLocalModules() {
		return fmt.Errorf("local module %q: %w", key, xerr.ErrNotGranted)
	}
	return nil
}

// requireCall matches a static require("...") in transpiled code. esbuild
// emits every import as require("...") with a double-quoted specifier.
var requireCall = regexp.MustCompile(`\brequire\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)

// CheckUseRequires resolves a use statement's module (which checks the module
// itself against the grant) and, for pack-local code, checks every static
// require("...") in its transpiled source, following pack-local requires
// transitively, against the grant. It is the load-time counterpart of the
// gate in LoadRequire. A nil grant returns nil.
func (r *Registry) CheckUseRequires(localFrom string, libFrom []string, fileDir string) error {
	if r.grant == nil {
		return nil
	}
	mod, err := r.PrepareUse(localFrom, libFrom, fileDir)
	if err != nil {
		return err
	}
	return r.checkStaticRequires(mod, map[string]bool{})
}

func (r *Registry) checkStaticRequires(m *ModuleSpec, seen map[string]bool) error {
	if m.Builtin || seen[m.KeyOrPath()] {
		return nil
	}
	seen[m.KeyOrPath()] = true
	if _, err := r.programFor(m); err != nil {
		return err
	}
	for _, match := range requireCall.FindAllStringSubmatch(m.TranspiledCode, -1) {
		spec, err := strconv.Unquote(`"` + match[1] + `"`)
		if err != nil {
			continue
		}
		child, err := r.LoadRequire(m.Dir, spec)
		if err != nil {
			return fmt.Errorf("%s: require(%q): %w", m.KeyOrPath(), spec, err)
		}
		if err := r.checkStaticRequires(child, seen); err != nil {
			return err
		}
	}
	return nil
}
