# Go Test Conventions

Go tests use test-scoped contexts and extend each package's existing testify suite.

## Contexts

- In tests, prefer test-scoped contexts over `context.Background()`.
- Plain tests use `t.Context()`.
- Testify suite methods use `s.T().Context()` or `suite.T().Context()`, matching the local receiver or helper style.
- Use `context.Background()` in a test only when a real root context is required, and document that reason in a nearby comment.

## Suites

- When a package already has a documented shared testify suite, extend that suite instead of adding new standalone `Test...` functions, unless the file is the suite entrypoint. Most packages declare it in `suite_test.go` (for example `TestParserTestSuite`, `TestIndexTestSuite`, `TestJSTestSuite`).
- In `runtime/`, feature-specific tests are methods on `RuntimeTestSuite`, even when split across separate `*_test.go` files. The standalone `TestRuntimeTestSuite` entrypoint in `runtime/runtime_test.go` is the expected exception.
- Don't force the `RuntimeTestSuite` pattern onto packages that deliberately use a different test structure. Follow each package's local convention unless guidance says otherwise.

**Why:** Test-scoped contexts are cancelled when the test ends, so leaked goroutines and timers surface as failures rather than hiding. A single suite per package keeps shared setup in one place (standardised in #66).

**How to apply:** Applies to every `**/*_test.go` file. Before adding a test, find the package's `Test...Suite` entrypoint with `rg -n 'func Test\w+Suite' <pkg>/` and add a method to it.

Related: [Go project layout](go-project-layout.md), [Build and test](build-and-test.md)
