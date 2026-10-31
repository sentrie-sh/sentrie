## Binaek Sarkar (binaek89@gmail.com)

### 14:38 - Pick up #271
- Inside a derive body, casts, typed lets, parameter types and the return type resolve shape names in the caller's policy, while the index resolves them in the derive's own scope.
- Picked up with #269, #270, #271, #272 and #273 in one batch; one pickup push and a draft PR per branch before implementation.
- No open design choice: use the derive's shape scope at every type-check site, the same way `isShapeScope` already does for `is` (#150).
- _Model: claude-opus-5-5_
