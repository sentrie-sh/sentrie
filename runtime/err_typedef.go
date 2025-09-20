package runtime

import (
	"github.com/binaek/sentra/ast"
	"github.com/pkg/errors"
)

var (
	ErrTypeRef           = errors.New("typeref error")
	errConstraintFailed  = errors.Wrapf(ErrTypeRef, "constraint failed")
	errUnknownConstraint = errors.Wrapf(ErrTypeRef, "unknown constraint")
)

func ErrUnknownConstraint(c *ast.TypeRefConstraint) error {
	return errors.Wrapf(errUnknownConstraint, "unknown constraint: '%s' at %s", c.Name, c.Pos)
}

func ErrConstraintFailed(expr ast.Expression, c *ast.TypeRefConstraint, err error) error {
	return errors.Wrapf(errConstraintFailed, "constraint failed: '%s' at %s", c.Name, expr.Position())
}

func IsUnknownConstraint(err error) bool {
	return errors.Is(err, errUnknownConstraint)
}

func IsConstraintFailed(err error) bool {
	return errors.Is(err, errConstraintFailed)
}
