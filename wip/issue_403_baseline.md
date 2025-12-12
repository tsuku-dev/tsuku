# Issue 403 Baseline

## Environment
- Date: 2025-12-12T21:20:00Z
- Branch: feature/403-eval-command
- Base commit: f0cd9899a19c4119fd7c9de2919b4497017160ab

## Test Results
- Total packages: 20
- Passed: 18
- Failed: 2 (pre-existing, unrelated to this issue)

### Pre-existing Failures
1. `internal/builders` - LLM ground truth tests (trivy, k9s)
   - Generated recipes vary from expected patterns
   - Not deterministic due to LLM response variation

2. `internal/validate` - TestCleaner_CleanupStaleLocks
   - Permission denied errors on temp directories
   - Local environment issue (stale temp files)

## Build Status
Build successful - `go build -o tsuku ./cmd/tsuku`

## Coverage
Not tracked for baseline - will add tests for new code.

## Pre-existing Issues
- LLM ground truth tests are known to be flaky due to LLM response variations
- Temp directory cleanup failures are environment-specific
