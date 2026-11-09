// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/constants"
)

type ModuleSpec struct {
	Key      string        // canonical key used by registry (e.g., @sentrie/math or /abs/path/to/mod.ts)
	Path     string        // filesystem path if not builtin
	Dir      string        // base dir for resolving relative requires
	Builtin  bool          // this is a builtin module
	SourceTS string        // original TS/JS (for builtins or disk)
	Program  *goja.Program // compiled IIFE function returning factory (require,module,exports)=>void; read only after programFor succeeds

	TranspiledCode string // the factory expression, with its inline source map; read only after programFor succeeds

	BuiltInProvider ModuleProvider // if non-nil, this module is native Go-backed

	// Compilation state. Only a success is kept for the process lifetime. A
	// failure is remembered with the source text that produced it, so an
	// unchanged broken module is not re-transpiled on every call, while an
	// edited one is compiled again.
	compiled     atomic.Pointer[goja.Program]
	compileMu    sync.Mutex // serializes compile attempts; guards the fields below and Program, TranspiledCode
	failedSource string
	failedErr    error
}

type Registry struct {
	PackRoot string

	goBuiltins map[string]ModuleProvider // name -> Go module provider
	tsBuiltins map[string]string         // name -> TypeScript source

	modsMu sync.RWMutex
	mods   map[string]*ModuleSpec

	// resolved memoises successful LoadRequire resolutions, keyed by
	// fromDir and spec. Every call loads its module into a fresh VM,
	// so without it each call would redo the path containment and symlink
	// checks on disk. A module's content is already fixed after its first
	// compile, so a resolution is equally stable.
	resolved sync.Map // string -> *ModuleSpec

	// programsCompiled counts goja programs compiled by programFor. Each module
	// compiles once per registry; every fresh VM reuses the cached Program.
	programsCompiled atomic.Int64

	grant *capabilities.Capabilities // nil grants every module; see SetCapabilities
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

// confineToPackRoot resolves path and the pack root through every existing
// symlink and refuses path unless it lies inside the root. It returns the
// resolved path and its OS-separated relative form under the resolved root.
//
// Containment is decided only after both sides are resolved. A nested
// require joins onto the requiring module's Dir, which is already resolved,
// so comparing it lexically against an unresolved root (a pack under a
// symlinked deploy directory, or macOS /var -> /private/var) would refuse a
// valid pack (#242). path is cleaned before resolution, so a ".." never
// survives into the resolved form. The resolved path is read, except a
// probed .ts/.js, which relativeToLocal confines and then reads by its own
// name.
func (r *Registry) confineToPackRoot(path string) (resolvedPath, packRootRelative string, err error) {
	resolvedPackRoot, err := evalSymlinksExisting(r.PackRoot)
	if err != nil {
		return "", "", err
	}
	resolvedPath, err = evalSymlinksExisting(path)
	if err != nil {
		return "", "", err
	}

	packRootRelative, err = filepath.Rel(resolvedPackRoot, resolvedPath)
	if err != nil {
		return "", "", err
	}
	if isOutsidePackRoot(packRootRelative) {
		return "", "", fmt.Errorf("relative path is outside the packroot: %s", packRootRelative)
	}

	return resolvedPath, packRootRelative, nil
}

// moduleExtensions are tried, in order, for a local specifier without an
// extension.
var moduleExtensions = []string{".ts", ".js"}

func (r *Registry) relativeToLocal(fromDir, spec string) (key, path, dir string, err error) {
	path, packRootRelative, err := r.confineToPackRoot(filepath.Join(fromDir, spec))
	if err != nil {
		return "", "", "", err
	}

	key = "@local/" + filepath.ToSlash(packRootRelative)
	dir = filepath.Dir(path)

	// Probe extensions here rather than after resolution: the probed file can
	// itself be a symlink, so it is confined like any other path. Path keeps
	// the probed name (not the link target) so SourceName and Dir are those
	// of the specifier the pack wrote.
	if filepath.Ext(path) == "" {
		for _, ext := range moduleExtensions {
			candidate := path + ext
			if _, statErr := os.Stat(candidate); statErr != nil {
				continue
			}
			if _, _, err := r.confineToPackRoot(candidate); err != nil {
				return "", "", "", err
			}
			path = candidate
			break
		}
	}

	return key, path, dir, nil
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
		// relativeToLocal has already probed and confined the extension.
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
	if err := r.checkGrant(key, builtin); err != nil {
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

// CompiledPrograms reports how many module programs this registry has compiled.
func (r *Registry) CompiledPrograms() int64 {
	return r.programsCompiled.Load()
}

// programFor ensures the module is compiled to a goja.Program returning a factory function.
// A successful compile is reused for the process lifetime. A failure is returned again
// only while the source is unchanged; a read error is never cached.
func (r *Registry) programFor(m *ModuleSpec) (*goja.Program, error) {
	if m.BuiltInProvider != nil {
		// No JS program to run - Go provider will fabricate exports.
		return nil, nil
	}
	if pgm := m.compiled.Load(); pgm != nil {
		return pgm, nil
	}

	m.compileMu.Lock()
	defer m.compileMu.Unlock()
	if pgm := m.compiled.Load(); pgm != nil {
		return pgm, nil
	}

	raw, err := m.readSource()
	if err != nil {
		return nil, err
	}
	if m.failedErr != nil && raw == m.failedSource {
		return nil, m.failedErr
	}

	code, pgm, err := compileModule(m, raw)
	if err != nil {
		m.failedSource, m.failedErr = raw, err
		return nil, err
	}
	r.programsCompiled.Add(1)
	m.failedSource, m.failedErr = "", nil
	m.TranspiledCode = code
	m.Program = pgm
	m.compiled.Store(pgm) // publishes Program and TranspiledCode to lock-free readers
	return pgm, nil
}

func (m *ModuleSpec) readSource() (string, error) {
	if m.Builtin {
		if m.SourceTS == "" {
			return "", fmt.Errorf("builtin not found: %s", m.Key)
		}
		return m.SourceTS, nil
	}
	b, err := os.ReadFile(m.Path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// compileModule transpiles and compiles a module's source. goja reads the
// inline source map, so stack frames and compile errors name the original file.
func compileModule(m *ModuleSpec, raw string) (string, *goja.Program, error) {
	out, err := TranspileTS(m, raw)
	if err != nil {
		return "", nil, err
	}
	name := m.SourceName()
	pgm, err := goja.Compile(name, out.Code, true)
	if err != nil {
		return "", nil, compileError(name, err)
	}
	return out.Code, pgm, nil
}

// SourceName is the name errors and stack frames report for the module: the
// pack-relative key with the resolved file extension (for example
// @local/lib/helpers.ts), or the key itself for a builtin.
func (m *ModuleSpec) SourceName() string {
	key := m.KeyOrPath()
	if m.Builtin || m.Path == "" || m.Key == "" {
		return key
	}
	return path.Join(path.Dir(key), filepath.Base(m.Path))
}

func (m *ModuleSpec) KeyOrPath() string {
	if m.Key != "" {
		return m.Key
	}
	return m.Path
}

// LoadRequire resolves & compiles a dependency of another module by spec.
func (r *Registry) LoadRequire(fromDir, spec string) (*ModuleSpec, error) {
	memoKey := fromDir + "\x00" + spec
	if m, ok := r.resolved.Load(memoKey); ok {
		mod := m.(*ModuleSpec)
		_, err := r.programFor(mod)
		return mod, err
	}
	mod, err := r.loadRequire(fromDir, spec)
	if err == nil {
		r.resolved.Store(memoKey, mod)
	}
	return mod, err
}

func (r *Registry) loadRequire(fromDir, spec string) (*ModuleSpec, error) {
	key, path, dir, builtin, err := r.resolveRequire(fromDir, spec)
	if err != nil {
		return nil, err
	}
	if err := r.checkGrant(key, builtin); err != nil {
		return nil, err
	}
	mod := r.getOrCreateModule(key, path, dir, builtin)
	if mod == nil {
		return nil, fmt.Errorf("module %s not found", key)
	}
	_, err = r.programFor(mod)
	return mod, err
}
