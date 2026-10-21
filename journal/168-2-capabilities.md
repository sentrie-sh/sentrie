## Binaek Sarkar (binaek89@gmail.com)

### 14:47 - Start the capabilities layer
- Second layer of the #168 stack, cut from `168-1-nondeterminism-metadata`, which supplies the builtin call walker and the module metadata this layer builds on.
- Plan: a leaf `capabilities` package for the deployment-level JSON grant (allowed builtins, allowed `@sentrie/*` modules, whether local TypeScript modules are permitted); `--capabilities` on `exec`, `serve` and `validate`; an index-time check that fails with the offending span; paired runtime gates in builtin dispatch and in `js.Registry` `PrepareUse`/`LoadRequire`. An absent file grants everything, which is today's behaviour.
- _Model: claude-opus-5-5_
