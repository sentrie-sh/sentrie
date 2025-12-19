// SPDX-License-Identifier: Apache-2.0

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
	"context"
	"errors"
	"sync"

	"github.com/dop251/goja"
)

// AliasRuntime hosts one clean VM per `use ... as alias` binding.
// It installs a CJS-like require() and caches module exports by resolved key.
type AliasRuntime struct {
	VM      *goja.Runtime
	Reg     *Registry
	BaseDir string // directory of the primary "use" module (for resolving relative requires)
	cacheMu sync.Mutex
	cache   map[string]*goja.Object // key -> module.exports object
}

func NewAliasRuntime(reg *Registry, baseDir string) *AliasRuntime {
	rt := goja.New()
	return &AliasRuntime{
		VM:      rt,
		Reg:     reg,
		BaseDir: baseDir,
		cache:   map[string]*goja.Object{},
	}
}

// installInterrupt wires context cancellation into goja's interrupt mechanism.
func (ar *AliasRuntime) installInterrupt(ctx context.Context) (stop func()) {
	if ctx == nil {
		return func() {}
	}
	done := make(chan struct{})
	ar.VM.ClearInterrupt() // clear previous
	go func() {
		select {
		case <-ctx.Done():
			ar.VM.Interrupt(ctx.Err())
		case <-done:
		}
	}()
	return func() { close(done); ar.VM.ClearInterrupt() }
}

// Require implements CommonJS require(spec) with caching.
func (ar *AliasRuntime) Require(ctx context.Context, fromDir, spec string) (*goja.Object, error) {
	// Resolve module
	mod, err := ar.Reg.LoadRequire(fromDir, spec)
	if err != nil {
		return nil, err
	}

	key := mod.KeyOrPath()

	// Go-native path: fabricate exports directly, no program execution
	if mod.BuiltInProvider != nil {
		ex, gerr := mod.BuiltInProvider(ar.VM)
		if gerr != nil {
			ar.cacheMu.Lock()
			delete(ar.cache, key)
			ar.cacheMu.Unlock()
			return nil, gerr
		}
		// Cache real exports and return
		ar.cacheMu.Lock()
		ar.cache[key] = ex
		ar.cacheMu.Unlock()
		return ex, nil
	}

	// Cache check
	ar.cacheMu.Lock()
	if ex, ok := ar.cache[key]; ok {
		ar.cacheMu.Unlock()
		return ex, nil
	}
	// Prepare container early for circular deps
	moduleObj := ar.VM.NewObject()
	exportsObj := ar.VM.NewObject()
	_ = moduleObj.Set("exports", exportsObj)

	// Place placeholder in cache (for circular dependency support)
	ar.cache[key] = exportsObj
	ar.cacheMu.Unlock()

	// Compile & run program to obtain the module factory
	pgm, err := ar.Reg.programFor(mod)
	if err != nil {
		// cleanup on failure
		ar.cacheMu.Lock()
		delete(ar.cache, key)
		ar.cacheMu.Unlock()
		return nil, err
	}

	stop := ar.installInterrupt(ctx)
	defer stop()

	// Install per-module bound require and clean it up afterwards
	prevRequire := ar.VM.Get("__require")
	childRequire := func(call goja.FunctionCall) goja.Value {
		childSpec := call.Argument(0).String()
		ex, err := ar.Require(ctx, mod.Dir, childSpec)
		if err != nil {
			// throw a proper Go error to JS
			panic(ar.VM.NewGoError(err))
		}
		return ex // return as goja.Value
	}
	_ = ar.VM.Set("__require", childRequire)
	defer func() { _ = ar.VM.Set("__require", prevRequire) }()

	fnVal, err := ar.VM.RunProgram(pgm)
	if err != nil {
		ar.cacheMu.Lock()
		delete(ar.cache, key)
		ar.cacheMu.Unlock()
		return nil, err
	}
	fn, ok := goja.AssertFunction(fnVal)
	if !ok {
		ar.cacheMu.Lock()
		delete(ar.cache, key)
		ar.cacheMu.Unlock()
		return nil, errors.New("module did not evaluate to a function")
	}

	// Execute the module factory: (require, module, exports) => { ... }
	if _, err = fn(fnVal, ar.VM.Get("__require"), moduleObj, exportsObj); err != nil {
		ar.cacheMu.Lock()
		delete(ar.cache, key)
		ar.cacheMu.Unlock()
		return nil, err
	}

	// IMPORTANT: capture final module.exports (factory may have reassigned it)
	finalVal := moduleObj.Get("exports")
	finalObj := finalVal.ToObject(ar.VM)

	// Update cache with the final exports (not the initial placeholder)
	ar.cacheMu.Lock()
	ar.cache[key] = finalObj
	ar.cacheMu.Unlock()

	return finalObj, nil
}
