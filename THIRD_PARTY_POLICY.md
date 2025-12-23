# Third-Party Dependency Contribution Policy

To maintain the legal clarity and integrity of the Sentrie project, any contribution that introduces or modifies third-party dependencies must follow the requirements below.

---

## 1. Allowed License Types

Contributions **may use** third-party libraries only if they are licensed under one of:

- Apache 2.0
- MIT
- BSD (2-clause or 3-clause)
- ISC
- Public Domain (CC0)

These licenses are permissive and compatible with both open-source and commercial redistribution. See [LICENSE-DUAL.md](LICENSE-DUAL.md) for details.

---

## 2. Prohibited or Restricted Licenses

The following licenses may **not** be added as direct or transitive dependencies:

- GPL (all versions)
- AGPL (all versions)
- LGPL (all versions)
- EPL
- MPL (without explicit compatibility review)
- Licenses requiring source redistribution on link (copyleft)
- Licenses requiring redistribution of “interactive network use” source

These licenses introduce obligations incompatible with Sentrie's open licensing model. Sentrie is released under the Apache License 2.0 and as such, hinder enterprise adoption and distribution of the project. See [LICENSE.md](LICENSE.md) for details.

> Sentrie has a Github Actions workflow that will check for prohibited licenses and will fail the build if any are found. This workflow can be audited at [.github/workflows/deps-policy-check.yml](.github/workflows/deps-policy-check.yml).

---

## 3. Requirements When Adding Dependencies

When adding or updating a dependency, contributors must include:

1. **License identification**  
   Include license name, link, SPDX identifier.

2. **Compatibility statement**  
   Confirm it falls under the allowed category.

3. **Motivation**  
   Brief explanation of why this dependency is needed.

4. **Vendoring status**  
   State whether the dependency is vendored, module-based, or optional.

All new dependencies will be reviewed during PR approval.

---

## 4. Transitive Dependencies

Contributors are responsible for confirming that **transitive dependencies** also satisfy this policy.

Tools like `go mod graph` or [go-licenses](https://github.com/google/go-licenses) may be used to verify license ancestry.

---

## 5. Security & Compliance

All dependencies must pass:

- Basic security checks (`govulncheck`)
- Active maintenance status (no abandoned packages), unless declared stable by the maintainers
- Stable versioning (no random Git SHAs unless justified)

The maintainers may request replacement of dependencies that do not meet these standards.

---

## 6. Violations

If a dependency is added that violates this policy:

- The contributor may be asked to rework their PR
- Future contributions may be restricted if violations are repeated

---

This policy ensures Sentrie remains sustainable, redistributable, and legally safe for enterprises to adopt and use for the long term.
