// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"

	"github.com/dop251/goja"
)

var BuiltinHashGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("md5", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("md5 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := md5.Sum([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha1", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("sha1 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha1.Sum([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha256", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("sha256 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha256.Sum256([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha512", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("sha512 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha512.Sum512([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("hmac", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return throwErr(vm, errors.New("hmac requires exactly 3 arguments"))
		}
		algorithm := call.Argument(0).String()
		data := call.Argument(1).String()
		key := call.Argument(2).String()

		var h hash.Hash
		switch algorithm {
		case "md5":
			h = hmac.New(md5.New, []byte(key))
		case "sha1":
			h = hmac.New(sha1.New, []byte(key))
		case "sha256":
			h = hmac.New(sha256.New, []byte(key))
		case "sha384":
			h = hmac.New(sha512.New384, []byte(key))
		case "sha512":
			h = hmac.New(sha512.New, []byte(key))
		default:
			return throwErr(vm, errors.New("unsupported algorithm: "+algorithm))
		}

		h.Write([]byte(data))
		hashBytes := h.Sum(nil)
		return vm.ToValue(hex.EncodeToString(hashBytes))
	})

	return ex, nil
}
