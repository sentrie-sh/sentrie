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

package runtime

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/dop251/goja"
	"github.com/fatih/structs"
	"github.com/jackc/puddle/v2"
	"github.com/sentrie-sh/sentrie/constants"
)

// JSInstance is a context-aware binding for an alias VM.
type JSInstance struct {
	rt      *goja.Runtime
	exports map[string]goja.Value
}

type ModuleBinding struct {
	CanonicalKey string
	Alias        string
	instancePool *puddle.Pool[*JSInstance]
}

func (m ModuleBinding) Call(ctx context.Context, ec *ExecutionContext, fn string, args ...any) (any, error) {
	if m.instancePool == nil {
		return nil, fmt.Errorf("module has no JS binding")
	}
	binding, err := m.instancePool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer binding.Release()

	vm := binding.Value()
	if err := vm.rt.Set(constants.ExecutionStartTimeUnixKey, ec.CreatedAt().UTC().Unix()); err != nil {
		return nil, err
	}

	val, ok := vm.exports[fn]
	if !ok {
		return nil, fmt.Errorf("function '%q' not found in module %q", fn, m.Alias)
	}
	fnc, ok := goja.AssertFunction(val)
	if !ok {
		return nil, fmt.Errorf("export '%q' is not callable", fn)
	}

	// Install an interrupt to honor context cancel
	done := make(chan struct{})
	if ctx != nil {
		vm.rt.ClearInterrupt()
		go func() {
			select {
			case <-ctx.Done():
				vm.rt.Interrupt(ctx.Err())
			case <-done:
				// clear the interrupt
				vm.rt.ClearInterrupt()
			}
		}()
		defer close(done)
	}

	ga := make([]goja.Value, 0, len(args))
	for _, a := range args {
		ga = append(ga, vm.rt.ToValue(a))
	}
	out, err := fnc(goja.Undefined(), ga...)
	if err != nil {
		return nil, err
	}

	acceptedReturnTypes := []reflect.Kind{
		reflect.Map,
		reflect.Slice,
		reflect.Array,
		reflect.String,
		reflect.Int64,
		reflect.Float64,
		reflect.Bool,
		reflect.Struct,
	}

	if !slices.Contains(acceptedReturnTypes, out.ExportType().Kind()) {
		return nil, fmt.Errorf("unexpected return type %T", out.ExportType())
	}

	result := out.Export()

	// if it's a struct, convert to a map[string]any
	if structs.IsStruct(result) {
		result = structs.Map(result)
	}

	return result, nil
}
