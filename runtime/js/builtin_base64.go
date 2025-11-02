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
	"encoding/base64"

	"github.com/dop251/goja"
)

var BuiltinBase64Go = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("encode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("decode", func(call goja.FunctionCall) goja.Value {
		decoded, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return vm.NewGoError(err)
		}
		return vm.ToValue(string(decoded))
	})

	_ = ex.Set("urlEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(base64.URLEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("urlDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := base64.URLEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return vm.NewGoError(err)
		}
		return vm.ToValue(string(decoded))
	})

	return ex, nil
}
