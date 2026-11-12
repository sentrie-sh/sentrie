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
		err := idx.commit(ctx)
		if err != nil {
			err = fmt.Errorf("commit error: %w", err)
		}
		idx.stateMu.Lock()
		idx.commitError = err
		idx.committed = 1
		idx.stateMu.Unlock()
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
		if err := shape.resolveDependency(idx); err != nil {
			return err
		}
	}

	return nil
}
