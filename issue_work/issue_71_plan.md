# Issue 71 Implementation Plan

## Summary
Remove the `go mod tidy` hook from `.goreleaser.yaml` to prevent the worktree from becoming dirty during the release build, which causes the Go compiler to embed `vcs.modified=true` in the binary.

## Approach
The simplest and most reliable fix is to remove the `before.hooks` section that runs `go mod tidy`. This hook runs after GoReleaser's git dirty check but before the Go build, allowing `go.sum` modifications to taint the binary metadata.

Instead, `go mod tidy` should be run as part of normal development and committed before tagging a release. The CI test workflow already runs `go mod tidy` and fails if it produces changes, so this is already enforced.

### Alternatives Considered
- **Add CI pre-release check**: Run `go mod tidy` in CI before tagging to ensure `go.sum` is up-to-date. Not chosen because it adds complexity and the test workflow already enforces this.
- **Use ldflags to override version**: Explicitly set version via ldflags in GoReleaser. Not chosen because the current `runtime/debug.ReadBuildInfo()` approach is cleaner and the real issue is the dirty worktree.

## Files to Modify
- `.goreleaser.yaml` - Remove the `before.hooks` section

## Files to Create
None

## Implementation Steps
- [x] Remove `before.hooks` section from `.goreleaser.yaml`
- [x] Verify local build produces clean version
- [x] Run tests to ensure nothing is broken

Mark each step [x] after it is implemented and committed. This enables clear resume detection.

## Testing Strategy
- Unit tests: Run existing test suite (no new tests needed - this is a config change)
- Manual verification: Build locally from clean checkout and verify `--version` output is clean

## Risks and Mitigations
- **Risk**: Future `go.sum` drift could cause build issues
  - **Mitigation**: The test workflow already runs `go mod tidy` and fails on changes, so this is already enforced before code reaches main branch

## Success Criteria
- [ ] `.goreleaser.yaml` no longer has `before.hooks` section
- [ ] Local build from clean checkout shows clean version (no `+dirty` suffix)
- [ ] All tests pass

## Open Questions
None
