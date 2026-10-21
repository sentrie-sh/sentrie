// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"

	"github.com/sentrie-sh/sentrie/capabilities"
)

// WithCapabilities installs the operator's grant on the executor and its
// JavaScript registry. Builtin dispatch then refuses a builtin the grant does
// not list, and module resolution refuses an ungranted `@sentrie/*` module or,
// when local modules are not granted, any pack-local file. These are the
// runtime halves of index.CheckCapabilities, which callers run at load time
// so a pack fails before it serves. A nil grant allows everything.
func WithCapabilities(grant *capabilities.Capabilities) NewExecutorOption {
	return func(e *executorImpl) {
		e.capabilities = grant
		e.jsRegistry.SetCapabilities(grant)
	}
}

// grant returns the executor's capabilities, tolerating a nil executor.
func (e *executorImpl) grant() *capabilities.Capabilities {
	if e == nil {
		return nil
	}
	return e.capabilities
}

// CheckModuleGrants is the load-time check for TypeScript: for every `use` in
// the executor's index it compiles the module and checks each static
// require("...") in its transpiled code, and transitively in every pack-local
// module it requires, against the grant (js.Registry.CheckUseRequires). A
// dynamic require(expr) cannot be seen statically and is still refused when it
// runs. It returns nil when no grant is installed or exec is not this
// package's executor.
func CheckModuleGrants(exec Executor) error {
	e, ok := exec.(*executorImpl)
	if !ok || e.capabilities == nil {
		return nil
	}
	var errs []error
	for _, nsName := range slices.Sorted(maps.Keys(e.index.Namespaces)) {
		ns := e.index.Namespaces[nsName]
		for _, polName := range slices.Sorted(maps.Keys(ns.Policies)) {
			p := ns.Policies[polName]
			fileDir, err := filepath.Abs(filepath.Dir(p.FilePath))
			if err != nil {
				return err
			}
			for _, alias := range slices.Sorted(maps.Keys(p.Uses)) {
				use := p.Uses[alias]
				if err := e.jsRegistry.CheckUseRequires(use.RelativeFrom, use.LibFrom, fileDir); err != nil {
					errs = append(errs, fmt.Errorf("%s: use %s in %s: %w", use.Span(), alias, p.FQN, err))
				}
			}
		}
	}
	return errors.Join(errs...)
}
