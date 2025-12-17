# Contributing to Sentrie

Thanks for thinking about contributing to Sentrie 💙  
This project is meant to be used in serious places (access control, guardrails, privacy), so we try to keep the contribution workflow simple, explicit, and boringly reliable.

This document explains how to contribute code, how the CLA works, and what we expect in pull requests.

---

## Ways to contribute

There are many ways to help:

- Fix bugs
- Improve documentation or examples
- Add small, focused features
- Improve error messages or UX
- Tidy internals (refactors, tests, performance)

For large or architectural changes, **open an issue first** so we can align on the direction before you invest a lot of time.

---

## Contributor License Agreement (CLA)

Before we can merge your pull request, you must agree to the **Sentrie Contributor License Agreement (CLA)**.

The CLA:

- Confirms that you have the right to contribute the code
- Grants the project the necessary rights to use, modify, and redistribute your contribution
- Lets you **keep ownership** of your code
- Applies to all your current and future contributions under the same GitHub handle, for the same `cla_version`

The CLA text lives in:

- [`CLA.md`](CLA.md) in the repository

> **The CLA check is enforced by a GitHub Action.**  
> If you have not signed for the current `cla_version`, the PR check will fail with a message telling you what to do.

---

## How to sign the CLA (individuals)

If you are contributing as an individual, follow these steps:

1. **Read the CLA** in `CLA.md`.
2. In your pull request, edit the file `cla-signers.yaml`.
3. Under the `individuals:` section, add an entry with your GitHub handle and the current `cla_version`:

```yaml
individuals:
  - handle: your-github-username
    type: individual
    cla_version: 1
```

4. Replace your-github-username with your actual GitHub handle.

> Once this PR is merged, future PRs from the same GitHub handle will pass the CLA check automatically for this CLA version. If we ever bump `cla_version` in `cla-signers.yaml`, you may be asked to update your entry again.

# How to sign the CLA (organizations)

If you are contributing on behalf of a company or organization:

1. Have an authorized representative of your organization:

   - Read CLA.md
   - Open a PR that updates cla-signers.yaml and adds an entry like:

   ```yaml
   organizations:
     - name: "Acme Corp"
       cla_version: 1
       representatives:
         - handle: acme-dev-1
           type: representative
         - handle: acme-dev-2
           type: representative
   ```

2. After this is merged:
   - Any handle listed under representatives is treated as contributing on behalf of that organization for the specified cla_version.
   - Additional representatives can be added later via small PRs updating `cla-signers.yaml`.

> Right now, we expect that most contributors will probably use the individual path. The organization flow is there so we don’t have to redesign this later.

# What the CLA GitHub Action does

On each pull request, a GitHub Action will:

1. Read `cla-signers.yaml` from the base branch (usually main).
2. Check if your GitHub handle:
   - Appears under individuals with `cla_version >= cla_version` at the top, or
   - Appears under `organizations[*].representatives` where the organization’s `cla_version >= cla_version` at the top.
3. If not found or outdated:

- Check whether this PR modifies `cla-signers.yaml` to add/update your entry.

4. If the CLA is still not satisfied:

- The Action will post a comment explaining how to update `cla-signers.yaml`.
- The check will fail and the PR cannot be merged until you add the entry.

> All of this state lives in Git; there is no external service or hidden database.

---

# General PR guidelines

To keep the project maintainable and reviewable:

- Keep PRs small and focused
  - One logical change per PR is ideal.
- Tests:
  - Add or update tests for any behavior changes.
- Style:
  - Follow the existing code style and patterns.
- Breaking changes:
  - Call them out explicitly in the PR description.
  - For anything user-facing or policy-breaking, always discuss in an issue first.

---

# Development workflow (high level)

This will evolve over time, but the rough flow is:

1. Fork the repo and create a topic branch.

2. Make your changes.

3. Add or update tests where appropriate.

4. If this is your first contribution (or the CLA version has changed), update cla-signers.yaml as described above.

5. Open a pull request:

- Explain what changed and why.

- Mention any backwards-incompatible behavior.

> We’ll do our best to review in a timely manner.

---

# Ending notes

We’re looking forward to your contributions!

If you have any questions, please feel free to reach out to us.

Thanks again for contributing to Sentrie.

You’re helping build a deterministic, explainable policy engine that people can actually trust in production.
