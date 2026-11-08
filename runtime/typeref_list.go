// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstListTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ListTypeRef, pos tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	items, ok := v.ListValue()
	if !ok {
		return errMismatch("value %v is not an array at %s - expected array", v, pos)
	}

	for _, item := range items {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, item, typeRef.ElemType, pos); err != nil {
			return fmt.Errorf("item is not valid at %s: %w", pos, err) // TODO: improve this error message
		}
	}

	return checkConstraints(ctx, v, bound, pos)
}
