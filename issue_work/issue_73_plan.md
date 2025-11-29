# Issue 73 Implementation Plan

## Summary
Add a CI step to test.yml that verifies `go mod tidy` produces no changes, preventing untidy module files from being merged.

## Approach
Add a "Verify go.mod is tidy" step to the unit-tests job that runs early (before tests). This catches issues before expensive test runs.

### Alternatives Considered
- Separate job: Would add overhead for a simple check
- Run after tests: Would waste time if go.mod is untidy

## Files to Modify
- `.github/workflows/test.yml` - Add verification step to unit-tests job

## Implementation Steps
- [ ] Add "Verify go.mod is tidy" step after "Download dependencies" in unit-tests job
- [ ] Step runs `go mod tidy` then `git diff --exit-code go.mod go.sum`
- [ ] Include clear error message if check fails

## Testing Strategy
- Manual verification: Create a PR with intentionally untidy go.mod to confirm CI fails
- Verify current codebase passes (go.mod should already be tidy)

## Risks and Mitigations
- Go version differences could cause tidy behavior differences: CI pins Go version via go.mod, so this is mitigated

## Success Criteria
- [ ] CI fails if `go mod tidy` produces any diff
- [ ] Clear error message tells contributor to run `go mod tidy` locally
- [ ] Check runs on all PRs targeting main
