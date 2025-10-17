package constraints

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/index"
)

var ListContraintCheckers map[string]ConstraintDefinition[[]any] = map[string]ConstraintDefinition[[]any]{
	"not_empty": {
		Name:    "not_empty",
		NumArgs: 0,
		Checker: func(ctx context.Context, p *index.Policy, val []any, args []any) error {
			if len(val) == 0 {
				return fmt.Errorf("list is empty - expected non-empty list")
			}
			return nil
		},
	},
}
