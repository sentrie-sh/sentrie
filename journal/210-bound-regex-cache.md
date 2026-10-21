## Binaek Sarkar (binaek89@gmail.com)

### 22:03 - Pick up #210
- Branch cut from `209-regex-replace-first`. Both change `builtin_regex.go`, so this layer stacks on #209 to avoid a conflict. Third layer of the chain 208, 209, 210, 211, 212, 213.
- Plan: replace the unbounded global map with a `perch` LRU of a named-constant entry count, as #140 did for the executor caches.
- _Model: claude-opus-5-5_
