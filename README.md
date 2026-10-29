# Sentrie

[![CLA Required](https://img.shields.io/badge/CLA-required-blue.svg)](CLA.md)
[![License](https://img.shields.io/badge/License-Apache_2.0-green.svg)](LICENSE)
[![Dual License](https://img.shields.io/badge/license-open--source%20%7C%20commercial-orange)](LICENSE-DUAL.md)

Sentrie is an open-source policy enforcement engine that lets you write business rules in a dedicated language. Instead of embedding policy logic in your application code, you define rules declaratively and let Sentrie evaluate them.

## Installation

Sentrie is distributed as a single binary with no external dependencies.

### Quick Install

**macOS (recommended):**

```bash
brew install sentrie-sh/tap/sentrie
```

> **Note:** We use Homebrew Formulas instead of Casks because macOS Gatekeeper flags binaries installed via Casks, requiring workarounds. Formulas are treated as regular binaries by Gatekeeper, providing a better user experience.

**macOS, Linux, and WSL2:**

```bash
curl -fsSL https://sentrie.sh/install.sh | bash
```

**Windows:**

```powershell
irm https://sentrie.sh/install.ps1 | iex
```

For detailed installation instructions and platform-specific options, see the [installation guide](https://sentrie.sh/getting-started/installation/).

## Basic Usage

### Write a Policy

Create a policy file `policy.sentrie`:

```sentrie
namespace user_management

shape User {
  role: string
  status: string
}

policy user_access {
  fact user:User

  rule allow = {
    yield user.role == "admin" or (user.role == "user" and user.status == "active")
  }

  export decision of allow
}
```

### Execute a Policy

```bash
sentrie exec user_management/user_access/allow --facts '{"user":{"role":"admin","status":"active"}}'
```

### Run as HTTP Service

```bash
sentrie serve
```

Then make a request:

```bash
curl -X POST http://localhost:7529/decision/user_management/user_access/allow \
  -H "Content-Type: application/json" \
  -d '{"facts":{"user":{"role":"admin","status":"active"}}}'
```

### Evaluation Limits

TypeScript and JavaScript modules run inside Sentrie, so evaluation is bounded by default:

| Resource | Bound | Configure |
| :--- | :--- | :--- |
| Time per decision | 10s for each `exec` evaluation and each `serve` request; JavaScript still running is interrupted and the decision fails (`serve` answers 503) | `--eval-timeout 2s` on `exec` and `serve`; `0` disables it (`serve` still closes a response after its 30s write timeout) |
| JavaScript call depth | 10,000 frames per module call; deeper recursion throws a stack overflow | Not configurable |
| Memory | **Not bounded per evaluation.** goja has no per-VM heap cap | Set a process limit, for example `GOMEMLIMIT=1GiB`, plus a container or cgroup memory limit |

The deadline stops module JavaScript wherever it runs: module loading, export getters and Proxy traps read during binding and lookup, the call itself, and the conversion of its return value. Sentrie's own expression evaluation is not interrupted mid-expression; it ends when it next calls into a module, or when it finishes.

## Learn More

- **[Getting Started](https://sentrie.sh/getting-started/)** - Write your first policy
- **[Language Reference](https://sentrie.sh/reference/)** - Complete language documentation
- **[CLI Reference](https://sentrie.sh/cli-reference/)** - Command-line interface guide
- **[TypeScript Modules](https://sentrie.sh/typescript-modules/)** - Extend policies with JavaScript
- **[Running Sentrie](https://sentrie.sh/running-sentrie/)** - Production deployment guide

### Working on Sentrie itself

- **[Codebase Knowledge Graph](docs/codebase-graph/overview.md)** - Architecture, dependencies, and contracts for every package and boundary. The fastest way to orient in the source.
- **[Contributing](CONTRIBUTING.md)** - Development workflow and CLA

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) and [LICENSE-DUAL.md](LICENSE-DUAL.md) for details.
