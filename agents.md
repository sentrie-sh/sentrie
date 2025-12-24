The name of the project is "Sentrie"

Use the `sandbox.go` file to write sandboxed programs you need to run to test your code.

**IMPORTANT: After testing is complete, RESET `sandbox.go` to its template state:**

- Remove all test code from the `main()` function
- Remove any imports that were added for testing
- Restore the `main()` function to only contain the template code:
  ```go
  func main() {
      ctx := context.Background()
      // Your test code here
      fmt.Println("Sandbox test")
      os.Exit(0)
  }
  ```
- Keep only the standard imports (context, fmt, os) unless they're needed for the template

When you make any changes, update the `@PR_DESCRIPTION.md` file to describe all the changes in the current branch. The content of `@PR_DESCRIPTION.md` should be in markdown format. If you find existing content in the file that contradicts the changes you have made, overwrite the existing content. The file will be used to generate the pull request description.

The title in the `@PR_DESCRIPTION.md` file should be such that it completes the sentence: "If merged, this pull request will ...". Keep it below 60 words. You may use markdown formatting.

You MUST put the following after the title:

- A description that explained WHAT changed and WHY
- Notes on what to focus on in a review of the PR
- Clear answers to questions that reviewers may have

Follow the format:

```markdown
## Summary

<!-- What does this PR change and why? -->

## Testing

<!-- How did you test this change? -->

## Dependencies

<!--
If this PR adds or changes third-party dependencies, list them here.

Example:

- Added: github.com/google/go-licenses (Apache-2.0)
- Reason: Used for license scanning in CI.
- Type: direct dependency

If no dependencies were added/changed, write:

- None
-->
```

Make sure that all sources files have the correct license headers.
