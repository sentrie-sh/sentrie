// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/dop251/goja"
	"github.com/fatih/structs"
	"github.com/jackc/puddle/v2"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constants"
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

// jsModule is a loaded module and its VM pool, shared by every `use` of the
// same resolved path. It carries no export restriction.
type jsModule struct {
	key  string
	pool *puddle.Pool[*JSInstance]
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

// checkExports verifies that the module provides every name in exports, so a
// `use` naming a missing export fails when it is bound rather than when called.
func (m *jsModule) checkExports(ctx context.Context, exports []string) error {
	if len(exports) == 0 {
		return nil
	}
	res, err := m.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer res.Release()
	for _, name := range exports {
		if res.Value().exports.Get(name) == nil {
			return fmt.Errorf("module %s missing required export %q", m.key, name)
		}
	}
	return nil
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
	if m.module == nil || m.module.pool == nil {
		return nil, fmt.Errorf("module has no JS binding")
	}
	poolInstance, err := m.module.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer poolInstance.Release()

	instance := poolInstance.Value()
	if err := instance.rt.Set(constants.ExecutionStartTimeUnixKey, ec.CreatedAt().UTC().Unix()); err != nil {
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

	// Install an interrupt to honor context cancel
	done := make(chan struct{})
	if ctx != nil {
		instance.rt.ClearInterrupt()
		go func() {
			select {
			case <-ctx.Done():
				instance.rt.Interrupt(ctx.Err())
			case <-done:
				// clear the interrupt
				instance.rt.ClearInterrupt()
			}
		}()
		defer close(done)
	}

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
