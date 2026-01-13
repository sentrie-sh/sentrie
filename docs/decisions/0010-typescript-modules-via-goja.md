# ADR-0010: TypeScript Module Support via goja

**Status:** Accepted  
**Date:** 2025-11-02  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** runtime, integration, extensibility

## Context

While Sentrie's language is intentionally non-Turing complete for safety, some policy logic requires complex computations that are difficult or impossible to express in the core language. Examples include:

- Cryptographic operations (hashing, signing, verification)
- Complex string manipulation and parsing
- Date/time calculations
- Network operations
- Advanced data transformations

The language needs a way to extend functionality while maintaining safety and performance.

## Decision

Sentrie supports **TypeScript modules** that are executed in a JavaScript runtime (goja). Policies can import and use functions from TypeScript modules. The modules are:

- Compiled from TypeScript to JavaScript using esbuild
- Executed in an isolated goja runtime
- Subject to permission controls (file system, network access, etc.)
- Sandboxed from the main Go runtime

## Rationale

1. **Extensibility**: TypeScript modules provide an escape hatch for complex computations that can't be expressed in the core language.

2. **Familiar language**: TypeScript/JavaScript is widely known, making it accessible to many developers.

3. **Rich ecosystem**: JavaScript has extensive libraries for common operations (crypto, date handling, string manipulation, etc.).

4. **Isolation**: goja provides a separate JavaScript runtime that's isolated from the Go runtime, providing security boundaries.

5. **Performance**: goja is a fast JavaScript implementation written in Go, providing good performance for module execution.

6. **Type safety**: TypeScript provides type checking for modules, catching errors before runtime.

7. **Permission system**: Modules can be restricted based on packfile permissions, controlling what operations they can perform.

8. **Build-time compilation**: esbuild compiles TypeScript to JavaScript at build time, providing fast execution.

## Consequences

### Positive

- Extends language capabilities without compromising core language safety
- Familiar to many developers
- Rich ecosystem of JavaScript libraries
- Isolated execution environment
- Permission-based security controls
- TypeScript provides type safety
- Good performance with goja

### Negative

- Introduces JavaScript runtime dependency
- More complex than pure Go implementation
- Requires understanding of both Sentrie language and TypeScript
- Permission system adds complexity
- Potential security concerns (JavaScript execution)
- Build process must compile TypeScript

### Neutral

- Modules are optional (policies can work without them)
- Permission system controls module capabilities
- Modules are sandboxed but not completely isolated

## Alternatives Considered

### Alternative A: Pure Go Extensions

**Description:** Allow writing extensions in Go that are compiled into Sentrie.

**Pros:**

- Native performance
- Full access to Go ecosystem
- Type safety with Go types
- No JavaScript runtime needed

**Cons:**

- Requires recompiling Sentrie for new extensions
- Not accessible to policy authors (requires Go knowledge)
- Security concerns (full Go access)
- Distribution complexity

**Why not chosen:** TypeScript modules are more accessible to policy authors and don't require recompiling Sentrie. The isolation provided by goja is sufficient for the use case.

### Alternative B: WebAssembly (WASM)

**Description:** Compile extensions to WebAssembly and execute in a WASM runtime.

**Pros:**

- Language agnostic (can use many languages)
- Strong isolation
- Good performance
- Portable

**Cons:**

- More complex build process
- Limited access to host capabilities
- Requires WASM runtime
- Less familiar to most developers
- Tooling is less mature

**Why not chosen:** goja provides sufficient isolation with better developer experience and tooling. WASM adds complexity without clear benefits for this use case.

### Alternative C: Lua Extensions

**Description:** Use Lua for extensions, similar to many other systems.

**Pros:**

- Lightweight runtime
- Simple language
- Good embedding support
- Fast execution

**Cons:**

- Less familiar than JavaScript
- Smaller ecosystem
- Less type safety
- Less tooling support

**Why not chosen:** TypeScript/JavaScript is more familiar to developers and has a richer ecosystem. The performance difference is not significant enough to justify the trade-off.

### Alternative D: No Extensions

**Description:** Keep the language pure and require all logic to be expressible in the core language.

**Pros:**

- Simpler system
- No security concerns from extensions
- Consistent language model
- Easier to reason about

**Cons:**

- Severely limits what policies can do
- Many real-world scenarios require complex computations
- Forces workarounds or limitations
- Less practical for production use

**Why not chosen:** Real-world policies need capabilities beyond what a non-Turing complete language can provide. Extensions are necessary for practical use.

## Implementation Notes

- TypeScript modules are compiled using esbuild at pack load time
- Modules are executed in goja runtime instances
- Modules export functions that can be called from Sentrie policies
- Module execution is sandboxed but shares some Go runtime capabilities
- Built-in modules provide common functionality (crypto, time, JSON, etc.)

## References

- Language documentation: [Using TypeScript](https://sentrie.sh/reference/using-typescript/)
- Implementation: `runtime/js/`, `runtime/js/builtin_go.go`
- Related: ADR-0001 (The Sentrie Language), ADR-0002 (Non-Turing Complete Design)
