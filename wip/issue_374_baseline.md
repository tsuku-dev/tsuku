# Issue 374 Baseline

## Environment
- Date: 2025-12-10
- Branch: feature/374-create-yes-flag
- Base commit: 6b2aad135c57767d03361bc2fb1763f4b5da1b1d

## Test Results
- Total: 17 packages
- Passed: 16 packages
- Failed: 1 package (internal/builders - TestLLMGroundTruth)

## Build Status
Pass - no warnings

## Pre-existing Issues
- `TestLLMGroundTruth` in `internal/builders` fails on LLM integration tests (ast-grep, trivy, k9s)
- This is an integration test for LLM-generated recipes and is unrelated to issue 374
- The test requires external API calls and produces non-deterministic results
