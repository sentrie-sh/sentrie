// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
			return throwErr(vm, err)
		}
		return vm.ToValue(string(decoded))
	})

	_ = ex.Set("urlEncode", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(base64.URLEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("urlDecode", func(call goja.FunctionCall) goja.Value {
		decoded, err := base64.URLEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(string(decoded))
	})

	return ex, nil
}
