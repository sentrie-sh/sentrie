package index

import (
	"context"

	"github.com/pkg/errors"
)

func (idx *Index) Commit(ctx context.Context) error {
	idx.commitOnce.Do(func() {
		idx.commitError = idx.commit(ctx)
		idx.commitError = errors.Wrapf(idx.commitError, "commit error")
		idx.committed = 1
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
