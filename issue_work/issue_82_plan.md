# Issue 82 Implementation Plan

## Summary

Create `internal/telemetry` package with Event struct, Client for fire-and-forget HTTP sending, and environment variable opt-out support.

## Approach

Simple package with minimal dependencies. Use standard library `net/http` for HTTP client, `encoding/json` for marshaling. Follow existing patterns from `internal/config` for environment variable handling.

### Alternatives Considered

- Using a third-party HTTP client: Not needed, stdlib is sufficient for fire-and-forget
- Singleton client: Rejected in favor of explicit client creation for testability

## Files to Create

- `internal/telemetry/event.go` - Event struct and constructor helpers
- `internal/telemetry/event_test.go` - Event construction tests
- `internal/telemetry/client.go` - Client struct with Send method
- `internal/telemetry/client_test.go` - Client tests with mock HTTP server

## Implementation Steps

- [x] Create Event struct with all schema fields
- [x] Create constructor helpers (NewInstallEvent, NewUpdateEvent, NewRemoveEvent)
- [x] Create Client struct with disabled/debug flags
- [x] Implement NewClient() that checks env vars
- [x] Implement Send() with fire-and-forget goroutine
- [x] Add unit tests for Event construction
- [x] Add unit tests for Client (mock HTTP server)
- [x] Add unit tests for env var opt-out logic

## Testing Strategy

- Unit tests: Event construction, client initialization, opt-out logic
- Mock HTTP server for Send() tests
- Verify timeout behavior
- Verify debug mode prints to stderr

## Risks and Mitigations

- Goroutine leaks: Use context with timeout to ensure cleanup
- Race conditions: Client is immutable after creation, Event is passed by value

## Success Criteria

- [ ] All acceptance criteria from issue #82 met
- [ ] Tests pass
- [ ] Build succeeds

## Open Questions

None
