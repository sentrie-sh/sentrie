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
	"encoding/json"
	"errors"

	"github.com/dop251/goja"
)

var BuiltinJsonGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("isValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isValid requires exactly 1 argument"))
		}

		jsonStr := call.Argument(0).String()

		var result interface{}
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			return vm.ToValue(false)
		}

		return vm.ToValue(true)
	})

	return ex, nil
}
