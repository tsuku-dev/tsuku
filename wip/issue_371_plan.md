# Issue 371 Implementation Plan

## Summary

Add `llm.daily_budget` and `llm.hourly_rate_limit` configuration settings to the userconfig package, following the existing patterns for `llm.enabled` and `llm.providers`.

## Approach

Extend the existing `LLMConfig` struct with two new fields using pointer types to distinguish "not set" from explicit values. Follow the established pattern of using `AvailableKeys()`, `Get()`, and `Set()` methods.

### Alternatives Considered
- **Single "limits" struct**: Rejected - inconsistent with flat key pattern (`llm.key`)
- **Non-pointer types with sentinel values**: Rejected - current codebase uses pointers for optional fields

## Files to Modify
- `internal/userconfig/userconfig.go` - Add fields, constants, Get/Set logic
- `internal/userconfig/userconfig_test.go` - Add unit tests

## Files to Create
- None

## Implementation Steps
- [x] Add `DailyBudget` and `HourlyRateLimit` fields to `LLMConfig` struct
- [x] Add default constants `DefaultDailyBudget` and `DefaultHourlyRateLimit`
- [x] Add helper methods `LLMDailyBudget()` and `LLMHourlyRateLimit()` returning effective values
- [x] Update `AvailableKeys()` with new keys and descriptions
- [x] Update `Get()` to handle new keys
- [x] Update `Set()` to handle new keys with validation
- [x] Add unit tests for get/set of new fields
- [x] Add unit tests for file save/load of new fields
- [x] Add unit test for zero value behavior (disables limit)

## Testing Strategy
- Unit tests: Get/Set methods, file persistence, default values, zero-value behavior
- Manual verification: `tsuku config get/set llm.daily_budget` and `llm.hourly_rate_limit`

## Risks and Mitigations
- **TOML parsing edge cases**: Mitigation - follow existing pattern for pointer types with `omitempty`
- **Float precision in config display**: Mitigation - use `%g` format for clean display

## Success Criteria
- [ ] `tsuku config get llm.daily_budget` returns default (5.0) or configured value
- [ ] `tsuku config set llm.daily_budget 10` persists to config file
- [ ] `tsuku config get llm.hourly_rate_limit` returns default (10) or configured value
- [ ] `tsuku config set llm.hourly_rate_limit 20` persists to config file
- [ ] Setting `0` for either field is accepted (disables limit)
- [ ] All unit tests pass
- [ ] `go vet` and `golangci-lint` pass

## Open Questions
None - design document specifies all required details.
