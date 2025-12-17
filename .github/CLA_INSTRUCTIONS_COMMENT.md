[Sentrie CLA]

Hi @{{PR_AUTHOR}}, and thanks for your contribution.

Before we can merge this pull request, you need to agree to the Sentrie Contributor License Agreement (CLA).

1. Read the [CLA document](CLA.md) in this repository.
2. In this PR, edit the [`cla-signers.yaml`](cla-signers.yaml) file and add yourself:

For individuals:
```yaml
- handle: {{PR_AUTHOR}}
  cla_version: {{CURRENT_VERSION}}
```

If you are contributing on behalf of an organization:

- If your organization is already listed under `organizations`, add yourself as an additional representative. For example, if it currently looks like:

```yaml
organizations:
  - name: "Existing Org Name"
    cla_version: {{CURRENT_VERSION}}
    representatives:
      - handle: existing-rep
```

you can update it to:

```yaml
organizations:
  - name: "Existing Org Name"
    cla_version: {{CURRENT_VERSION}}
    representatives:
      - handle: existing-rep
      - handle: {{PR_AUTHOR}}
```

- If your organization is not yet listed, add a new entry under `organizations`, for example:

```yaml
organizations:
  - name: "Your Organization Name"
    cla_version: {{CURRENT_VERSION}}
    representatives:
      - handle: {{PR_AUTHOR}}
```

Once this PR includes your entry and it is merged, future PRs from this account will pass the CLA check automatically for this CLA version.
