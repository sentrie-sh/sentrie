// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime/js"
	"github.com/sentrie-sh/sentrie/trinary"
)

// JSInstance is one loaded copy of a module, in its own VM.
type JSInstance struct {
	rt *goja.Runtime
	// exports is the module's whole `module.exports` object. Which names a
	// caller may reach is decided per `use` by the ModuleBinding, not here.
	exports *goja.Object
	// keys holds the own enumerable export names, captured at load. A `use`
	// that names no exports reaches exactly these.
	keys map[string]struct{}
}

func newJSInstance(rt *goja.Runtime, exports *goja.Object) *JSInstance {
	keys := map[string]struct{}{}
	for _, k := range exports.Keys() {
		keys[k] = struct{}{}
	}
	return &JSInstance{rt: rt, exports: exports, keys: keys}
}

// maxLiveInstances bounds how many VMs one module may have running at once.
// Calls beyond it wait for a slot, so contention shows as latency.
const maxLiveInstances = 10

// instanceLoader builds a fresh VM and loads the module into it, with the
// module's notion of "now" pinned to start.
type instanceLoader func(ctx context.Context, start time.Time) (*JSInstance, error)

// jsModule is a loaded module, shared by every `use` of the same resolved path.
// It carries no export restriction.
//
// Every call runs in a VM that no other call has touched (#105): acquire loads
// the module into a fresh goja.Runtime, reusing the compiled goja.Program the
// registry holds, and the VM is dropped after the call. Module-level state such
// as a counter or a memo table therefore lives for exactly one call, whatever
// the evaluation order or concurrency.
type jsModule struct {
	key   string
	load  instanceLoader
	slots chan struct{}

	// probe is one loaded instance kept to answer export checks at bind time.
	// It never serves a call. probeMu serialises access, since a goja.Runtime
	// is single-threaded and reading an export may run a getter.
	probeMu sync.Mutex
	probe   *JSInstance

	// closed is set by close. It is atomic so that close and acquire never
	// wait on probeMu, which an export check holds while it runs module code.
	closed atomic.Bool
}

// newJSModule loads the module once, which both verifies that it loads and
// gives the module its probe instance.
func newJSModule(ctx context.Context, key string, load instanceLoader) (*jsModule, error) {
	probe, err := load(ctx, time.Now())
	if err != nil {
		return nil, err
	}
	// The load succeeded, but ctx may have ended in the window between the
	// module factory returning and the load's interrupt being stopped, which
	// leaves the interrupt set (js.InterruptOnDone keeps a fired interrupt).
	// The probe outlives this call and runs export getters later, so clear it;
	// the load's stop has returned, so nothing else touches the VM now.
	probe.rt.ClearInterrupt()
	return &jsModule{
		key:   key,
		load:  load,
		slots: make(chan struct{}, maxLiveInstances),
		probe: probe,
	}, nil
}

// close makes later acquisitions and export checks fail with
// ErrExecutorClosed. Calls already running are unaffected: each holds its own
// VM. It never blocks, so the executor can close a module whose export check
// is stuck in module code.
func (m *jsModule) close() {
	m.closed.Store(true)
}

func (m *jsModule) closedErr() error {
	return fmt.Errorf("%s: %w", m.key, ErrExecutorClosed)
}

// acquire loads a fresh, never-used instance of the module for one call. The
// caller must call release once it is done with the instance.
func (m *jsModule) acquire(ctx context.Context, start time.Time) (_ *JSInstance, release func(), _ error) {
	if m.closed.Load() {
		return nil, nil, m.closedErr()
	}
	select {
	case m.slots <- struct{}{}:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	instance, err := m.load(ctx, start)
	if err != nil {
		<-m.slots
		return nil, nil, err
	}
	return instance, func() { <-m.slots }, nil
}

// recoverUncaughtJS turns a goja throw, interrupt or stack overflow that
// panicked out of Go-side property access (getters, Proxy traps) into *err.
// Other panics are re-raised. It must be deferred directly.
func recoverUncaughtJS(err *error) {
	r := recover()
	if r == nil {
		return
	}
	jsErr, ok := r.(error)
	if !ok || !isGojaThrow(jsErr) {
		panic(r)
	}
	*err = jsErr
}

// runGuardedJS runs f, which touches vm from Go and so may run module code,
// under ctx's interrupt and with uncaught JS turned into an error (#142).
func runGuardedJS(ctx context.Context, vm *goja.Runtime, f func() error) (err error) {
	stop := js.InterruptOnDone(ctx, vm)
	defer stop()
	defer recoverUncaughtJS(&err)
	return f()
}

// checkExports verifies that the module provides every name in exports, so a
// `use` naming a missing export fails when it is bound rather than when called.
// Reading an export may run a getter, so it runs under ctx's deadline.
func (m *jsModule) checkExports(ctx context.Context, exports []string) error {
	if len(exports) == 0 {
		return nil
	}
	if m.closed.Load() {
		return m.closedErr()
	}
	m.probeMu.Lock()
	defer m.probeMu.Unlock()
	if m.closed.Load() {
		return m.closedErr()
	}
	err := runGuardedJS(ctx, m.probe.rt, func() error {
		for _, name := range exports {
			if m.probe.exports.Get(name) == nil {
				return fmt.Errorf("module %s missing required export %q", m.key, name)
			}
		}
		return nil
	})
	// The probe outlives this check: never leave a fired interrupt on it. The
	// guard's stop has returned, and probeMu is held, so nothing else runs.
	m.probe.rt.ClearInterrupt()
	return err
}

// ModuleBinding is one `use` statement's view of a shared jsModule: the alias
// it was bound under and the exports it asked for.
type ModuleBinding struct {
	CanonicalKey string
	Alias        string
	// exports lists the names the `use` statement requested. Empty means the
	// module's own enumerable exports.
	exports []string
	module  *jsModule
}

// lookupExport applies this binding's export projection to a loaded instance.
func (m ModuleBinding) lookupExport(instance *JSInstance, name string) (goja.Value, bool) {
	if len(m.exports) > 0 {
		if !slices.Contains(m.exports, name) {
			return nil, false
		}
	} else if _, ok := instance.keys[name]; !ok {
		return nil, false
	}
	v := instance.exports.Get(name)
	return v, v != nil
}

// normalizeBoundaryForJS maps a value produced by box.TryToBoundaryAny to the Go
// value goja should convert for a JavaScript argument. Booleans cross as real
// JavaScript booleans and trinary.Unknown crosses as undefined, so JavaScript
// never sees a wrapped trinary object whose truthiness disagrees with Sentrie.
func normalizeBoundaryForJS(v any) (any, error) {
	if box.IsBoundaryUndefined(v) {
		return goja.Undefined(), nil
	}
	switch t := v.(type) {
	case nil, bool, float64, string:
		return t, nil
	case trinary.Value:
		switch t {
		case trinary.True:
			return true, nil
		case trinary.False:
			return false, nil
		case trinary.Unknown:
			return goja.Undefined(), nil
		default:
			return nil, fmt.Errorf("trinary value %d cannot cross into JavaScript", int(t))
		}
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			x, err := normalizeBoundaryForJS(item)
			if err != nil {
				return nil, err
			}
			out = append(out, x)
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			x, err := normalizeBoundaryForJS(item)
			if err != nil {
				return nil, err
			}
			out[k] = x
		}
		return out, nil
	default:
		// A host document (box.ValueDocument) keeps goja's reflection wrapping.
		return t, nil
	}
}

func (m ModuleBinding) Call(ctx context.Context, ec *ExecutionContext, fn string, args ...any) (_ any, err error) {
	if m.module == nil {
		return nil, fmt.Errorf("module has no JS binding")
	}
	instance, release, err := m.module.acquire(ctx, ec.CreatedAt())
	if err != nil {
		return nil, err
	}
	defer release()

	// Interrupt the call if ctx ends, from the export lookup (which may run a
	// getter) through the return conversion. stop waits for an in-flight
	// interrupt, so no goroutine touches the VM once Call returns.
	stop := js.InterruptOnDone(ctx, instance.rt)
	defer stop()
	defer recoverUncaughtJS(&err)

	if err := js.BeginCall(instance.rt, ec.CreatedAt(), m.CanonicalKey, fn); err != nil {
		return nil, err
	}

	val, ok := m.lookupExport(instance, fn)
	if !ok {
		return nil, fmt.Errorf("function %s not found in module %q", fn, m.Alias)
	}
	fnc, ok := goja.AssertFunction(val)
	if !ok {
		return nil, fmt.Errorf("export '%q' is not callable", fn)
	}

	ga := make([]goja.Value, 0, len(args))
	for _, a := range args {
		x, err := normalizeBoundaryForJS(a)
		if err != nil {
			return nil, err
		}
		ga = append(ga, instance.rt.ToValue(x))
	}
	out, err := fnc(goja.Undefined(), ga...)
	if err != nil {
		return nil, err
	}

	return jsToSentrie(ctx, instance.rt, out)
}
