# Issue 401 Baseline

## Environment
- Date: 2025-12-10
- Branch: feature/401-installation-plan-types
- Base commit: 203b0f2650e86e920c7aa442ef24283ebc6204e0

## Test Results
- Passed: 17 packages
- Failed: 2 packages (pre-existing)
  - `internal/builders`: TestLLMGroundTruth - LLM-based test with non-deterministic output
  - `internal/validate`: TestCleaner_CleanupStaleLocks - Local filesystem permission issue

## Build Status
Pass - `go build ./cmd/tsuku` succeeds with no warnings

## Pre-existing Issues
- TestLLMGroundTruth failures are expected due to LLM output variability
- TestCleaner_CleanupStaleLocks failure is a local environment issue (stale temp directories with incorrect permissions)
