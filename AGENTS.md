The name of the project is "Sentrie"

---

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

---

When you make any changes, update the **PR_DESCRIPTION_<branch_name>.md** file describing all the changes in the branch comprehensively.

**PR_DESCRIPTION_<branch_name>.md MUST describe only the changes introduced in this barnch** (compared to the base branch, e.g. origin/main). Do not include ot summarize wirk that already exists on the base or that was merged from other branches, To determine scope, get the list of commits and files changes using `git log base..HEAD` and `git diff base..HEAD` (base is typically origin/main), then document only those changes.

- Do not include changes for code/logic that were added in the branch and then removed and never re-added
- Summarize in a PR description format

The PR description should cover:
- **PR title** (see below)
- Summary
- What this PR does
- Changes by area
- Review notes
- Testing notes
- Dependency changes

Use short, bulleted points so it is easy to scan; one main idea per bullet. Lean toward sub-sullets over paragraphs.

For `## Changes By Area`, format each area as a Markdown `###` subheading (for example, `### Runtime evaluator and execution`) followed by bullets, instead of bold bullet area titles.

**PR title:** Put the actual pull request title **once** as the first line of the file: a single `#` heading whose text is the full title string (for Bitbucket/Github). **Do not** add a separate `# Title` section or any second top-level title - the first `#` line **is** the PR title, not a placeholder heading above the different title. The title **must end with** the issue number from the branch (the number prefix, e.g. branch `64-replace-pervas...` -> `64`). The title **must be suffixed** by `Closes #<issue_number>`

**Review notes:** Always make sure that `Review Notes` in the PR description is up to date and covers all the critical areas that reviewers should concentrate on.

--

**License Headers:** Make sure that all sources files have the correct license headers.

--

**Code commits:** Do not commit unless explicitly told to do so. **NEVER** commit any `*.md` files. Whenever you commit, Do NOT add cursor signatures in the commit messages. Make commits in reasonable chunks. Group changes into a commit if and only if the changes are tightly relates.

**Commit messages:** Use short imperative subject lines. Do not use conventional commit prefixes (e.g. no `refactor(scope):`, `feat(scope):` etc.). Prefer plain imperative messages: "Align X with Y", "Wire policy loading with parsing". Do not add further descriptions to the message if it doesn't add value.

--
