# ADR-0002: Non-Turing Complete Language Design

**Status:** Accepted  
**Date:** 2025-09-14  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, safety, security

## Context

Policy engines are used in security-critical contexts where:

- Decisions must be made in bounded time
- Infinite loops or stack overflows cannot occur
- All computations must be traceable and auditable
- Policies should be easy to reason about and verify

Traditional programming languages are Turing-complete, meaning they can express any computation, including computations that never terminate. This creates risks in policy evaluation contexts.

## Decision

Sentrie's language is **intentionally non-Turing complete**. The language is designed to:

- Always terminate in bounded time
- Have predictable performance characteristics
- Be easily auditable and traceable
- Focus on data evaluation rather than arbitrary computation

The language does not support:

- General recursion
- Loops (except bounded collection operations)
- Unbounded iteration
- Function calls that could create unbounded call stacks

## Rationale

1. **Safety guarantees**: Non-Turing completeness ensures that policy evaluation will always terminate, preventing denial-of-service attacks through infinite loops or stack overflows.

2. **Predictable performance**: Bounded execution time makes it possible to set timeouts and resource limits with confidence. This is critical for production policy evaluation.

3. **Auditability**: All policy decisions can be traced through a finite execution path. This is essential for compliance and security auditing.

4. **Security**: Preventing arbitrary computation reduces the attack surface. Policies cannot execute arbitrary code that could exploit the runtime.

5. **Focus on purpose**: Policy evaluation is about data transformation and decision-making, not general computation. The language is optimized for this specific use case.

6. **Verification**: Non-Turing complete languages are easier to formally verify and reason about, which is valuable for security-critical policies.

## Consequences

### Positive

- Guaranteed termination of all policy evaluations
- Predictable performance characteristics
- Better security posture (no arbitrary code execution)
- Easier to audit and trace policy decisions
- Can set reliable timeouts and resource limits
- Easier to reason about policy correctness

### Negative

- Some computations that are possible in general-purpose languages are not expressible
- Policy authors may need to use TypeScript modules for complex business logic
- Learning curve for developers used to Turing-complete languages
- Some algorithms cannot be expressed directly in the language

### Neutral

- The language is still expressive enough for most policy evaluation needs
- TypeScript modules provide an escape hatch for complex computations
- Collection operations (any, all, reduce) provide bounded iteration

## Alternatives Considered

### Alternative A: Turing-Complete Language with Timeouts

**Description:** Make the language Turing-complete but add execution timeouts and resource limits.

**Pros:**

- Maximum expressiveness
- Can implement any algorithm
- Familiar to developers

**Cons:**

- Timeouts are reactive, not proactive
- Difficult to set appropriate timeout values
- Still vulnerable to resource exhaustion attacks
- Harder to audit (unbounded execution paths)
- No compile-time guarantees

**Why not chosen:** Timeouts don't provide the same level of safety guarantees as non-Turing completeness. They're a mitigation, not a prevention.

### Alternative B: Restricted Recursion

**Description:** Allow recursion but with depth limits and other restrictions.

**Pros:**

- More expressive than no recursion
- Can still bound execution

**Cons:**

- Complex to implement correctly
- Hard to determine appropriate depth limits
- Still allows some forms of unbounded computation
- More complex semantics

**Why not chosen:** The complexity doesn't provide enough benefit. The language's collection operations and TypeScript modules provide sufficient expressiveness without the risks.

### Alternative C: Functional Language with Bounded Evaluation

**Description:** Use a functional language model but restrict to primitive recursion or other bounded forms.

**Pros:**

- More expressive than current approach
- Still provides termination guarantees
- Familiar functional programming patterns

**Cons:**

- More complex language design
- Harder for non-functional programmers to use
- Still requires careful design to maintain bounds

**Why not chosen:** The current design provides sufficient expressiveness while being simpler and more accessible to policy authors.

## Implementation Notes

- The language supports bounded collection operations: `any`, `all`, `reduce`, `transform`
- These operations iterate over collections but are bounded by collection size
- TypeScript modules can be used for complex computations that require Turing-complete features
- The runtime enforces execution limits as a defense-in-depth measure
- All language constructs are designed to guarantee termination

## References

- Language documentation: [What is Sentrie?](https://sentrie.sh/getting-started/what-is-sentrie/)
- TypeScript modules: [Using TypeScript](https://sentrie.sh/reference/using-typescript/)
- Related: ADR-0001 (The Sentrie Language), ADR-0010 (TypeScript Module Support)
