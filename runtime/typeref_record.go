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

func validateAgainstRecordTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.RecordTypeRef, pos tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	var entries []box.Value
	if arr, ok := v.ListValue(); ok {
		entries = arr
	} else {
		return errMismatch("value %v is not a record", v) // TODO: improve this error message
	}

	if len(entries) != len(typeRef.Fields) {
		return errMismatch("fields length mismatch: %v", v) // TODO: improve this error message
	}

	for i, field := range typeRef.Fields {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, entries[i], field, pos); err != nil {
			return fmt.Errorf("%v is not a valid record field: %w", v, err)
		}
	}

	return checkConstraints(ctx, v, bound, pos)
}
