// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/fatih/structs"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime/js"
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
	closed  bool // set by close; guarded by probeMu
}

// newJSModule loads the module once, which both verifies that it loads and
// gives the module its probe instance.
func newJSModule(ctx context.Context, key string, load instanceLoader) (*jsModule, error) {
	probe, err := load(ctx, time.Now())
	if err != nil {
		return nil, err
	}
	return &jsModule{
		key:   key,
		load:  load,
		slots: make(chan struct{}, maxLiveInstances),
		probe: probe,
	}, nil
}

// errModuleClosed is returned by acquire after the owning executor closed.
var errModuleClosed = errors.New("module is closed")

// close drops the probe and makes later acquisitions fail. Calls already
// running are unaffected: each holds its own VM.
func (m *jsModule) close() {
	m.probeMu.Lock()
	defer m.probeMu.Unlock()
	m.closed = true
	m.probe = nil
}

// acquire loads a fresh, never-used instance of the module for one call. The
// caller must call release once it is done with the instance.
func (m *jsModule) acquire(ctx context.Context, start time.Time) (_ *JSInstance, release func(), _ error) {
	m.probeMu.Lock()
	closed := m.closed
	m.probeMu.Unlock()
	if closed {
		return nil, nil, fmt.Errorf("%s: %w", m.key, errModuleClosed)
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

// checkExports verifies that the module provides every name in exports, so a
// `use` naming a missing export fails when it is bound rather than when called.
func (m *jsModule) checkExports(exports []string) error {
	if len(exports) == 0 {
		return nil
	}
	m.probeMu.Lock()
	defer m.probeMu.Unlock()
	if m.closed {
		return fmt.Errorf("%s: %w", m.key, errModuleClosed)
	}
	for _, name := range exports {
		if m.probe.exports.Get(name) == nil {
			return fmt.Errorf("module %s missing required export %q", m.key, name)
		}
	}
	return nil
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

func normalizeBoundaryForJS(v any) any {
	if box.IsBoundaryUndefined(v) {
		return goja.Undefined()
	}
	switch t := v.(type) {
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, normalizeBoundaryForJS(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, item := range t {
			out[k] = normalizeBoundaryForJS(item)
		}
		return out
	default:
		return v
	}
}

func (m ModuleBinding) Call(ctx context.Context, ec *ExecutionContext, fn string, args ...any) (any, error) {
	if m.module == nil {
		return nil, fmt.Errorf("module has no JS binding")
	}
	instance, release, err := m.module.acquire(ctx, ec.CreatedAt())
	if err != nil {
		return nil, err
	}
	defer release()

	val, ok := m.lookupExport(instance, fn)
	if !ok {
		return nil, fmt.Errorf("function %s not found in module %q", fn, m.Alias)
	}
	fnc, ok := goja.AssertFunction(val)
	if !ok {
		return nil, fmt.Errorf("export '%q' is not callable", fn)
	}

	// Interrupt the call if ctx ends. stop waits for an in-flight interrupt,
	// so no goroutine touches the VM once Call returns.
	stop := js.InterruptOnDone(ctx, instance.rt)
	defer stop()

	ga := make([]goja.Value, 0, len(args))
	for _, a := range args {
		ga = append(ga, instance.rt.ToValue(normalizeBoundaryForJS(a)))
	}
	out, err := fnc(goja.Undefined(), ga...)
	if err != nil {
		return nil, err
	}

	if goja.IsUndefined(out) {
		return box.ToBoundaryAny(box.Undefined()), nil
	}
	if goja.IsNull(out) {
		return nil, nil
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
		return nil, fmt.Errorf("unexpected return type %s", out.ExportType().String())
	}

	result := out.Export()

	// if it's a struct, convert to a map[string]any
	if structs.IsStruct(result) {
		result = structs.Map(result)
	}

	return result, nil
}
