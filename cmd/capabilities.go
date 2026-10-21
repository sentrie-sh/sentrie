// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime"
)

const capabilitiesFlagName = "capabilities"

// capabilitiesFlag is the --capabilities flag shared by exec, serve and validate.
func capabilitiesFlag() cling.CmdFlag {
	return cling.
		NewStringCmdInput(capabilitiesFlagName).
		WithDefault("").
		WithDescription("Capabilities file limiting the builtins and modules the pack may use (default: allow everything)").
		AsFlag()
}

// newGrantedExecutor loads the capabilities file at path (none when empty),
// checks the validated index against it, builds the executor with the grant
// installed, and then checks every static require() in the pack's local
// TypeScript against the grant. A pack that uses something outside the grant
// therefore fails at load time, naming the offending span or module, rather
// than on the first request that reaches it. opts are passed to NewExecutor.
func newGrantedExecutor(ctx context.Context, idx *index.Index, path string, opts ...runtime.NewExecutorOption) (runtime.Executor, error) {
	grant, err := capabilities.Load(path)
	if err != nil {
		return nil, err
	}
	if err := idx.CheckCapabilities(ctx, grant); err != nil {
		return nil, err
	}
	exec, err := runtime.NewExecutor(idx, append(opts, runtime.WithCapabilities(grant))...)
	if err != nil {
		return nil, err
	}
	if err := runtime.CheckModuleGrants(exec); err != nil {
		return nil, err
	}
	return exec, nil
}
