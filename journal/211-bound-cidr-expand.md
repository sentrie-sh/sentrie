## Binaek Sarkar (binaek89@gmail.com)

### 22:05 - Pick up #211
- Branch cut from `210-bound-regex-cache`. #111 rewrites the error sites in `builtin_net.go`, and #213 (two layers up) changes the same file, so this stays in the chain 208 to 213 rather than branching from `main`.
- Plan: `cidrExpand` throws above a named 65,536-host limit, checked from the prefix length before any allocation.
- _Model: claude-opus-5-5_
