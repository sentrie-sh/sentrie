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

// var BuiltinStringsGo = func(vm *goja.Runtime) (*goja.Object, error) {
// 	ex := vm.NewObject()

// 	// Helpers produce goja.Callable that honor JS strings
// 	set := func(name string, fn func(a string) (any, error)) error {
// 		return ex.Set(name, func(call goja.FunctionCall) goja.Value {
// 			a := call.Argument(0).ToString().String()
// 			out, err := fn(a)
// 			if err != nil {
// 				panic(vm.ToValue(err))
// 			}
// 			return vm.ToValue(out)
// 		})
// 	}

// 	_ = set("len", func(a string) (any, error) { return len(a), nil })

// 	return ex, nil
// }
