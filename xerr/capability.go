// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"errors"
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// ErrNotGranted is the sentinel for a builtin or module the operator's
// capabilities file does not grant. Static (index-time) and runtime denials
// both wrap it.
var ErrNotGranted = errors.New("not granted by the capabilities file")

// CapabilityError is a span-anchored, index-time capability denial. It wraps
// both ErrNotGranted and ErrIndex.
type CapabilityError struct {
	At      tokens.Range
	Message string
}

func (e CapabilityError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrNotGranted.Error())
}

func (e CapabilityError) Unwrap() []error { return []error{ErrNotGranted, ErrIndex} }

// ErrCapability reports a span-anchored capability denial.
func ErrCapability(at tokens.Range, message string) error {
	return CapabilityError{At: at, Message: message}
}
