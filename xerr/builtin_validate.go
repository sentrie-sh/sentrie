// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// BuiltinArgKindError is a validate-time builtin argument kind mismatch.
type BuiltinArgKindError struct {
	At      tokens.Range
	Message string
}

func (e BuiltinArgKindError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrIndex.Error())
}

func (e BuiltinArgKindError) Unwrap() error { return ErrIndex }

// ErrBuiltinArgKind reports a span-anchored builtin argument kind violation.
func ErrBuiltinArgKind(at tokens.Range, message string) error {
	return BuiltinArgKindError{At: at, Message: message}
}

// BuiltinCallableArityError is a validate-time HOF callable arity mismatch.
type BuiltinCallableArityError struct {
	At      tokens.Range
	Message string
}

func (e BuiltinCallableArityError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrIndex.Error())
}

func (e BuiltinCallableArityError) Unwrap() error { return ErrIndex }

// ErrBuiltinCallableArity reports a span-anchored builtin callable arity violation.
func ErrBuiltinCallableArity(at tokens.Range, message string) error {
	return BuiltinCallableArityError{At: at, Message: message}
}

// BuiltinCallArityError is a validate-time builtin call arity mismatch.
type BuiltinCallArityError struct {
	At      tokens.Range
	Message string
}

func (e BuiltinCallArityError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrIndex.Error())
}

func (e BuiltinCallArityError) Unwrap() error { return ErrIndex }

// ErrBuiltinCallArity reports a span-anchored builtin call arity violation.
func ErrBuiltinCallArity(at tokens.Range, message string) error {
	return BuiltinCallArityError{At: at, Message: message}
}
