// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

type ConstraintChecker func(ctx context.Context, p *index.Policy, val box.Value, args []box.Value) error

type ConstraintDefinition struct {
	Name    string
	NumArgs int
	Checker ConstraintChecker
}

// ErrInvalidArgument marks a checker failure caused by the constraint's own
// arguments (wrong count, wrong kind, an invalid regexp, a zero divisor)
// rather than by the value under test. The runtime propagates such a failure
// as an error instead of reading it as "the value does not conform", so a
// broken constraint never makes `x is not T` true.
var ErrInvalidArgument = errors.New("invalid constraint argument")

// invalidArgumentError keeps the checker's message and matches
// ErrInvalidArgument under errors.Is.
type invalidArgumentError struct {
	err error
}

func (e *invalidArgumentError) Error() string   { return e.err.Error() }
func (e *invalidArgumentError) Unwrap() []error { return []error{e.err, ErrInvalidArgument} }

// invalidArgf builds an argument failure with an unchanged message.
func invalidArgf(format string, args ...any) error {
	return &invalidArgumentError{err: fmt.Errorf(format, args...)}
}
