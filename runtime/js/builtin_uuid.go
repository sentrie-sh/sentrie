// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
			return throwErr(vm, err)
		}
		return vm.ToValue(v6.String())
	})

	_ = ex.Set("v7", func(call goja.FunctionCall) goja.Value {
		v7, err := uuid.NewV7()
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(v7.String())
	})

	return ex, nil
}
