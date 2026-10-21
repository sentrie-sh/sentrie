## Binaek Sarkar (binaek89@gmail.com)

### 14:40 - Plan the #168 stack
- #168 asks for operator capabilities (a deployment-level grant of builtins, `@sentrie/*` modules and local TypeScript modules) plus nondeterminism metadata, and optionally deterministic `Date.now`/`Math.random` shims. Well over 800 hand-written lines, so it ships as a stack:
  - `168-1-nondeterminism-metadata` (this branch): a `Nondeterministic` flag on `builtins.Decl`, a nondeterministic-export table for the Go and TypeScript `@sentrie/*` modules, derive rejection of nondeterministic builtins (static and runtime, paired per the evalDerive rule) and `sentrie validate` warnings.
  - `168-2-capabilities`: the capabilities JSON file, `--capabilities` on `exec`, `serve` and `validate`, the index-time check with spans and the paired runtime gates.
  - `168-3-deterministic-js`: optional seeded `Date.now`/`Math.random` shims behind a capabilities flag.
- Nondeterministic sources found: builtin `now()`; `@sentrie/time.now`; `@sentrie/uuid.v4`, `v6`, `v7`; `@sentrie/js.now` (`Date.now`) and `@sentrie/js.random` (`Math.random`). `@sentrie/jwt` does not read the clock (no `exp` checking), and no other Go module calls `time.Now` or a random source.
- Other agents touch `builtins/` (#153 adds `now_ms()`), `runtime/js/registry.go` `programFor` (#156) and `runtime/executor.go` binding code, so the JS metadata goes in a new `runtime/js/nondeterminism.go` instead of changing `RegisterGoBuiltin`.
- _Model: claude-opus-5-5_
