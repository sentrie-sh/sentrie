# Sandbox Programs

Use the `sandbox.go` file to write sandboxed programs you need to run to test your code. It carries the `sandbox` build tag, so it never builds into the binary. Run it on its own with `go run -tags sandbox sandbox.go`. Don't use `go run -tags sandbox .`, which also compiles the root `main.go` and fails with two `main` functions.

**After testing is complete, RESET `sandbox.go` to its template state:**

- Remove all test code from the `main()` function.
- Remove any imports that were added for testing.
- Restore the `main()` function to only contain the template code:
  ```go
  func main() {
      ctx := context.Background()
      // Your test code here
      fmt.Println("Sandbox test")
      os.Exit(0)
  }
  ```
- Keep only the standard imports (`context`, `fmt`, `os`) unless they're needed for the template.

**Why:** One scratch file keeps throwaway programs out of the packages and out of review. Resetting it keeps one person's experiment out of the next one's diff.

**How to apply:** Never commit scratch code in `sandbox.go`. Before committing, `git diff -- sandbox.go` must be empty.

Known gotcha: the template as written fails `go vet` and won't compile without a use of `ctx` (`declared and not used: ctx`). Use `ctx` in your test code, or add `_ = ctx` while experimenting.

Related: [Go project layout](go-project-layout.md), [Commits](commits.md)
