# Issue 489 Baseline

## Environment
- Date: 2025-12-13
- Branch: feature/489-dependency-tree-discovery
- Base commit: f48484be8952d1ba45537ab1e68fb7aa8e383bb0

## Test Results
- Total: ~260 test functions (5218 RUN/PASS/FAIL lines including subtests)
- Passed: All except 1 pre-existing failure
- Failed: 1 (pre-existing)

## Build Status
- Build: PASS (no warnings)

## Pre-existing Issues

### TestNixRealizeAction_Execute_PackageFallback
- Location: `internal/actions/nix_realize_test.go`
- Cause: Test passes nil Context to `exec.CommandContext`, panics
- Note: This is an environment-specific test for nix-portable functionality
- Not related to this issue's scope (Homebrew builder)
