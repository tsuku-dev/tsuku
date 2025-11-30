# Issue 23 Implementation Plan

## Summary

Add configurable timeout handling for API requests to prevent the CLI from hanging indefinitely when network issues occur.

## Approach

Add `TSUKU_API_TIMEOUT` environment variable support with a default of 30 seconds. Apply this timeout to the version resolver's HTTP client (which makes GitHub, PyPI, npm, crates.io, RubyGems API calls). The existing 60-second timeout in the resolver is hardcoded - we'll make it configurable while keeping reasonable defaults.

### Alternatives Considered

1. **Per-provider timeouts**: Different timeouts for GitHub vs PyPI vs npm
   - Why not: Over-complicated for the problem; a single timeout is simpler

2. **Context-based timeout only**: Use context.WithTimeout at call sites
   - Why not: Already have http.Client.Timeout; would be redundant. Context timeouts are for cancellation (already done in #22)

3. **Create a new internal/http package**: Centralize all HTTP client creation
   - Why not: Over-engineering for this issue; can be done later if needed

## Files to Modify

- `internal/config/config.go` - Add environment variable constant and timeout getter
- `internal/version/resolver.go` - Use configurable timeout in newHTTPClient()
- `internal/registry/registry.go` - Use configurable timeout (optional, already has 30s)

## Files to Create

None - all changes are to existing files.

## Implementation Steps

- [ ] Step 1: Add timeout configuration to config package
- [ ] Step 2: Update version resolver to use configurable timeout
- [ ] Step 3: Update registry client to use configurable timeout (optional consolidation)
- [ ] Step 4: Add improved error messages for timeout errors
- [ ] Step 5: Add unit tests for timeout configuration
- [ ] Step 6: Test timeout behavior manually

Mark each step [x] after it is implemented and committed. This enables clear resume detection.

## Testing Strategy

- Unit tests: Test timeout configuration parsing and defaults
- Manual verification:
  1. Set `TSUKU_API_TIMEOUT=1s` and run `tsuku versions gh` - should timeout quickly
  2. Unset variable and verify 30s default works
  3. Set invalid value and verify error handling

## Risks and Mitigations

- **Risk**: Too aggressive timeout may fail for slow networks
  - **Mitigation**: Default 30 seconds is conservative; user can increase via env var

- **Risk**: Breaking change if default is different from current 60s
  - **Mitigation**: 30s is still generous; most API calls complete in <5s

## Success Criteria

- [ ] Default timeout of 30 seconds for API requests
- [ ] Configurable via TSUKU_API_TIMEOUT environment variable
- [ ] Clear error message when timeout occurs
- [ ] All existing tests pass
- [ ] No new linter warnings

## Open Questions

None - implementation approach is clear.
