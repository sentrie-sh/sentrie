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
			return vm.NewGoError(errors.New("md5 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := md5.Sum([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha1", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sha1 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha1.Sum([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha256", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sha256 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha256.Sum256([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("sha512", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sha512 requires exactly 1 argument"))
		}
		data := call.Argument(0).String()
		hash := sha512.Sum512([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	_ = ex.Set("hmac", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return vm.NewGoError(errors.New("hmac requires exactly 3 arguments"))
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
			return vm.NewGoError(errors.New("unsupported algorithm: " + algorithm))
		}

		h.Write([]byte(data))
		hashBytes := h.Sum(nil)
		return vm.ToValue(hex.EncodeToString(hashBytes))
	})

	return ex, nil
}

