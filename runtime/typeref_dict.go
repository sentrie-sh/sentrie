// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstDictTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.DictTypeRef, pos tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	entries, ok := v.DictValue()
	if !ok {
		return errMismatch("value %v is not a dict", v)
	}

	// check every value against the element type, in key order so the
	// reported key is the same on every run. An entry holding undefined
	// counts as absent and is skipped: undefined is not a value (#148).
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		if entries[key].IsUndefined() {
			continue
		}
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, entries[key], typeRef.ValueType, pos); err != nil {
			return fmt.Errorf("value for key %q is not valid at %s: %w", key, pos, err)
		}
	}

	return checkConstraints(ctx, v, bound, pos)
}
