// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import "github.com/dop251/goja"

// throwErr raises err as a JavaScript exception from inside a Go-backed
// builtin. It never returns.
//
// goja treats a value returned from a func(goja.FunctionCall) goja.Value as a
// normal result, so returning vm.NewGoError(err) hands the caller a truthy
// Error object and a failed check reads as a success (#111). Every builtin
// failure goes through this helper instead, written as
// `return throwErr(vm, err)`; the goja.Value result type exists only so call
// sites keep their return statement.
//
// The error conventions every builtin follows:
//   - argument-count and argument-type errors throw;
//   - predicates (isValid, is*, jwt.verify) answer false for input they cannot
//     parse;
//   - a genuine "no result" answer is null;
//   - every other failure throws.
func throwErr(vm *goja.Runtime, err error) goja.Value {
	panic(vm.NewGoError(err))
}
