# 102-binding-export-key

## Binaek Sarkar (binaek89@gmail.com)

### 14:41 - Reproduce #102 and pick the design
- Bottom layer of a chain of five branches (#102, #105, #191, #154, #142), each its own `Closes` PR, each cut from the one below.
- Cause confirmed: `getModuleBinding` keys perch on `ms.KeyOrPath()`, but `jsBindingConstructor` closes over one `use` statement and projects exports by it.
  - With `ttl = -1` (#191) perch only deduplicates concurrent loads, so the leak needs two different `use`s of one module loading at the same time.
  - That is exactly the layout of one policy with two aliases of one module under `ExecPolicy`: every exported rule runs `bindUses` in its own goroutine, and the second alias joins the first alias's in-flight load.
  - `example_pack/minimal.sentrie` already has this shape (`@sentrie/js` as `math` and as `str` in `jsglobalpolicy`); that policy currently stops earlier on a `let` conflict, so the smoke test never reaches it.
- Regression tests in `runtime/module_exports_test.go`, through a new temp-pack helper `newTempPackExecutor` in `runtime/module_pack_test.go`:
  - `TestTwoUsesOfOneModuleInOnePolicy` fails on `main` with `function a not found in module "mb"`.
  - `TestUseExportRestrictionIsPerUseStatement` (two policies, both orders, own export reachable, other's not) and `TestUseMissingRequiredExportIsReportedPerUse` pass on `main` only because nothing is cached; they guard #191's branch, which turns caching on.
- Design (issue option 2): cache the loaded module by path, and make the export projection a per-`use` property of the binding.
  - `JSInstance` keeps the module's whole exports object and its own enumerable keys, instead of a pre-filtered map.
  - `ModuleBinding` becomes a cheap per-`use` value: canonical key, alias, requested export names, and a pointer to the shared module (the pool).
  - `Call` applies the projection: a named `use` reaches only its names; a bare `use` reaches only the module's own enumerable exports, as before.
  - The "missing required export" check runs per `use` at bind time, against an instance acquired from the shared pool.
- _Model: claude-opus-5-5_
