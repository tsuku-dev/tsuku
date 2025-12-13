# Issue 472 Baseline

## Environment
- Date: 2025-12-13
- Branch: feature/472-expose-resolve-version
- Base commit: f94fc69a65609d0e8a20c15772d8db9a89404ed2

## Test Results
- Total: 32 packages
- Passed: 30
- Failed: 2 (pre-existing, unrelated to this issue)

### Pre-existing Failures
1. `internal/actions.TestNixRealizeAction_Execute_PackageFallback` - nil context panic in test setup
2. `internal/builders.TestLLMGroundTruth/L16_minikube` - LLM integration test (flaky, recipe generation difference)

## Build Status
Pass - `go build -o tsuku ./cmd/tsuku` succeeds without errors

## Pre-existing Issues
- The two test failures above are pre-existing and not related to this issue's scope
- Both failures are in areas unrelated to executor/version resolution
