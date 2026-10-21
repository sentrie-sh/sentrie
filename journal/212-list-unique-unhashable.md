## Binaek Sarkar (binaek89@gmail.com)

### 22:07 - Pick up #212
- Branch cut from `211-bound-cidr-expand`. #111 rewrites `builtin_collection_utils.go`'s error sites, so this stays in the chain 208 to 213 rather than branching from `main`.
- Plan: `list_unique` keys each exported element by a stable JSON encoding instead of using it as a Go map key, keeping first-occurrence order.
- _Model: claude-opus-5-5_
