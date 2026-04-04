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

package js

import "github.com/dop251/goja"

func (s *JSTestSuite) TestRegisterTSBuiltin() {
	reg := NewRegistry("/tmp/test")
	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)
	key := "@sentrie/test"
	s.Contains(reg.tsBuiltins, key)
	s.Equal(tsSource, reg.tsBuiltins[key])
}

func (s *JSTestSuite) TestGetOrCreateModule_TSBuiltin() {
	reg := NewRegistry("/tmp/test")
	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)
	mod := reg.getOrCreateModule("@sentrie/test", "", "", true)
	s.Require().NotNil(mod)
	s.True(mod.Builtin)
	s.Equal(tsSource, mod.SourceTS)
	s.Nil(mod.BuiltInProvider)
}

func (s *JSTestSuite) TestGetOrCreateModule_GoProviderTakesPrecedence() {
	reg := NewRegistry("/tmp/test")
	goProvider := func(vm *goja.Runtime) (*goja.Object, error) {
		return vm.NewObject(), nil
	}
	reg.RegisterGoBuiltin("test", goProvider)
	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)
	mod := reg.getOrCreateModule("@sentrie/test", "", "", true)
	s.Require().NotNil(mod)
	s.NotNil(mod.BuiltInProvider)
	s.Equal("", mod.SourceTS)
}

func (s *JSTestSuite) TestPrepareUse_TSBuiltin() {
	reg := NewRegistry("/tmp/test")
	tsSource := `export const test = () => 42;`
	reg.RegisterTSBuiltin("test", tsSource)
	mod, err := reg.PrepareUse("", []string{"sentrie", "test"}, "/tmp/test")
	s.Require().NoError(err)
	s.Require().NotNil(mod)
	s.True(mod.Builtin)
	s.Equal(tsSource, mod.SourceTS)
}
