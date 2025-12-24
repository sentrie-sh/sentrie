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

import (
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

var BuiltinUuidGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("v4", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(uuid.New().String())
	})

	_ = ex.Set("v6", func(call goja.FunctionCall) goja.Value {
		v6, err := uuid.NewV6()
		if err != nil {
			return vm.NewGoError(err)
		}
		return vm.ToValue(v6.String())
	})

	_ = ex.Set("v7", func(call goja.FunctionCall) goja.Value {
		v7, err := uuid.NewV7()
		if err != nil {
			return vm.NewGoError(err)
		}
		return vm.ToValue(v7.String())
	})

	return ex, nil
}
