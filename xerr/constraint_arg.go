// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// ConstraintArgError is a validate-time constraint whose own arguments are
// invalid (wrong count, wrong kind, an invalid regexp, a zero divisor). Err is
// the constraint's argument failure, which callers can match under errors.Is.
type ConstraintArgError struct {
	At   tokens.Range
	Name string
	Err  error
}

func (e ConstraintArgError) Error() string {
	return fmt.Sprintf("%s: invalid argument to constraint '%s': %s: %s", e.At.String(), e.Name, e.Err.Error(), ErrIndex.Error())
}

func (e ConstraintArgError) Unwrap() []error { return []error{ErrIndex, e.Err} }

// ErrConstraintArg reports a span-anchored invalid constraint argument.
func ErrConstraintArg(at tokens.Range, name string, err error) error {
	return ConstraintArgError{At: at, Name: name, Err: err}
}
