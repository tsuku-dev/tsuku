# Issue 329 Implementation Plan

## Summary

Implement a provider factory that manages Claude and Gemini providers with per-provider circuit breakers, auto-detecting available providers from environment variables and supporting automatic failover.

## Approach

Following the design doc (DESIGN-llm-slice-3-repair-loop.md), the factory:
1. Auto-detects available providers based on environment variables
2. Maintains per-provider circuit breakers for independent failure domains
3. Returns the primary provider if available, otherwise falls back to secondary
4. Provides methods to report success/failure for circuit breaker updates

### Alternatives Considered
- Global circuit breaker: Rejected because it defeats the purpose of having multiple providers - a failure on one would block both
- Simple failover without breaker: Rejected because every request to a dead provider incurs timeout cost

## Files to Create
- `internal/llm/factory.go` - Provider factory with circuit breakers

## Implementation Steps
- [ ] Create Factory struct with providers map and breakers map
- [ ] Implement NewFactory() with auto-detection of providers from env vars
- [ ] Implement GetProvider() with circuit breaker checking and failover logic
- [ ] Implement ReportSuccess() and ReportFailure() for breaker updates
- [ ] Implement AvailableProviders() helper method
- [ ] Add comprehensive unit tests with mock providers

## Testing Strategy
- Unit tests: Mock providers to test factory logic without API calls
- Test cases:
  - Factory with no providers returns error
  - Factory with single provider returns that provider
  - Factory with both providers returns primary
  - GetProvider respects circuit breaker state
  - Failover when primary breaker is open
  - ReportSuccess/ReportFailure update correct breaker
  - AvailableProviders returns providers with closed/half-open breakers

## Risks and Mitigations
- Risk: GeminiProvider requires context for initialization (NewGeminiProvider(ctx))
  - Mitigation: Pass context to NewFactory or use background context for initialization

## Success Criteria
- [ ] Factory auto-detects available providers from env vars
- [ ] GetProvider returns available provider respecting breaker state
- [ ] ReportSuccess and ReportFailure update circuit breakers
- [ ] Falls back to secondary provider when primary trips
- [ ] Returns error when no providers available
- [ ] Unit tests with mock providers achieve >80% coverage

## Open Questions
None - design is clear from the design document.
