// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/sentrie-sh/sentrie/capabilities"

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
