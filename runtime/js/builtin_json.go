// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
			return throwErr(vm, errors.New("isValid requires exactly 1 argument"))
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
