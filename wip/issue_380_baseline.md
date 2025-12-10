# Issue 380 Baseline

## Environment
- Date: 2025-12-10
- Branch: feature/380-actionable-error-templates
- Base commit: 31b92fe8d8387f67a506fd4c57d7366c5dd9b41d

## Test Results
- Total: 19 packages tested
- Passed: 19 packages
- Failed: 0 (excluding TestLLMGroundTruth - see Pre-existing Issues)

## Build Status
- `go build -o tsuku ./cmd/tsuku`: Pass
- `go vet ./...`: Pass

## Coverage
Not tracked for this baseline.

## Pre-existing Issues
- `TestLLMGroundTruth` in `internal/builders` fails due to LLM response variability
