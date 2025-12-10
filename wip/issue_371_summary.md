# Issue 371 Summary

## What Was Implemented

Added configurable settings for LLM daily budget and hourly rate limit to user configuration, enabling cost controls for LLM-powered recipe generation.

## Changes Made

- `internal/userconfig/userconfig.go`:
  - Added `DailyBudget` and `HourlyRateLimit` pointer fields to `LLMConfig` struct
  - Added `DefaultDailyBudget` (5.0 USD) and `DefaultHourlyRateLimit` (10) constants
  - Added `LLMDailyBudget()` and `LLMHourlyRateLimit()` helper methods
  - Updated `Get()` to return values for `llm.daily_budget` and `llm.hourly_rate_limit`
  - Updated `Set()` to parse and validate values (must be non-negative)
  - Updated `AvailableKeys()` with descriptions for new settings

- `internal/userconfig/userconfig_test.go`:
  - Added tests for Get/Set of new fields
  - Added tests for default value behavior
  - Added tests for file save/load persistence
  - Added tests for zero value (disables limit)
  - Added tests for invalid value handling

## Key Decisions

- **Pointer types for optional fields**: Consistent with existing `LLM.Enabled` pattern, allows distinguishing "not set" (use default) from "explicitly set to 0" (disabled)
- **Non-negative validation**: Negative budgets/limits make no sense; validation rejects them with clear error messages
- **Zero means disabled**: Following the design doc, setting 0 disables the respective limit rather than erroring

## Trade-offs Accepted

- **No upper bound validation**: Users can set arbitrarily high limits; acceptable since this is user-controlled cost protection

## Test Coverage

- New tests added: 11
- All tests pass (37 total in userconfig package)

## Known Limitations

- Configuration values are per-machine (stored in config file, not synced)
- This issue only adds config storage; enforcement is implemented in dependent issues (#378, #379)

## Future Improvements

- Could add `tsuku config describe llm.daily_budget` for detailed help on each setting
