# ADR-0011: Pack System for Policy Organization

**Status:** Accepted  
**Date:** 2025-09-28  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** organization, packaging, distribution

## Context

Policies need to be organized and distributed. A single policy file may not be sufficient for complex scenarios. The system needs:

- A way to group related policies together
- Metadata about policy collections (version, author, description)
- A way to specify dependencies and requirements
- Support for multiple files (policies, shapes, TypeScript modules)
- A mechanism for distribution and sharing

Different approaches exist: single files, directory-based organization, package manifests, or archive formats.

## Decision

Sentrie uses a **pack system** where policies are organized in directories with a `sentrie.pack.toml` manifest file. The pack file defines:

- Pack metadata (name, version, description, license)
- Schema version
- Engine requirements
- Permissions for TypeScript modules
- Authors and repository information

Packs can contain multiple `.sentrie` files, TypeScript modules, and other resources.

## Rationale

1. **Organization**: Packs provide a natural way to group related policies, making it easy to organize complex policy sets.

2. **Metadata**: Pack files provide versioning, licensing, and other metadata needed for distribution and compliance.

3. **Permissions**: Pack files declare what permissions TypeScript modules need, enabling security controls.

4. **Discoverability**: Pack files make it easy to understand what a policy collection contains and requires.

5. **Distribution**: Packs can be versioned, shared, and distributed as units.

6. **Dependency management**: Pack files can specify engine version requirements, ensuring compatibility.

7. **Flexibility**: Packs can contain multiple files, supporting complex policy structures.

8. **Standard format**: TOML is human-readable and widely supported.

## Consequences

### Positive

- Clear organization of policy collections
- Versioning and metadata support
- Permission declarations for security
- Easy to distribute and share
- Supports complex multi-file policies
- Human-readable manifest format

### Negative

- Must create pack file for each policy collection
- Additional file to maintain
- Pack file discovery (walking up directory tree)
- Must understand pack file schema

### Neutral

- Pack files are optional for simple use cases (can use single files)
- Pack discovery walks up directory tree to find pack file
- Pack files use TOML format

## Alternatives Considered

### Alternative A: Single File Policy

**Description:** All policies in a single file, no packaging system.

**Pros:**

- Simple, no extra files
- Easy to understand
- No organization overhead

**Cons:**

- Doesn't scale to complex policies
- No metadata or versioning
- Hard to organize multiple related policies
- No permission declarations
- Difficult to distribute

**Why not chosen:** Real-world policies need organization, versioning, and metadata. Single files don't scale.

### Alternative B: Directory-Based (No Manifest)

**Description:** Organize by directory structure without a manifest file.

**Pros:**

- Simple, just use directories
- No manifest to maintain
- Familiar file system organization

**Cons:**

- No metadata (version, author, etc.)
- No permission declarations
- Hard to understand requirements
- No standard structure
- Difficult to distribute

**Why not chosen:** Manifest files provide essential metadata and structure that directory-only organization lacks.

### Alternative C: Archive Format (ZIP, TAR)

**Description:** Package policies as archives (ZIP, TAR) with a manifest inside.

**Pros:**

- Single file distribution
- Can include all resources
- Standard archive formats
- Easy to distribute

**Cons:**

- Must extract before use
- Less convenient for development
- Harder to version control
- More complex workflow

**Why not chosen:** Directory-based packs with manifest provide better developer experience while still supporting distribution. Archives can be created from packs when needed.

### Alternative D: Package Registry System

**Description:** Use a package registry (like npm, PyPI) for policy distribution.

**Pros:**

- Centralized distribution
- Version management
- Dependency resolution
- Discovery and search

**Cons:**

- Requires registry infrastructure
- More complex system
- Overkill for initial implementation
- May not be needed for all use cases

**Why not chosen:** Pack system provides foundation for future registry support without requiring it initially. Packs can be distributed via Git, file sharing, or future registries.

## Implementation Notes

- Pack files are named `sentrie.pack.toml`
- Pack discovery walks up directory tree from policy file location
- Pack files are parsed using `pelletier/go-toml`
- Pack metadata is available in the index system
- Permissions from pack files control TypeScript module capabilities
- Pack location is stored in the PackFile structure

## References

- Implementation: `pack/pack.go`, `loader/pack.go`
- Example: `example_pack/sentrie.pack.toml`
- Related: ADR-0001 (The Sentrie Language)
