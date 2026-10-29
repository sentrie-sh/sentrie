// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"context"

	"github.com/sentrie-sh/sentrie/pack"
)

// SetupStdLib installs the standard library into the VM, and the
// deterministic Date and Math.random shims when the registry's capabilities
// grant enables them (see installDeterministicShims).
// This MUST be called before Require is used or the VM is invoked.
func (ar *AliasRuntime) SetupStdLib(ctx context.Context, pack *pack.PackFile) error {
	if err := ar.setupEnvStdLib(ctx, pack); err != nil {
		return err
	}
	if ar.Reg != nil && ar.Reg.grant.DeterministicJS() {
		if err := installDeterministicShims(ar.VM); err != nil {
			return err
		}
	}
	return nil
}
