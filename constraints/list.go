// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/box"
)

var ListContraintCheckers map[string]ConstraintDefinition = map[string]ConstraintDefinition{
	"not_empty": define(ConstraintDefinition{
		Name:    "not_empty",
		NumArgs: 0,
		Checker: func(ctx context.Context, val box.Value, args []box.Value) error {
			if val.Kind() != box.ValueList {
				return fmt.Errorf("expected list, got %s", val.Kind())
			}
			if lst, ok := val.ListValue(); ok && len(lst) == 0 {
				return fmt.Errorf("list is empty - expected non-empty list")
			}
			return nil
		},
	}),
}
