// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"
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

	for _, shape := range traversal {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := shape.resolveDependency(idx, nil); err != nil {
			return err
		}
	}

	return nil
}
