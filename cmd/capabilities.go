// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/index"
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

// loadAndCheckCapabilities loads the capabilities file at path (none when
// empty) and checks the validated index against it, so a pack that uses
// something outside the grant fails at load time with the offending spans.
// The returned grant is handed to the executor for the runtime gates.
func loadAndCheckCapabilities(ctx context.Context, path string, idx *index.Index) (*capabilities.Capabilities, error) {
	grant, err := capabilities.Load(path)
	if err != nil {
		return nil, err
	}
	if err := idx.CheckCapabilities(ctx, grant); err != nil {
		return nil, err
	}
	return grant, nil
}
