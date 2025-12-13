# Issue 477 Implementation Plan

## Summary

Implement `getOrGeneratePlan()` function in `cmd/tsuku/install_deps.go` that orchestrates the two-phase plan retrieval flow: version resolution (always), cache lookup (unless --fresh), and plan generation (if needed).

## Approach

Follow the design from `docs/DESIGN-deterministic-execution.md`. The function:
1. Always calls `exec.ResolveVersion()` to get resolved version
2. Generates cache key from resolution output
3. Checks cache using `stateMgr.GetCachedPlan()` (unless `--fresh`)
4. Validates cached plan using `executor.ValidateCachedPlan()`
5. Converts cached plan using `executor.FromStoragePlan()`
6. On miss/invalid: calls `exec.GeneratePlan()`

### Alternatives Considered
- Put orchestration in executor package: Not chosen because design explicitly calls for keeping executor focused on execution
- Export computeRecipeHash: Not needed - can compute hash directly in install_deps.go using recipe.ToTOML()

## Files to Modify
- `cmd/tsuku/install_deps.go` - Add `getOrGeneratePlan()` function and `planRetrievalConfig` struct

## Files to Create
- `cmd/tsuku/install_deps_test.go` - Unit tests with mocked executor and state manager

## Implementation Steps
- [ ] Add `planRetrievalConfig` struct definition
- [ ] Add helper function `computeRecipeHash()` for hashing recipe TOML
- [ ] Implement `getOrGeneratePlan()` function
- [ ] Add unit tests for cache hit scenario
- [ ] Add unit tests for cache miss scenario
- [ ] Add unit tests for --fresh flag bypassing cache
- [ ] Add unit tests for invalid cached plan scenarios
- [ ] Verify all tests pass and update design doc

## Testing Strategy
- Unit tests: Mock executor and state manager to test all code paths
- Test cases:
  - Cache hit with valid plan: returns cached plan without calling GeneratePlan
  - Cache miss: calls GeneratePlan
  - --fresh flag: skips cache lookup, calls GeneratePlan
  - Invalid cached plan (format version mismatch): regenerates
  - Invalid cached plan (platform mismatch): regenerates
  - Invalid cached plan (recipe hash mismatch): regenerates
  - Version resolution error: propagates error

## Risks and Mitigations
- Risk: Interface mismatch with existing tests
  - Mitigation: Follow existing patterns in install_deps.go and install_test.go

## Success Criteria
- [ ] `getOrGeneratePlan()` function implemented
- [ ] `planRetrievalConfig` struct defined
- [ ] All test scenarios covered
- [ ] `go test ./cmd/tsuku/...` passes
- [ ] Design doc updated to mark #477 as done

## Open Questions
None - design doc provides clear specification.
