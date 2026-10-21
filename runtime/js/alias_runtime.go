// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"context"
	"errors"
	"sync"

	"github.com/dop251/goja"
)

// AliasRuntime hosts one clean VM for one loaded copy of a module.
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

// InterruptOnDone interrupts vm with ctx.Err() if ctx ends before the returned
// stop function is called. stop waits for an interrupt already in flight, so
// nothing touches vm after stop returns. An interrupt that fired is left set:
// the context is over, so nothing more should run on this VM, and an outer
// frame sharing the VM must see it too.
func InterruptOnDone(ctx context.Context, vm *goja.Runtime) (stop func()) {
	if ctx == nil {
		return func() {}
	}
	fired := make(chan struct{})
	stopAfter := context.AfterFunc(ctx, func() {
		defer close(fired)
		vm.Interrupt(ctx.Err())
	})
	return func() {
		if !stopAfter() {
			<-fired
		}
	}
}

// installInterrupt wires context cancellation into goja's interrupt mechanism
// for the duration of one module load.
func (ar *AliasRuntime) installInterrupt(ctx context.Context) (stop func()) {
	return InterruptOnDone(ctx, ar.VM)
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
