package constraints

import (
	"context"

	"github.com/sentrie-sh/sentrie/index"
)

type ConstraintChecker[T any] func(ctx context.Context, p *index.Policy, val T, args []any) error

type ConstraintDefinition[T any] struct {
	Name    string
	NumArgs int
	Checker ConstraintChecker[T]
}
