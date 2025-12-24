// SPDX-License-Identifier: Apache-2.0
//
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
	"errors"
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/constants"
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
			return vm.NewGoError(errors.New("now requires exactly 0 arguments"))
		}

		// Get execution timestamp from VM global if available
		timestampVal := vm.Get(constants.ExecutionStartTimeUnixKey)
		if timestampVal != nil && timestampVal != goja.Undefined() && timestampVal != goja.Null() {
			return timestampVal
		}

		// Fallback to current time if execution timestamp not set
		return vm.ToValue(time.Now().Unix())
	})

	_ = ex.Set("parse", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("parse requires exactly 1 argument"))
		}
		timeStr := call.Argument(0).String()

		// Try RFC3339 first (most common)
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			// Try RFC3339Nano
			t, err = time.Parse(time.RFC3339Nano, timeStr)
			if err != nil {
				return vm.NewGoError(err)
			}
		}

		return vm.ToValue(t.Unix())
	})

	_ = ex.Set("format", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("format requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		formatStr := call.Argument(1).String()

		t := time.Unix(timestamp, 0)
		formatted := t.Format(formatStr)

		return vm.ToValue(formatted)
	})

	_ = ex.Set("isBefore", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("isBefore requires exactly 2 arguments"))
		}
		ts1 := int64(call.Argument(0).ToFloat())
		ts2 := int64(call.Argument(1).ToFloat())
		return vm.ToValue(ts1 < ts2)
	})

	_ = ex.Set("isAfter", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("isAfter requires exactly 2 arguments"))
		}
		ts1 := int64(call.Argument(0).ToFloat())
		ts2 := int64(call.Argument(1).ToFloat())
		return vm.ToValue(ts1 > ts2)
	})

	_ = ex.Set("isBetween", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return vm.NewGoError(errors.New("isBetween requires exactly 3 arguments"))
		}
		ts := int64(call.Argument(0).ToFloat())
		start := int64(call.Argument(1).ToFloat())
		end := int64(call.Argument(2).ToFloat())
		return vm.ToValue(ts >= start && ts <= end)
	})

	_ = ex.Set("addDuration", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("addDuration requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		durationStr := call.Argument(1).String()

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		t := time.Unix(timestamp, 0)
		result := t.Add(duration)

		return vm.ToValue(result.Unix())
	})

	_ = ex.Set("subtractDuration", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("subtractDuration requires exactly 2 arguments"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		durationStr := call.Argument(1).String()

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		t := time.Unix(timestamp, 0)
		result := t.Add(-duration)

		return vm.ToValue(result.Unix())
	})

	_ = ex.Set("unix", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("unix requires exactly 1 argument"))
		}
		timestamp := int64(call.Argument(0).ToFloat())
		return vm.ToValue(time.Unix(timestamp, 0).Unix())
	})

	return ex, nil
}
