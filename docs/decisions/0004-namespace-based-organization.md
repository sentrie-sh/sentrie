# ADR-0004: Namespace-Based Organization

**Status:** Accepted  
**Date:** 2025-09-14  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, organization, scoping

## Context

As policies grow in complexity and number, organization becomes critical. Policies need:

- A way to group related policies together
- A mechanism to avoid naming conflicts
- A clear way to reference policies from other contexts
- Support for modular policy development

Different organizational models exist: flat naming, hierarchical namespaces, modules/packages, or file-based organization.

## Decision

Sentrie uses **namespace-based organization**. Each policy file declares a namespace, and all policies, rules, and shapes within that file belong to that namespace. Policies are referenced using a fully qualified name: `namespace/policy/rule`.

## Rationale

1. **Clear organization**: Namespaces provide a natural way to group related policies (e.g., `user_management`, `payment_processing`, `access_control`).

2. **Avoid naming conflicts**: Different namespaces can have policies with the same name without conflict.

3. **Modularity**: Teams can work on different namespaces independently without coordination.

4. **Explicit references**: Fully qualified names make it clear where a policy comes from, improving readability and reducing errors.

5. **Simple mental model**: One namespace per file is easy to understand and reason about.

6. **Compatibility with REST APIs**: The namespace/policy/rule structure maps naturally to URL paths in HTTP APIs.

7. **Scalability**: As the number of policies grows, namespaces provide a scalable organization mechanism.

## Consequences

### Positive

- Clear organization of policies
- No naming conflicts between different policy domains
- Easy to understand policy structure
- Natural mapping to API endpoints
- Supports team collaboration on different namespaces
- Explicit policy references

### Negative

- Fully qualified names can be verbose
- Must remember namespace when referencing policies
- One namespace per file (cannot mix namespaces in a single file)

### Neutral

- Namespaces are flat (no nested namespaces)
- File-based namespace declaration

## Alternatives Considered

### Alternative A: Flat Naming with Prefixes

**Description:** Use flat names with prefixes like `user_management_allow_access`, `payment_process_transaction`.

**Pros:**

- Simple, no namespace concept
- Easy to reference (just the name)
- No file organization needed

**Cons:**

- Verbose names
- Easy to have naming conflicts
- Hard to organize large numbers of policies
- Prefixes can become inconsistent

**Why not chosen:** Namespaces provide better organization and avoid the verbosity and conflict issues of flat naming.

### Alternative B: Nested/Hierarchical Namespaces

**Description:** Support nested namespaces like `company.team.feature.policy`.

**Pros:**

- Very fine-grained organization
- Can model organizational structure
- Supports deep hierarchies

**Cons:**

- More complex to implement
- Harder to understand and use
- Can lead to overly deep nesting
- More verbose references

**Why not chosen:** Flat namespaces provide sufficient organization without the complexity. Most policy organizations don't need deep nesting.

### Alternative C: File-Based Organization

**Description:** Organize by file/directory structure without explicit namespace declarations.

**Pros:**

- Simple - just use file paths
- Familiar to developers
- No namespace syntax needed

**Cons:**

- Couples organization to file system
- Harder to reorganize files
- Less explicit in code
- Platform-dependent (path separators)

**Why not chosen:** Explicit namespaces in code are clearer and more portable than file-based organization.

### Alternative D: Module/Package System

**Description:** Use a module or package system similar to Go, Rust, or Python.

**Pros:**

- Familiar to many developers
- Can support imports/exports
- Well-established patterns

**Cons:**

- More complex than needed
- Requires import/export syntax
- Can lead to dependency management complexity
- Overkill for policy organization

**Why not chosen:** Namespaces provide sufficient organization without the complexity of a full module system. Policies are typically self-contained.

## Implementation Notes

- Namespaces are declared at the top of each `.sentrie` file
- All policies, rules, and shapes in a file belong to that namespace
- References use the format: `namespace/policy/rule`
- The index system (`index/`) manages namespace resolution
- CLI and API use the same namespace/policy/rule format

## References

- Language documentation: [Namespaces](https://sentrie.sh/reference/namespaces/)
- Implementation: `parser/namespace.go`, `index/namespace.go`
- Related: ADR-0001 (The Sentrie Language)
