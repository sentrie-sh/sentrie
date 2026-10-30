// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/tokens"
)

var (
	ErrTypeRef           = errors.New("typeref error")
	errConstraintFailed  = fmt.Errorf("constraint failed: %w", ErrTypeRef)
	errUnknownConstraint = fmt.Errorf("unknown constraint: %w", ErrTypeRef)
)

func ErrUnknownConstraint(c *ast.TypeRefConstraint) error {
	return fmt.Errorf("unknown constraint: '%s' at %s: %w", c.Name, c.Span(), errUnknownConstraint)
}

// ErrConstraintFailed reports a constraint whose checker failed. It is a type
// mismatch (see isTypeMismatch) unless the checker blamed the constraint's own
// arguments (constraints.ErrInvalidArgument): a broken constraint is an error
// in the policy, not a value that does not conform.
func ErrConstraintFailed(pos tokens.Range, c *ast.TypeRefConstraint, err error) error {
	if err != nil {
		return &typeRefFailure{
			err:      fmt.Errorf("constraint failed: '%s' at %s: %w", c.Name, c.Span(), errors.Join(errConstraintFailed, err)),
			mismatch: !errors.Is(err, constraints.ErrInvalidArgument),
		}
	}
	return &typeRefFailure{err: fmt.Errorf("constraint failed: '%s' at %s: %w", c.Name, c.Span(), errConstraintFailed), mismatch: true}
}

// typeRefFailure marks why validateValueAgainstTypeRef failed, without
// changing the wrapped message. A mismatch says the value does not conform to
// the type. A failure that is not a mismatch says the check itself could not
// be completed, such as a constraint argument that failed to evaluate; it
// shields any mismatch raised inside that evaluation. Errors with no marker
// (an unknown shape or constraint, an unexported shape) are not mismatches.
type typeRefFailure struct {
	err      error
	mismatch bool
}

func (f *typeRefFailure) Error() string { return f.err.Error() }
func (f *typeRefFailure) Unwrap() error { return f.err }

// errMismatch builds a type mismatch error.
func errMismatch(format string, args ...any) error {
	return &typeRefFailure{err: fmt.Errorf(format, args...), mismatch: true}
}

// errCheckFailed marks err as a failure to complete a type check.
func errCheckFailed(err error) error {
	return &typeRefFailure{err: err}
}

// isTypeMismatch reports whether a validateValueAgainstTypeRef error means
// only that the value does not conform. The outermost marker in the chain
// decides, so a nested field or item mismatch is a mismatch of the whole
// value, and a mismatch inside a failed constraint argument is not.
// `x is T` turns a mismatch into false and propagates any other error.
func isTypeMismatch(err error) bool {
	var f *typeRefFailure
	return errors.As(err, &f) && f.mismatch
}

func IsUnknownConstraint(err error) bool {
	return errors.Is(err, errUnknownConstraint)
}

func IsConstraintFailed(err error) bool {
	return errors.Is(err, errConstraintFailed)
}
