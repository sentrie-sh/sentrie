// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/sentrie-sh/sentrie/box"
)

// ConstraintChecker reports whether val satisfies the constraint. A checker
// built with define runs CheckArgs first, so it may assume args are valid.
type ConstraintChecker func(ctx context.Context, val box.Value, args []box.Value) error

// ArgCheck checks one argument of a constraint, independent of any value.
// A failure must be built with invalidArgf.
type ArgCheck func(name string, arg box.Value) error

type ConstraintDefinition struct {
	Name string
	// NumArgs is the exact number of arguments, or -1 for one or more.
	NumArgs int
	// Arg checks each argument, whatever its position. Nil when the
	// constraint takes no arguments.
	Arg ArgCheck
	// Args, when set, checks the arguments together once every argument has
	// passed Arg, for rules that span arguments (range: minimum <= maximum).
	Args    func(args []box.Value) error
	Checker ConstraintChecker
}

// CheckArgs reports whether args are valid arguments for the constraint: the
// count, each argument, and the rules that span arguments. It never looks at
// a value, so it is the check the index makes when a pack is validated and
// the runtime makes before the value's kind, and every failure matches
// ErrInvalidArgument.
func (d ConstraintDefinition) CheckArgs(args []box.Value) error {
	switch {
	case d.NumArgs == -1 && len(args) < 1:
		return invalidArgf("%s constraint requires at least 1 argument", d.Name)
	case d.NumArgs == 0 && len(args) != 0:
		return invalidArgf("%s constraint takes no arguments", d.Name)
	case d.NumArgs == 1 && len(args) != 1:
		return invalidArgf("%s constraint requires 1 argument", d.Name)
	case d.NumArgs > 1 && len(args) != d.NumArgs:
		return invalidArgf("%s constraint requires %d arguments", d.Name, d.NumArgs)
	}
	if d.Arg != nil {
		for _, arg := range args {
			if err := d.Arg(d.Name, arg); err != nil {
				return err
			}
		}
	}
	if d.Args != nil {
		return d.Args(args)
	}
	return nil
}

// define returns d with a Checker that runs CheckArgs before d's own check,
// so a broken argument is reported whatever the value, and d's check may
// assume valid arguments.
func define(d ConstraintDefinition) ConstraintDefinition {
	check := d.Checker
	d.Checker = func(ctx context.Context, val box.Value, args []box.Value) error {
		if err := d.CheckArgs(args); err != nil {
			return err
		}
		return check(ctx, val, args)
	}
	return d
}

// numberArg requires a number argument.
func numberArg(_ string, arg box.Value) error {
	if _, ok := arg.NumberValue(); !ok {
		return invalidArgf("expected number, got %s", arg.Kind())
	}
	return nil
}

// stringArg requires a string argument.
func stringArg(_ string, arg box.Value) error {
	if _, ok := arg.StringValue(); !ok {
		return invalidArgf("expected string, got %s", arg.Kind())
	}
	return nil
}

// trinaryArg requires a trinary argument.
func trinaryArg(name string, arg box.Value) error {
	if _, ok := arg.TrinaryValue(); !ok {
		return invalidArgf("%s constraint expects a boolean argument", name)
	}
	return nil
}

// regexpArg requires a string argument that compiles as a regexp.
func regexpArg(name string, arg box.Value) error {
	if err := stringArg(name, arg); err != nil {
		return err
	}
	pattern, _ := arg.StringValue()
	if _, err := regexp.Compile(pattern); err != nil {
		return invalidArgf("invalid regexp pattern: %v", err)
	}
	return nil
}

// divisorArg requires a non-zero number argument.
func divisorArg(name string, arg box.Value) error {
	if err := numberArg(name, arg); err != nil {
		return err
	}
	if n, _ := arg.NumberValue(); n == 0 {
		return invalidArgf("divisor cannot be zero")
	}
	return nil
}

// numberSetArg requires a number or a list of numbers.
func numberSetArg(_ string, arg box.Value) error {
	_, err := numberConstraintSet(arg)
	return err
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
