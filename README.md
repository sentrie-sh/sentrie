# Sentrie

[![CLA Required](https://img.shields.io/badge/CLA-required-blue.svg)](CLA.md)
[![License](https://img.shields.io/badge/License-Apache_2.0-green.svg)](LICENSE)
[![Dual License](https://img.shields.io/badge/license-open--source%20%7C%20commercial-orange)](LICENSE-DUAL.md)

Sentrie is an open-source policy enforcement engine that lets you write business rules in a dedicated language. Instead of embedding policy logic in your application code, you define rules declaratively and let Sentrie evaluate them.

## Installation

Sentrie is distributed as a single binary with no external dependencies.

### Quick Install

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

## Learn More

- **[Getting Started](https://sentrie.sh/getting-started/)** - Write your first policy
- **[Language Reference](https://sentrie.sh/reference/)** - Complete language documentation
- **[CLI Reference](https://sentrie.sh/cli-reference/)** - Command-line interface guide
- **[TypeScript Modules](https://sentrie.sh/typescript-modules/)** - Extend policies with JavaScript
- **[Running Sentrie](https://sentrie.sh/running-sentrie/)** - Production deployment guide

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) and [LICENSE-DUAL.md](LICENSE-DUAL.md) for details.
