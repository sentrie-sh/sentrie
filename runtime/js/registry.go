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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/binaek/sentra/constants"
	"github.com/dop251/goja"
)

type ModuleSpec struct {
	Key      string        // canonical key used by registry (e.g., @sentra/math or /abs/path/to/mod.ts)
	Path     string        // filesystem path if not builtin
	Dir      string        // base dir for resolving relative requires
	Builtin  bool          // this is a builtin module
	SourceTS string        // original TS/JS (for builtins or disk)
	Program  *goja.Program // compiled IIFE function returning factory (require,module,exports)=>void

	GoProvider GoModuleProvider // if non-nil, this module is native Go-backed
	once       sync.Once
	err        error
}

type Registry struct {
	PackRoot string

	builtins   map[string]string           // name -> TS source
	gobuiltins map[string]GoModuleProvider // name -> Go module provider

	modsMu sync.RWMutex
	mods   map[string]*ModuleSpec
}

func NewRegistry(packRoot string) *Registry {
	return &Registry{
		PackRoot:   packRoot,
		builtins:   map[string]string{},
		gobuiltins: map[string]GoModuleProvider{},
		mods:       map[string]*ModuleSpec{},
	}
}

func (r *Registry) RegisterTSBuiltin(name, tsSource string) {
	r.builtins[fmt.Sprintf("@%s/%s", constants.APPNAME, name)] = tsSource
}

func (r *Registry) RegisterGoBuiltin(name string, provider GoModuleProvider) {
	r.gobuiltins[fmt.Sprintf("@%s/%s", constants.APPNAME, name)] = provider
}

// Resolve a "use" style reference into a canonical registry key + filesystem path.
func (r *Registry) resolveUse(localFrom string, libFrom []string, fileDir string) (key, path, dir string, builtin bool, err error) {
	if len(libFrom) > 0 {
		// we have a use statement with a `@vendor/lib/sublib` style reference
		switch libFrom[0] {
		case constants.APPNAME:
			key = "@" + constants.APPNAME + "/" + filepath.ToSlash(filepath.Join(libFrom[1:]...))
			return key, "", "", true, nil
		case "local":
			key = "@local/" + filepath.ToSlash(filepath.Join(libFrom[1:]...))
			path = filepath.Join(r.PackRoot, filepath.ToSlash(filepath.Join(libFrom[1:]...)))
			return key, path, filepath.Dir(path), false, nil
		default:
			// we should be able to resolve a @vendor/lib/sublib style reference - later on where vendor libs are installed in a known location
			return "", "", "", false, fmt.Errorf("unsupported library from: %v", libFrom)
		}
	}

	// treat localFrom as a file-relative anchor
	path = filepath.Join(fileDir, localFrom)
	return filepath.Clean(path), path, filepath.Dir(path), false, nil
}

// Resolve a require() from within a module file.
func (r *Registry) resolveRequire(fromDir, spec string) (key, path, dir string, builtin bool, err error) {
	if strings.HasPrefix(spec, "@"+constants.APPNAME+"/") {
		return spec, "", "", true, nil
	}

	if strings.HasPrefix(spec, "@local/") {
		return spec, filepath.Join(r.PackRoot, spec[len("@local/"):]), r.PackRoot, false, nil
	}

	if strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") {
		path = spec
		if !filepath.IsAbs(path) {
			path = filepath.Join(fromDir, spec)
		}
		// add default extension if missing
		if filepath.Ext(path) == "" {
			if _, statErr := os.Stat(path + ".ts"); statErr == nil {
				path = path + ".ts"
			} else if _, statErr2 := os.Stat(path + ".js"); statErr2 == nil {
				path = path + ".js"
			}
		}
		path = filepath.Clean(path)
		return path, path, filepath.Dir(path), false, nil
	}
	// bare spec (e.g. "leftpad") not supported yet; could add node_modules later
	return "", "", "", false, fmt.Errorf("unsupported require spec: %q", spec)
}

func (r *Registry) getOrCreateModule(key, path, dir string, builtin bool) *ModuleSpec {
	r.modsMu.RLock()
	if m := r.mods[key]; m != nil {
		r.modsMu.RUnlock()
		return m
	}
	r.modsMu.RUnlock()

	r.modsMu.Lock()
	defer r.modsMu.Unlock()
	if m := r.mods[key]; m != nil {
		return m
	}

	m := &ModuleSpec{
		Key:     key,
		Path:    path,
		Dir:     dir,
		Builtin: builtin,
	}
	if builtin {
		// prefer Go module provider over TS source
		if gp, ok := r.gobuiltins[key]; ok {
			m.GoProvider = gp
		}

		// fallback to TS source - if exists
		if src, ok := r.builtins[key]; ok && m.GoProvider == nil {
			m.SourceTS = src
		}
	} else {
		if filepath.Ext(path) == "" {
			if _, statErr := os.Stat(path + ".ts"); statErr == nil {
				path = path + ".ts"
			} else if _, statErr2 := os.Stat(path + ".js"); statErr2 == nil {
				path = path + ".js"
			}
		}
		m.Path = path
	}
	r.mods[key] = m
	return m
}

// PrepareUse compiles (or schedules lazy compilation) for a "use" reference.
func (r *Registry) PrepareUse(localFrom string, libFrom []string, fileDir string) (*ModuleSpec, error) {
	key, path, dir, builtin, err := r.resolveUse(localFrom, libFrom, fileDir)
	if err != nil {
		return nil, err
	}
	mod := r.getOrCreateModule(key, path, dir, builtin)

	// Warm compile best-effort
	_, err = r.programFor(mod)
	return mod, err
}

// programFor ensures the module is compiled to a goja.Program returning a factory function.
func (r *Registry) programFor(m *ModuleSpec) (*goja.Program, error) {
	if m.GoProvider != nil {
		// No JS program to run — Go provider will fabricate exports.
		return nil, nil
	}
	m.once.Do(func() {
		var raw string
		if m.Builtin {
			raw = m.SourceTS
			if raw == "" {
				m.err = fmt.Errorf("builtin not found: %s", m.Key)
				return
			}
		} else {
			b, err := os.ReadFile(m.Path)
			if err != nil {
				m.err = err
				return
			}
			raw = string(b)
		}

		out, err := TranspileTS(m, raw)
		if err != nil {
			m.err = err
			return
		}
		wrapped := WrapAsIIFE(out.Code)

		// Compile once to a reusable Program (returns function)
		pgm, cerr := goja.Compile(m.KeyOrPath(), wrapped, true)
		if cerr != nil {
			m.err = cerr
			return
		}
		m.Program = pgm
	})
	return m.Program, m.err
}

func (m *ModuleSpec) KeyOrPath() string {
	if m.Key != "" {
		return m.Key
	}
	return m.Path
}

// LoadRequire resolves & compiles a dependency of another module by spec.
func (r *Registry) LoadRequire(fromDir, spec string) (*ModuleSpec, error) {
	key, path, dir, builtin, err := r.resolveRequire(fromDir, spec)
	if err != nil {
		return nil, err
	}
	mod := r.getOrCreateModule(key, path, dir, builtin)
	_, err = r.programFor(mod)
	return mod, err
}
