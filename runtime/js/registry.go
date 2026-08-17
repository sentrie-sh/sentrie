// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/constants"
)

type SourceMap struct {
	Version  int
	Sources  []string
	Mappings string
	Names    []string
}

type ModuleSpec struct {
	Key      string        // canonical key used by registry (e.g., @sentra/math or /abs/path/to/mod.ts)
	Path     string        // filesystem path if not builtin
	Dir      string        // base dir for resolving relative requires
	Builtin  bool          // this is a builtin module
	SourceTS string        // original TS/JS (for builtins or disk)
	Program  *goja.Program // compiled IIFE function returning factory (require,module,exports)=>void

	TranspiledCode string
	TranspiledMap  *SourceMap

	BuiltInProvider ModuleProvider // if non-nil, this module is native Go-backed
	once            sync.Once
	err             error
}

type Registry struct {
	PackRoot string

	goBuiltins map[string]ModuleProvider // name -> Go module provider
	tsBuiltins map[string]string         // name -> TypeScript source

	modsMu sync.RWMutex
	mods   map[string]*ModuleSpec
}

func NewRegistry(packRoot string) *Registry {
	return &Registry{
		PackRoot:   packRoot,
		goBuiltins: map[string]ModuleProvider{},
		tsBuiltins: map[string]string{},
		mods:       map[string]*ModuleSpec{},
	}
}

func (r *Registry) RegisterGoBuiltin(name string, provider ModuleProvider) {
	r.goBuiltins[fmt.Sprintf("@%s/%s", constants.APPNAME, name)] = provider
}

func (r *Registry) RegisterTSBuiltin(name, tsSource string) {
	r.tsBuiltins[fmt.Sprintf("@%s/%s", constants.APPNAME, name)] = tsSource
}

// Resolve a "use" style reference into a canonical registry key + filesystem path.
func (r *Registry) resolveUse(localFrom string, libFrom []string, fileDir string) (key, path, dir string, builtin bool, err error) {
	if len(libFrom) > 0 {
		switch libFrom[0] {
		case constants.APPNAME:
			key = "@" + constants.APPNAME + "/" + filepath.ToSlash(filepath.Join(libFrom[1:]...))
			return key, "", "", true, nil
		case "local":
			key, path, dir, err = r.relativeToLocal(r.PackRoot, filepath.Join(libFrom[1:]...))
			return key, path, dir, false, err
		default:
			// we should be able to resolve a @vendor/lib/sublib style reference - later on where vendor libs are installed in a known location
			return "", "", "", false, fmt.Errorf("unsupported library from: %v", libFrom)
		}
	}

	key, path, dir, err = r.relativeToLocal(fileDir, localFrom)
	if err != nil {
		return "", "", "", false, err
	}
	return key, path, dir, false, nil

}

func isOutsidePackRoot(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func evalSymlinksExisting(path string) (string, error) {
	path = filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	dir := filepath.Dir(path)
	if dir == path {
		return path, nil
	}

	resolvedDir, err := evalSymlinksExisting(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedDir, filepath.Base(path)), nil
}

func (r *Registry) confineToPackRoot(path string) (string, error) {
	packRoot := filepath.Clean(r.PackRoot)
	path = filepath.Clean(path)

	packRootRelative, err := filepath.Rel(packRoot, path)
	if err != nil {
		return "", err
	}
	if isOutsidePackRoot(packRootRelative) {
		return "", fmt.Errorf("relative path is outside the packroot: %s", packRootRelative)
	}

	resolvedPackRoot, err := evalSymlinksExisting(packRoot)
	if err != nil {
		return "", err
	}
	resolvedPath, err := evalSymlinksExisting(path)
	if err != nil {
		return "", err
	}

	packRootRelative, err = filepath.Rel(resolvedPackRoot, resolvedPath)
	if err != nil {
		return "", err
	}
	if isOutsidePackRoot(packRootRelative) {
		return "", fmt.Errorf("relative path is outside the packroot: %s", packRootRelative)
	}

	return resolvedPath, nil
}

func (r *Registry) relativeToLocal(fromDir, spec string) (key, path, dir string, err error) {
	path, err = r.confineToPackRoot(filepath.Join(fromDir, spec))
	if err != nil {
		return "", "", "", err
	}

	resolvedPackRoot, err := evalSymlinksExisting(filepath.Clean(r.PackRoot))
	if err != nil {
		return "", "", "", err
	}
	packRootRelative, err := filepath.Rel(resolvedPackRoot, path)
	if err != nil {
		return "", "", "", err
	}

	key = "@local/" + filepath.ToSlash(packRootRelative)
	return key, path, filepath.Dir(path), nil
}

// Resolve a require() from within a module file.
func (r *Registry) resolveRequire(fromDir, spec string) (key, path, dir string, builtin bool, err error) {
	if strings.HasPrefix(spec, "@"+constants.APPNAME+"/") {
		return spec, "", "", true, nil
	}

	if strings.HasPrefix(spec, "@local/") {
		key, path, dir, err := r.relativeToLocal(r.PackRoot, spec[len("@local/"):])
		return key, path, dir, false, err
	}

	if strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") {
		localKey, localPath, localDir, err := r.relativeToLocal(fromDir, spec)
		if err != nil {
			return "", "", "", false, err
		}
		return localKey, localPath, localDir, false, nil
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
		if gp, ok := r.goBuiltins[key]; ok {
			m.BuiltInProvider = gp
		} else if tsSource, ok := r.tsBuiltins[key]; ok {
			// fallback to TS builtin if no Go provider
			m.SourceTS = tsSource
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

	if m.Path == "" && m.BuiltInProvider == nil && m.SourceTS == "" {
		return nil
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

	if mod == nil {
		return nil, fmt.Errorf("module %s not found", key)
	}

	// Warm compile best-effort
	_, err = r.programFor(mod)
	return mod, err
}

// programFor ensures the module is compiled to a goja.Program returning a factory function.
func (r *Registry) programFor(m *ModuleSpec) (*goja.Program, error) {
	if m.BuiltInProvider != nil {
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
		m.TranspiledCode = out.Code
		tm := SourceMap{}
		err = json.Unmarshal([]byte(out.Map), &tm)
		if err != nil {
			m.err = err
			return
		}
		m.TranspiledMap = &tm
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
	if mod == nil {
		return nil, fmt.Errorf("module %s not found", key)
	}
	_, err = r.programFor(mod)
	return mod, err
}
