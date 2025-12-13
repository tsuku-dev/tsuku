# Issue 473 Baseline

## Environment
- Date: 2025-12-13
- Branch: feature/473-execute-plan-checksum
- Base commit: 2bb44409dd7db48bd5c24717d4650e2284de40c3

## Test Results
- Total: 20 packages
- Passed: 19 packages
- Failed: 1 package (internal/actions)

## Build Status
Pass - `go build -o tsuku ./cmd/tsuku` succeeds without errors

## Pre-existing Issues
- `TestNixRealizeAction_Execute_PackageFallback` in `internal/actions/nix_realize_test.go` panics with nil Context
- This is a pre-existing issue from feature/448-nix-realize-primitive branch work, unrelated to this issue
- The executor package tests pass cleanly
