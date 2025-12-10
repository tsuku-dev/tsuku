# Issue 378 Implementation Plan

## Summary

Enforce hourly rate limit (default 10/hour) for LLM recipe generation by integrating existing `CanGenerate` and `RecordGeneration` methods from the state package with the create command.

## Approach

Use the existing infrastructure from #371 (userconfig) and #372 (state tracking) to enforce rate limits in the create command. The rate limit check happens before starting LLM generation, and recording happens after successful generation.

### Alternatives Considered

- Add rate limiting inside the GitHub builder: Would make the builder aware of rate limits, but violates separation of concerns. The command should manage rate limits, not the builder.
- Use middleware pattern: Overkill for a single integration point.

## Files to Modify

- `cmd/tsuku/create.go` - Add rate limit check before LLM generation and recording after success

## Implementation Steps

- [ ] Import `install` and `userconfig` packages in create.go
- [ ] Add rate limit check before creating GitHubReleaseBuilder (only for GitHub source)
- [ ] Add `RecordGeneration` call after successful build (only for GitHub source)
- [ ] Add helper function to format wait time for error message

## Testing Strategy

- Unit tests: The rate limiting logic is already tested in `internal/install/state_test.go`
- Manual testing: Run `tsuku create` with GitHub source and verify rate limit enforcement
- Manual testing: Set `llm.hourly_rate_limit` to 0 and verify no enforcement

## Risks and Mitigations

- Risk: Race condition if multiple `tsuku create` commands run concurrently - Mitigation: State file has locking already
- Risk: Cost tracking not implemented yet - Mitigation: Pass 0 as cost for now; a future issue can add actual cost tracking

## Success Criteria

- [ ] Rate limit checked before LLM generation starts
- [ ] Error message shows count, limit, and wait time
- [ ] Error suggests increasing limit via `tsuku config set`
- [ ] Setting rate limit to 0 disables enforcement
- [ ] Existing tests pass
