# Issue 380 Summary

## What Was Implemented

Created structured error message templates for common LLM builder failures. Each error type provides clear what/why/how-to-fix information to help users recover from failures.

## Changes Made

- `internal/builders/errors.go`: New file with 6 error types:
  - `RateLimitError` - Hourly LLM generation limit exceeded
  - `BudgetError` - Daily LLM cost budget exhausted
  - `GitHubRateLimitError` - GitHub API rate limit with retry time
  - `GitHubRepoNotFoundError` - Repository not found
  - `LLMAuthError` - API authentication failure
  - `ValidationError` - Recipe validation failed after repairs

- `internal/builders/errors_test.go`: Comprehensive tests for all error types

- `internal/builders/github_release.go`: Updated `fetchReleases` and `fetchRepoMeta` to use new error types for 404, 403, and 429 status codes

## Key Decisions

- **Followed existing patterns**: Used the same error handling patterns as `version/errors.go` and `registry/errors.go` with `Suggestion()` method for actionable advice.

- **GitHub rate limit parsing**: Parse `X-RateLimit-Reset` header to provide accurate retry times.

- **Separate error types for GitHub**: Created distinct `GitHubRateLimitError` and `GitHubRepoNotFoundError` instead of a single generic error to provide targeted suggestions.

## Trade-offs Accepted

- **No rate limit tracking on builder**: The `Authenticated` field checks `os.Getenv("GITHUB_TOKEN")` at error time rather than storing token presence on the builder. This is simpler and avoids stale state.

## Test Coverage

- New tests added: 6 test functions covering all error types
- Coverage: All Error() and Suggestion() methods tested

## Known Limitations

- `RateLimitError` and `BudgetError` are defined but not yet used - they require integration with the rate limiting/budget checking code in future PRs.

## Future Improvements

- Integrate `RateLimitError` and `BudgetError` with the state manager's rate limiting
- Add validation error types for specific failure modes (binary not found, checksum mismatch)
