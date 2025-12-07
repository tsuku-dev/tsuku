# Issue 243 Summary

## Changes Made

Added `runtime_dependencies` overrides to 2 recipes that needed explicit documentation of their runtime behavior:

### ruff.toml
- Added `runtime_dependencies = []`
- Reason: Ruff is a compiled Rust binary distributed via PyPI. Although installed via `pipx_install`, it doesn't require Python at runtime.

### gore.toml
- Changed `dependencies = ["go"]` to `runtime_dependencies = ["golang"]`
- Reason: Gore is a Go REPL that needs Go installed at runtime to interpret code. The previous `dependencies` field was incorrect since it meant install-time dependency.

## Audit Results

### npm_install recipes (8 total)
All require Node.js at runtime - default behavior is correct:
- amplify, cdk, netlify-cli, serverless, serve, turbo, vercel, wrangler

### pipx_install recipes (4 total)
- black, httpie, poetry: Pure Python - need python-standalone at runtime (default)
- **ruff**: Compiled Rust binary - override added

### go_install recipes (8 total)
All compile to standalone binaries:
- cobra-cli, dlv, gofumpt, goimports, gopls, mockgen, staticcheck: No runtime deps (default)
- **gore**: Go REPL - override added for runtime Go dependency

### cargo_install recipes (3 total)
All compile to standalone binaries - default is correct:
- cargo-audit, cargo-edit, cargo-watch

## Test Results

- All unit tests pass
- Recipes validate with `--strict`
