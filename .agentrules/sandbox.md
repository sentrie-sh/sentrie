# Sandbox Programs

Use the `sandbox.go` file to write sandboxed programs you need to run to test your code. It carries the `sandbox` build tag, so it never builds into the binary. Run it on its own with `GOWORK=off go run -tags sandbox sandbox.go` ([Build and test](build-and-test.md)). Don't use `go run -tags sandbox .`, which also compiles the root `main.go` and fails with two `main` functions.

**After testing is complete, RESET `sandbox.go` to its template state:**

- Remove all test code from the `main()` function.
- Remove any imports that were added for testing.
- Restore the `main()` function to only contain the template code:
  ```go
  func main() {
      ctx := context.Background()
      _ = ctx // remove once your test code uses ctx
      // Your test code here
      fmt.Println("Sandbox test")
      os.Exit(0)
  }
  ```
- Keep only the standard imports (`context`, `fmt`, `os`) unless they're needed for the template.

**Why:** One scratch file keeps throwaway programs out of the packages and out of review. Resetting it keeps one person's experiment out of the next one's diff.

**How to apply:** Never commit scratch code in `sandbox.go`. Before committing, `git diff -- sandbox.go` must be empty.

The `_ = ctx` line keeps the template compiling and passing `go vet` while `ctx` is unused. Delete it once your test code uses `ctx`.

Related: [Go project layout](go-project-layout.md), [Build and test](build-and-test.md), [Commits](commits.md)
