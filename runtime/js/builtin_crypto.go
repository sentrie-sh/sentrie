// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/dop251/goja"
)

var BuiltinCryptoGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// sha256 returns the lowercase hex digest of its input, the same encoding
	// as @sentrie/hash.sha256.
	_ = ex.Set("sha256", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.NewGoError(errors.New("sha256 requires exactly 1 argument")))
		}
		sum := sha256.Sum256([]byte(call.Argument(0).String()))
		return vm.ToValue(hex.EncodeToString(sum[:]))
	})

	return ex, nil
}
