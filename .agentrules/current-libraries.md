# Prefer the current standard library and up-to-date SDKs

Before adding a dependency or writing against an older API, use what the pinned toolchain ships (Go 1.25, per `go.mod`). Take dependencies and SDKs at their current version.

**Why:** Every third-party package or superseded API is code we will migrate later. The newer standard library usually removes a dependency and brings stricter, better-maintained defaults. Precedent: #72 (#75) replaced `github.com/pkg/errors` with the standard `errors` package and typed wrappers. Writing new code against the old API only makes the eventual migration bigger.

**How to apply:**

- **Check the toolchain first.** `go doc <pkg>` on the pinned version shows what the standard library covers. For example, `slices`, `maps` and `cmp` cover most of what `golang.org/x/exp` once provided, and `errors.Join` covers multi-error wrapping.
- **New code uses the current API**, even where the code around it still uses the older one.
- **Don't convert existing code in passing.** A newer API can change behaviour that other code depends on. Migrating existing code is its own issue and PR, with the behaviour checked.
- **Keep one generation per file.** Where new code meets old, convert at the boundary rather than importing both the old and the new package into one file.
- **Dependencies:** when adding one, take its current major version and follow `THIRD_PARTY_POLICY.md`: permissive licences only, and the PR states licence, compatibility, motivation and vendoring status. `deps-policy-check.yml` enforces the licence list. When upgrading, move to the current version. If something holds a dependency back, say why in the PR and file an issue to finish the move.

Related: [Build and test](build-and-test.md), [PR conventions](pr-conventions.md)

Relevant keywords: standard library, stdlib, Go 1.25, toolchain, go.mod, dependency, third-party package, THIRD_PARTY_POLICY, licence, SDK, up to date, current version, upgrade, deprecated API, migration, pkg/errors, golang.org/x/exp.
