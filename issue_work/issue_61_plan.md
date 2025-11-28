# Issue 61 Implementation Plan

## Summary

Add `needs: matrix` and `if` condition to the `unit-tests` job to skip it when only docs change, same pattern as integration tests.

## Approach

Use the existing paths-filter output from the `matrix` job. Add:
1. `needs: matrix` to make unit-tests depend on matrix job
2. `if: ${{ needs.matrix.outputs.code == 'true' }}` to skip when docs-only

This matches the pattern used by `integration-linux` and `integration-macos` jobs.

### Alternatives Considered
- Separate paths-filter in unit-tests job: Duplicates work, adds latency
- Always run unit tests: Wastes ~2 minutes on docs-only PRs

## Files to Modify
- `.github/workflows/test.yml` - Add needs and if condition to unit-tests job

## Files to Create
None

## Implementation Steps
- [x] Add `needs: matrix` to unit-tests job
- [x] Add `if: ${{ needs.matrix.outputs.code == 'true' }}` to unit-tests job

## Testing Strategy
- This PR modifies a .yml file (code), so unit tests WILL run
- To verify the skip works, PR #60 (docs-only) can be rebased after merge

## Risks and Mitigations
- Risk: Breaking required status checks if unit-tests is required
- Mitigation: GitHub handles skipped jobs correctly for branch protection

## Success Criteria
- [ ] Unit tests skip for docs-only PRs
- [ ] Unit tests run for code PRs
- [ ] CI passes for this PR

## Open Questions
None
