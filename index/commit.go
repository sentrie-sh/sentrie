// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
