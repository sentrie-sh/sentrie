// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"time"

	"github.com/dop251/goja"
)

var BuiltinTimeGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// Constants
	_ = ex.Set("RFC3339", time.RFC3339)
	_ = ex.Set("RFC3339Nano", time.RFC3339Nano)
	_ = ex.Set("RFC1123", time.RFC1123)
	_ = ex.Set("RFC1123Z", time.RFC1123Z)
	_ = ex.Set("RFC822", time.RFC822)
	_ = ex.Set("RFC822Z", time.RFC822Z)

	_ = ex.Set("now", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			return throwErr(vm, errors.New("now requires exactly 0 arguments"))
		}

		// The pinned execution start, in Unix seconds. A VM without it is a
		// programming error, never a reason to read the wall clock (#153).
		now, err := executionStartSeconds(vm)
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(now)
	})

	_ = ex.Set("parse", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("parse requires exactly 1 argument"))
		}
		timeStr := call.Argument(0).String()

		// Try RFC3339 first (most common)
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			// Try RFC3339Nano
			t, err = time.Parse(time.RFC3339Nano, timeStr)
			if err != nil {
				return throwErr(vm, err)
			}
		}

		return vm.ToValue(t.Unix())
	})

	_ = ex.Set("format", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("format requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		formatStr := call.Argument(1).String()

		t := time.Unix(timestamp, 0)
		formatted := t.Format(formatStr)

		return vm.ToValue(formatted)
	})

	_ = ex.Set("isBefore", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("isBefore requires exactly 2 arguments"))
		}
		ts1 := int64(call.Argument(0).ToFloat())
		ts2 := int64(call.Argument(1).ToFloat())
		return vm.ToValue(ts1 < ts2)
	})

	_ = ex.Set("isAfter", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("isAfter requires exactly 2 arguments"))
		}
		ts1 := int64(call.Argument(0).ToFloat())
		ts2 := int64(call.Argument(1).ToFloat())
		return vm.ToValue(ts1 > ts2)
	})

	_ = ex.Set("isBetween", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return throwErr(vm, errors.New("isBetween requires exactly 3 arguments"))
		}
		ts := int64(call.Argument(0).ToFloat())
		start := int64(call.Argument(1).ToFloat())
		end := int64(call.Argument(2).ToFloat())
		return vm.ToValue(ts >= start && ts <= end)
	})

	_ = ex.Set("addDuration", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("addDuration requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		durationStr := call.Argument(1).String()

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return throwErr(vm, err)
		}

		t := time.Unix(timestamp, 0)
		result := t.Add(duration)

		return vm.ToValue(result.Unix())
	})

	_ = ex.Set("subtractDuration", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("subtractDuration requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		durationStr := call.Argument(1).String()

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return throwErr(vm, err)
		}

		t := time.Unix(timestamp, 0)
		result := t.Add(-duration)

		return vm.ToValue(result.Unix())
	})

	_ = ex.Set("unix", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("unix requires exactly 1 argument"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		return vm.ToValue(time.Unix(timestamp, 0).Unix())
	})

	return ex, nil
}
