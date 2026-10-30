## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #117
- `main` listens for `os.Interrupt` and the uncatchable `os.Kill`, not `SIGTERM`, and `StopServer` closes connections instead of draining them.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_
