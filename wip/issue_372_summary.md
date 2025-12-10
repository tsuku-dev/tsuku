# Issue 372 Summary

## What Was Implemented

Added LLM usage tracking to `internal/install/state.go` for rate limiting and daily budget enforcement. This enables the LLM builder feature to track generation history and costs, preventing runaway usage.

## Changes Made

- `internal/install/state.go`:
  - Added `LLMUsage` struct with `GenerationTimestamps`, `DailyCost`, and `DailyCostDate` fields
  - Added `LLMUsage` field to `State` struct (optional, backward compatible)
  - Added `RecordGeneration` method: records timestamp and cost, prunes old entries, resets daily cost at UTC midnight
  - Added `CanGenerate` method: checks both rate limit and budget constraints
  - Added `DailySpent` method: returns current day's total cost
  - Added `RecentGenerationCount` method: returns count of generations in last hour

- `internal/install/state_test.go`:
  - Added 13 new test functions covering all LLM usage tracking functionality
  - Tests for recording, accumulating, rate limiting, budget enforcement, concurrent access, and backward compatibility

## Key Decisions

- **Extended existing state rather than new package**: The `StateManager` already handles JSON persistence, file locking, and concurrent access. Adding LLM tracking to this established infrastructure avoids code duplication and coordination issues.
- **Optional LLMUsage field**: Uses pointer with `omitempty` tag for backward compatibility. Old state files without `llm_usage` load correctly with `nil` value.
- **UTC for date comparisons**: All daily cost tracking uses UTC to ensure consistent behavior regardless of local timezone.
- **Zero means unlimited**: Rate limit of 0 means no limit; budget of 0 means no budget cap. This allows users to disable restrictions if desired.

## Trade-offs Accepted

- **Timestamps stored as UTC**: Local time differences could cause unexpected reset times for users, but UTC provides consistent cross-timezone behavior.
- **Timestamp pruning on write only**: Old timestamps are pruned when `RecordGeneration` is called, not on read. This is acceptable since rate checking happens immediately before recording.

## Test Coverage

- New tests added: 13 test functions
- Coverage: All new methods tested including error paths and edge cases
- Concurrent access tested with 10 goroutines performing 5 operations each

## Known Limitations

- Clock manipulation can bypass rate limiting (documented as acceptable for CLI tool)
- Per-machine tracking only, not per-user (acceptable for single-user CLI tool)
- Timestamp pruning happens at write time only (sufficient for normal usage patterns)

## Future Improvements

- Integration with userconfig for budget/rate limit settings (Issue #371)
- Rate limiting enforcement in create command (Issue #378)
- Daily budget enforcement in create command (Issue #379)
