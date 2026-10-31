// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// CallTargetError is a validate-time problem with a call's callee: an unknown
// function, a call to a value that is not callable, or a wrong argument count
// for a let-bound lambda.
type CallTargetError struct {
	At      tokens.Range
	Message string
}

func (e CallTargetError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrIndex.Error())
}

func (e CallTargetError) Unwrap() error { return ErrIndex }

// ErrCallTarget reports a span-anchored callee problem.
func ErrCallTarget(at tokens.Range, message string) error {
	return CallTargetError{At: at, Message: message}
}
