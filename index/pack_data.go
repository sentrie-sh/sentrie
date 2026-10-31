// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/xerr"
)

// checkPackData verifies that every `data <name>` declaration names an entry of
// the pack manifest's `[data]` table that the loader has read. An index built
// without a pack has no data, so any declaration fails here rather than at
// evaluation. Iteration is sorted so the reported error is stable.
func (idx *Index) checkPackData(ctx context.Context) error {
	for _, nsName := range slices.Sorted(maps.Keys(idx.Namespaces)) {
		ns := idx.Namespaces[nsName]
		for _, polName := range slices.Sorted(maps.Keys(ns.Policies)) {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			p := ns.Policies[polName]
			for _, alias := range slices.Sorted(maps.Keys(p.Data)) {
				d := p.Data[alias]
				if idx.Pack == nil || idx.Pack.Data == nil {
					return fmt.Errorf("data %q at %s: the pack declares no [data] table: %w", d.Name, d.Span(), xerr.ErrIndex)
				}
				if _, ok := idx.Pack.Data[d.Name]; !ok {
					return fmt.Errorf("data %q at %s: no such entry in the pack's [data] table: %w", d.Name, d.Span(), xerr.ErrIndex)
				}
				if _, ok := idx.Pack.DataValues[d.Name]; !ok {
					return fmt.Errorf("data %q at %s: the entry was not loaded with the pack: %w", d.Name, d.Span(), xerr.ErrIndex)
				}
			}
		}
	}
	return nil
}
