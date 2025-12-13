# Issue 442 Baseline

## Environment
- Date: 2025-12-13
- Branch: feature/442-deterministic-flag
- Base commit: 1c5b055 (main)

## Test Results
- Total: 21 packages tested
- Passed: 20
- Failed: 1 (pre-existing)

## Pre-existing Test Failure

`TestPrimitives` in `internal/actions/decomposable_test.go:96` fails:
```
len(Primitives()) = 13, want 12
```

This is because recent PRs (#464, #465) added `gem_exec` and `pip_install` ecosystem primitives but the test's expected count wasn't updated. The test expects 11 but we now have 13 primitives (8 core + 5 ecosystem).

**Note**: This test failure exists on main and is not related to issue #442 work. Will fix as part of implementation since it's in the same file.

## Build Status
Build succeeded (go build -o tsuku ./cmd/tsuku)

## Current Primitives (13)

Core (8):
- download, extract, chmod, install_binaries
- set_env, set_rpath, link_dependencies, install_libraries

Ecosystem (5):
- cargo_build, gem_exec, go_build, npm_exec, pip_install
