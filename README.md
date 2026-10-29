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
| Time a `serve` request may ask for | Up to `--eval-timeout`, so a request can only shorten its deadline | `--eval-max-timeout 30s` on `serve` raises the ceiling (see below) |
| Memory | **Not bounded per evaluation.** goja has no per-VM heap cap | Set a process limit, for example `GOMEMLIMIT=1GiB`, plus a container or cgroup memory limit |

The deadline stops module JavaScript wherever it runs: module loading, export getters and Proxy traps read during binding and lookup, the call itself, and the conversion of its return value. It also stops Sentrie's own expression evaluation: the evaluator checks it at every rule and derive it enters and at every lambda it applies, which includes each element a collection operator such as `filter` or `any` visits, so a policy with no module calls is bounded too. A single step between those points (one comparison, one `in` over a list) still runs to its end.

A `serve` client that knows its own time budget, such as an Envoy `ext_authz` filter or an admission webhook, can choose the deadline for one request with the `Sentrie-Timeout` header, as a Go duration (`250ms`, `2s`, `1m30s`):

```bash
curl -i -X POST http://localhost:7529/decision/sh/sentrie/example/user_access \
  -H "Content-Type: application/json" \
  -H "Sentrie-Timeout: 500ms" \
  -d '{"facts":{"user":{"role":"admin","status":"active"}}}'
```

- `--eval-timeout` is the default for requests that don't send the header, and `--eval-max-timeout` is the ceiling a request may ask for, below or above the default:

  | `--eval-timeout` | `--eval-max-timeout` | A request may ask for | Without the header |
  | :--- | :--- | :--- | :--- |
  | D | unset | up to D | D |
  | D | M >= D | up to M | D |
  | D | M < D | `serve` refuses to start | |
  | 0 (no deadline) | unset | anything | no deadline |
  | 0 | M | up to M | M |

- A value above the ceiling, or one that is unparseable, zero, negative, longer than 32 bytes or sent more than once, gets a 400 problem-details response and nothing is evaluated. There is no clamping and no fallback to the default.
- Every decision response that ran under a deadline carries `Sentrie-Timeout-Applied` with the deadline used, in the same format, including the 503 when it passes.
- The write timeout is sized from the ceiling (ceiling + 5s, at least 30s), so a request that asks for the longest deadline can still get its response.
- A caller-chosen deadline lets any client that can reach the port hold an evaluation for up to the ceiling. Keep the ceiling to what your callers need, and put `serve` behind authentication ([#115](https://github.com/sentrie-sh/sentrie/issues/115)).

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
