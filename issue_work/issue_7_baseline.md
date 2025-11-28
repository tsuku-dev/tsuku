# Issue 7 Baseline

## Environment
- Date: 2025-11-27
- Branch: chore/7-use-go-mod-version-in-ci
- Base commit: edeb55d6a290eb89ee62414938123220686163c5

## Test Results
- Total: 8 packages with tests
- Passed: 8
- Failed: 0

## Build Status
Pass - no warnings

## Coverage
- cmd/tsuku: 2.9%
- internal/actions: 25.8%
- internal/buildinfo: 90.0%
- internal/executor: 30.8%
- internal/install: 12.3%
- internal/recipe: 55.3%
- internal/version: 49.7%
- Command used: `go test -cover ./...`

## Pre-existing Issues
None - all tests pass, build succeeds, no vet warnings.
