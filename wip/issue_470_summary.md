# Issue 470 Summary

## What Was Implemented

Created the foundational plan cache infrastructure for the deterministic execution feature. This includes types and functions for identifying cached plans, validating them, and reporting checksum mismatches with user-friendly error messages.

## Changes Made
- `internal/executor/plan_cache.go`: New file with:
  - `PlanCacheKey` struct for uniquely identifying cached plans
  - `CacheKeyFor()` function to generate cache keys from resolution output
  - `ValidateCachedPlan()` function to validate cached plans against current state
  - `ChecksumMismatchError` type with detailed error message and recovery instructions
- `internal/executor/plan_cache_test.go`: Comprehensive unit tests for all new types and functions

## Key Decisions
- **Cache key based on resolution output**: The cache key uses the resolved version, not the user's input. This means "ripgrep" and "ripgrep@14.1.0" that resolve to the same version share the same cache.
- **Include tool and version in ChecksumMismatchError**: The error message includes a specific recovery command (`tsuku install <tool>@<version> --fresh`) instead of a generic placeholder.
- **Use strings.Cut for platform parsing**: Safer than strings.Index for parsing "os-arch" format, handles edge cases gracefully.

## Trade-offs Accepted
- **ValidateCachedPlan takes InstallationPlan**: This requires conversion from install.Plan when validating cached plans, but keeps the validation logic in the executor package where it belongs.

## Test Coverage
- New tests added: 6 test functions with multiple subtests
- Tests cover: CacheKeyFor generation, ValidateCachedPlan with various mismatch scenarios, ChecksumMismatchError message formatting

## Known Limitations
- None for this foundation issue. The types and functions are designed to be used by downstream issues (#471-#479).

## Future Improvements
- Downstream issues will integrate these types into the actual plan caching flow.
