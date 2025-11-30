# Implementation Plan: Issue #21

## Summary

Handle GitHub API rate limits gracefully with clear feedback and actionable suggestions.

## Analysis

### Current State
- `internal/version/errors.go`: Has `ErrTypeRateLimit` defined but generic suggestion
- `internal/version/resolver.go`: Uses `go-github` library, handles network errors but not rate limits
- `go-github` provides `RateLimitError` type with `Rate` struct (Limit, Remaining, Reset)

### Key Requirements
1. Detect rate limit errors from GitHub API
2. Display rate limit info (limit, reset time)
3. Suggest using GITHUB_TOKEN
4. Suggest using specific version as fallback

## Implementation

### 1. Create GitHubRateLimitError type in errors.go

Add a specialized error type that captures GitHub rate limit details:
- Limit (requests per hour)
- Remaining requests
- Reset time
- Whether authenticated

The `Suggestion()` method will return context-aware suggestions.

### 2. Update resolver.go to detect and wrap rate limit errors

In `ResolveGitHub`, `resolveFromTags`, and `ListGitHubVersions`:
- Use `errors.As` to detect `github.RateLimitError`
- Create `GitHubRateLimitError` with rate info
- Return wrapped error

### 3. Tests

Add tests for:
- GitHubRateLimitError formatting
- Suggestion generation
- Time formatting (reset time display)

## Files to Modify

1. `internal/version/errors.go` - Add GitHubRateLimitError
2. `internal/version/errors_test.go` - Add tests
3. `internal/version/resolver.go` - Detect and wrap rate limit errors

## Out of Scope (Future)

- Caching version lists (would require state management)
- Warning when approaching limit (would require exposing rate info on success)
