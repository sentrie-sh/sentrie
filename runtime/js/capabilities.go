// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"

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
