# Issue 479 Implementation Plan

## Summary

Remove the legacy `Execute()` method from the Executor struct since all installation paths now use `ExecutePlan()`. This is a cleanup task that eliminates dead code after #478 wired up the plan-based flow.

## Approach

Direct removal of the `Execute()` method and its exclusive helper `verify()`. The method is only called from tests, so we'll update those tests to use `ExecutePlan()` instead, or remove them if they no longer serve a purpose (since they were testing the legacy flow).

### Alternatives Considered
- **Keep Execute() as deprecated**: Not chosen because it creates maintenance burden and the risk of accidental use
- **Keep verify() for future use**: Not chosen because verification is now handled at the install layer via `mgr.InstallWithOptions()`

## Files to Modify

- `internal/executor/executor.go` - Remove `Execute()` method and `verify()` helper
- `internal/executor/executor_test.go` - Update or remove tests that call `Execute(ctx)`

## Files to Create

None

## Implementation Steps

- [ ] 1. Remove the `Execute()` method from executor.go (lines 71-145)
- [ ] 2. Remove the `verify()` method from executor.go (lines 221-328)
- [ ] 3. Update `TestExecute_FallbackToDev` to test plan-based flow or remove if redundant
- [ ] 4. Update `TestExecute_NetworkFailureFallback` to test plan-based flow or remove if redundant
- [ ] 5. Run tests to verify no regressions: `go test ./...`
- [ ] 6. Run build to verify compilation: `go build ./cmd/tsuku`

Mark each step [x] after it is implemented and committed.

## Testing Strategy

- Unit tests: Ensure existing `ExecutePlan` tests pass
- Unit tests: Ensure `resolveVersionWith`, `shouldExecute`, `DryRun` tests still pass (they don't use `Execute()`)
- Integration tests: CI will verify all installation scenarios work

## Risks and Mitigations

- **Risk**: Tests that call `Execute()` will fail
  - **Mitigation**: Update those tests to use `ExecutePlan()` or remove them if they duplicate existing coverage
- **Risk**: Some code path still references `Execute()`
  - **Mitigation**: Grep search confirmed only test files call it; compiler will catch any missed references

## Success Criteria

- [ ] `Execute()` method removed from Executor
- [ ] `verify()` method removed (only used by Execute)
- [ ] All unit tests pass: `go test ./...`
- [ ] Build succeeds: `go build ./cmd/tsuku`
- [ ] CI integration tests pass

## Open Questions

None - the design is straightforward code deletion.
