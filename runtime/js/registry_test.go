// SPDX-License-Identifier: Apache-2.0

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

package js

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterTSBuiltin(t *testing.T) {
	reg := NewRegistry("/tmp/test")

	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)

	key := "@sentrie/test"
	assert.Contains(t, reg.tsBuiltins, key)
	assert.Equal(t, tsSource, reg.tsBuiltins[key])
}

func TestGetOrCreateModule_TSBuiltin(t *testing.T) {
	reg := NewRegistry("/tmp/test")

	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)

	mod := reg.getOrCreateModule("@sentrie/test", "", "", true)
	require.NotNil(t, mod)
	assert.True(t, mod.Builtin)
	assert.Equal(t, tsSource, mod.SourceTS)
	assert.Nil(t, mod.BuiltInProvider) // TS builtin should not have Go provider
}

func TestGetOrCreateModule_GoProviderTakesPrecedence(t *testing.T) {
	reg := NewRegistry("/tmp/test")

	// Register both Go and TS builtin
	goProvider := func(vm *goja.Runtime) (*goja.Object, error) {
		return vm.NewObject(), nil
	}
	reg.RegisterGoBuiltin("test", goProvider)

	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)

	mod := reg.getOrCreateModule("@sentrie/test", "", "", true)
	require.NotNil(t, mod)
	assert.NotNil(t, mod.BuiltInProvider) // Go provider should take precedence
	assert.Equal(t, "", mod.SourceTS)     // TS source should not be set when Go provider exists
}

func TestPrepareUse_TSBuiltin(t *testing.T) {
	reg := NewRegistry("/tmp/test")

	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)

	mod, err := reg.PrepareUse("", []string{"sentrie", "test"}, "/tmp/test")
	require.NoError(t, err)
	require.NotNil(t, mod)
	assert.True(t, mod.Builtin)
	assert.Equal(t, tsSource, mod.SourceTS)
}
