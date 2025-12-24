// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xerr

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/tokens"
)

// Error injected by calling the `error` function in sentrie code
type InjectedError struct {
}

func (e InjectedError) Error() string { return "runtime error" }

func ErrInjected(format string, args ...any) error {
	return errors.Wrapf(InjectedError{}, format, args...)
}

type InfiniteRecursionError struct{ stack []string }

func (e InfiniteRecursionError) Error() string {
	return "infinite recursion: " + strings.Join(e.stack, " -> ")
}

type ConflictError struct {
	what        string
	where, with tokens.Range // where the conflict is, with what
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s at %s with %s", e.what, e.where.String(), e.with.String())
}

func ErrConflict(what string, where, with tokens.Range) error {
	return ConflictError{what: what, where: where, with: with}
}

type InvalidTypeError struct{ got, expected string }

func (e InvalidTypeError) Error() string {
	return "invalid type: " + e.got + " -> expected: " + e.expected
}

func ErrInvalidType(got, expected string) error {
	return InvalidTypeError{got: got, expected: expected}
}

func ErrInfiniteRecursion(stack []string) error {
	return InfiniteRecursionError{stack: stack}
}

type InvalidInvocationError struct{}

func (e InvalidInvocationError) Error() string {
	return "invalid invocation"
}

func ErrInvalidInvocation(reason string) error {
	return errors.Wrap(InvalidInvocationError{}, reason)
}

func ErrUnresolvableFact(name string) error {
	return errors.Wrapf(InvalidInvocationError{}, "unresolvable fact: %s", name)
}

func ErrRequiredFact(name string) error {
	return errors.Wrapf(InvalidInvocationError{}, "required fact not found: %s", name)
}

func ErrRuleNotFound(fqn string) error {
	return errors.Wrapf(NotFoundError{}, "rule: %s", fqn)
}

func ErrPolicyNotFound(fqn string) error {
	return errors.Wrapf(NotFoundError{}, "policy: %s", fqn)
}

func ErrNamespaceNotFound(name string) error {
	return errors.Wrapf(NotFoundError{}, "namespace: %s", name)
}

func ErrShapeNotFound(name string) error {
	return errors.Wrapf(NotFoundError{}, "shape: %s", name)
}

func ErrNotExported(fqn string) error {
	return errors.Wrap(NotExportedError{}, fqn)
}

func ErrImportResolution(module, fn string) error {
	return errors.Wrapf(ImportResolutionError{}, "module: %s, fn: %s", module, fn)
}

func ErrShapeValidation(msg string) error {
	return errors.Wrap(ShapeValidationError{}, msg)
}

func ErrModuleInvocation(module, fn string) error {
	return errors.Wrapf(ModuleInvocationError{}, "module: %s, fn: %s", module, fn)
}

var ErrRuntimePanic = &RuntimePanic{}

type NotFoundError struct{}

func (e NotFoundError) Error() string {
	return "not found"
}

type NotExportedError struct{}

func (e NotExportedError) Error() string { return "rule is not exported" }

type ImportResolutionError struct{}

func (e ImportResolutionError) Error() string {
	return "unable to resolve import"
}

type ModuleInvocationError struct{}

func (e ModuleInvocationError) Error() string {
	return "invoke module function failed"
}

type ShapeValidationError struct{}

func (e ShapeValidationError) Error() string { return "shape validation failed" }

type RuntimePanic struct{}

func (e RuntimePanic) Error() string { return "runtime panic" }
