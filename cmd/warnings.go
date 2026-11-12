// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/runtime/js"
)

// printNondeterminismWarnings writes one "warning: <span>: <message>" line per
// use of a nondeterministic builtin or builtin-module export in the validated
// index. Warnings never fail validation.
func printNondeterminismWarnings(ctx context.Context, w io.Writer, idx *index.Index) error {
	warnings, err := idx.NondeterminismWarnings(ctx, js.IsNondeterministicExport)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(w, "warning: %s\n", warning.String()); err != nil {
			return err
		}
	}
	return nil
}

// printUnenforcedPermissionWarnings writes one "warning: ..." line for each of
// the fs_read and net pack permissions that is declared with a
// non-empty list. No module API consults either list today, so declaring them
// suggests a sandbox that does not exist. Warnings never fail the command.
func printUnenforcedPermissionWarnings(w io.Writer, p *pack.PackFile) error {
	if p == nil || p.Permissions == nil {
		return nil
	}
	for _, perm := range []struct {
		name   string
		values []string
	}{
		{"fs_read", p.Permissions.FSRead},
		{"net", p.Permissions.Net},
	} {
		if len(perm.values) == 0 {
			continue
		}
		if _, err := fmt.Fprintf(w, "warning: pack permission %q is declared but not enforced yet; it gates nothing in this release\n", perm.name); err != nil {
			return err
		}
	}
	return nil
}
