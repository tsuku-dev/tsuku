# Issue 76 Implementation Plan

## Summary
Move slow cargo/rust integration tests (T18, T27) from PR CI to a daily scheduled workflow to reduce PR feedback time.

## Approach
1. Update `test-matrix.json` to separate slow tests into a `ci.scheduled` list
2. Modify `test.yml` workflow to exclude scheduled tests from PR runs
3. Create a new `scheduled-tests.yml` workflow that runs daily and includes all tests

### Alternatives Considered
- **Run slow tests only on main branch**: Rejected - we want them on a schedule regardless of merges
- **Skip slow tests entirely**: Rejected - we still want coverage, just not blocking PRs

## Files to Modify
- `test-matrix.json` - Add `ci.scheduled` list with T18, T27
- `.github/workflows/test.yml` - Filter out scheduled tests from PR matrix

## Files to Create
- `.github/workflows/scheduled-tests.yml` - Daily workflow running full test suite

## Implementation Steps
- [x] Update test-matrix.json to add ci.scheduled list
- [x] Modify test.yml to exclude scheduled tests from PR runs (done via test-matrix.json)
- [x] Create scheduled-tests.yml workflow for daily runs
- [x] Verify workflows are syntactically correct

## Testing Strategy
- Syntax validation: YAML linting
- Manual verification: Check GitHub Actions UI after push

## Risks and Mitigations
- **Risk**: Bugs in cargo tests go unnoticed for up to 24 hours
  - **Mitigation**: GitHub Actions sends email on workflow failure by default

## Success Criteria
- [ ] PR workflow no longer includes T18 and T27
- [ ] Scheduled workflow is created with daily cron
- [ ] All workflow syntax is valid

## Open Questions
None
