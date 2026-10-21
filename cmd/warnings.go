// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/sentrie-sh/sentrie/index"
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
