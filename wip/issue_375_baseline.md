# Issue 375 Baseline

## Environment
- Date: 2025-12-10
- Branch: feature/375-recipe-preview
- Base commit: 6b2aad135c57767d03361bc2fb1763f4b5da1b1d

## Test Results
- Total: 18 packages
- Passed: 17 packages
- Failed: 1 package (internal/builders)

## Build Status
Pass - CLI builds successfully

## Pre-existing Issues

The `internal/builders` package has failing tests in `TestLLMGroundTruth`:
- GitHub API rate limit exceeded (no GITHUB_TOKEN set)
- Pattern mismatch issues with ast-grep and trivy recipes

These failures are pre-existing and unrelated to this issue.
