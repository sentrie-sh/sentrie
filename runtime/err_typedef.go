package runtime

import (
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

var (
	ErrTypeRef           = errors.New("typeref error")
	errConstraintFailed  = errors.Wrapf(ErrTypeRef, "constraint failed")
	errUnknownConstraint = errors.Wrapf(ErrTypeRef, "unknown constraint")
)

func ErrUnknownConstraint(c *ast.TypeRefConstraint) error {
	return errors.Wrapf(errUnknownConstraint, "unknown constraint: '%s' at %s", c.Name, c.Range)
}

func ErrConstraintFailed(pos tokens.Range, c *ast.TypeRefConstraint, err error) error {
	return errors.Wrapf(errConstraintFailed, "constraint failed: '%s' at %s", c.Name, pos)
}

func IsUnknownConstraint(err error) bool {
	return errors.Is(err, errUnknownConstraint)
}

func IsConstraintFailed(err error) bool {
	return errors.Is(err, errConstraintFailed)
}
