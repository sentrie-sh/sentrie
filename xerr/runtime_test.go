// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeCategoryErrorsWrapCategoryAndMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		is   error
		msg  string
	}{
		{name: "injected", err: ErrInjected("boom %d", 7), is: InjectedError{}, msg: "boom 7"},
		{name: "invalid invocation", err: ErrInvalidInvocation("missing argument"), is: InvalidInvocationError{}, msg: "missing argument"},
		{name: "unresolvable fact", err: ErrUnresolvableFact("user"), is: InvalidInvocationError{}, msg: "unresolvable fact: user"},
		{name: "required fact", err: ErrRequiredFact("org"), is: InvalidInvocationError{}, msg: "required fact not found: org"},
		{name: "rule not found", err: ErrRuleNotFound("ns/pol/r"), is: NotFoundError{}, msg: "rule: ns/pol/r"},
		{name: "policy not found", err: ErrPolicyNotFound("ns/pol"), is: NotFoundError{}, msg: "policy: ns/pol"},
		{name: "namespace not found", err: ErrNamespaceNotFound("ns"), is: NotFoundError{}, msg: "namespace: ns"},
		{name: "shape not found", err: ErrShapeNotFound("ns/shape"), is: NotFoundError{}, msg: "shape: ns/shape"},
		{name: "not exported", err: ErrNotExported("ns/pol/r"), is: NotExportedError{}, msg: "ns/pol/r"},
		{name: "import resolution", err: ErrImportResolution("mod", "fn"), is: ImportResolutionError{}, msg: "module: mod, fn: fn"},
		{name: "shape validation", err: ErrShapeValidation("invalid"), is: ShapeValidationError{}, msg: "invalid"},
		{name: "module invocation", err: ErrModuleInvocation("mod", "fn"), is: ModuleInvocationError{}, msg: "module: mod, fn: fn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			require.ErrorIs(t, tt.err, tt.is)
			require.Contains(t, tt.err.Error(), tt.msg)
		})
	}
}

func TestRuntimeCategoryErrorEmptyMessageFallsBackToCategory(t *testing.T) {
	err := ErrInvalidInvocation("")
	require.Equal(t, InvalidInvocationError{}.Error(), err.Error())
	require.True(t, errors.Is(err, InvalidInvocationError{}))
}
