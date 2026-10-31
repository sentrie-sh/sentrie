// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstNumberTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.NumberTypeRef, pos tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	if _, ok := v.NumberValue(); !ok {
		return errMismatch("value %v is not a number", v)
	}

	return checkConstraints(ctx, v, bound, pos)
}
