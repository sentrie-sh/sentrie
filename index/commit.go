// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"
	"slices"
)

func (idx *Index) Commit(ctx context.Context) error {
	idx.commitOnce.Do(func() {
		idx.commitError = idx.commit(ctx)
		if idx.commitError != nil {
			idx.commitError = fmt.Errorf("commit error: %w", idx.commitError)
		}
		idx.committed = 1
	})
	return idx.commitError
}

func (idx *Index) commit(ctx context.Context) error {
	traversal, err := idx.shapeDag.TopoSort()
	if err != nil {
		return err
	}

	// The shape graph has an edge from a shape to the shape it composes with,
	// and TopoSort lists an edge's source before its destination. That puts a
	// shape ahead of its parent, but a parent's fields must be complete before
	// the shape copies them, so hydrate in the reverse order: parents first.
	for _, shape := range slices.Backward(traversal) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := shape.resolveDependency(idx, nil); err != nil {
			return err
		}
	}

	return nil
}
