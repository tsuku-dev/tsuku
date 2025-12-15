# Issue 580 Baseline

## Environment
- Date: 2024-12-14
- Branch: chore/580-rename-homebrew-bottle
- Base commit: 9ffd58126688f87de8dfe32ebcede35d8e3446b3

## Test Results
- Total: 25 packages
- Passed: 23 packages
- Failed: 2 packages (pre-existing issues, not related to this work)

## Build Status
Pass - `go build ./...` succeeds with no errors

## Pre-existing Issues
The following test failures are pre-existing and unrelated to this refactoring:

1. **internal/builders** - LLM integration tests fail due to:
   - Missing API credits for Anthropic API calls
   - Missing testdata files: `readline-source.toml`, `python-source.toml`, `bash-source.toml`

2. **internal/llm** - Integration tests fail due to:
   - Missing API credits for Anthropic API calls

These failures are infrastructure-related (API credits) or missing test fixtures, not code issues.
