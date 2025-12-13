# Issue 489 Implementation Plan

## Summary

Implement dependency tree discovery for Homebrew formulas that traverses the JSON API before LLM invocation, allowing users to see full cost estimates and confirm before proceeding.

## Approach

Follow the design in `docs/DESIGN-homebrew-builder.md` (Section: Dependency Handling). The implementation:
1. Adds `DependencyNode` struct and tree traversal logic
2. Uses visited set for diamond dependency handling
3. Provides topological sort (leaves first) for generation order
4. Integrates with existing `HomebrewBuilder` via new `BuildWithDependencies` method

### Alternatives Considered
- **Lazy discovery during LLM calls**: Rejected because it doesn't allow upfront cost estimation
- **Parallel API calls**: Could be added later, but sequential is simpler for initial implementation

## Files to Modify

- `internal/builders/homebrew.go` - Add dependency tree discovery types and methods
- `internal/builders/homebrew_test.go` - Add unit tests for tree traversal and topological sort

## Files to Create

None - all code goes in existing homebrew.go file per existing patterns.

## Implementation Steps

- [x] 1. Add `DependencyNode` struct and `RegistryChecker` interface
- [x] 2. Implement `DiscoverDependencyTree()` with visited set for diamond deps
- [x] 3. Implement `ToGenerationOrder()` for topological sort
- [x] 4. Add `WithRegistryChecker` option for dependency injection
- [x] 5. Implement helper functions for tree display and cost estimation
- [x] 6. Add `BuildWithDependencies()` method with user confirmation hook
- [x] 7. Unit tests for tree traversal with diamond dependencies
- [x] 8. Unit tests for topological sort correctness
- [x] 9. Integration-style test with mock server and multiple formulas

## Testing Strategy

- **Unit tests**:
  - `TestDependencyNode_ToGenerationOrder` - verify leaves-first ordering
  - `TestDiscoverDependencyTree_Simple` - formula with no deps
  - `TestDiscoverDependencyTree_Linear` - A -> B -> C chain
  - `TestDiscoverDependencyTree_Diamond` - A -> B,C; B,C -> D (shared dep)
  - `TestDiscoverDependencyTree_WithExistingRecipes` - some deps already have recipes
  - `TestDiscoverDependencyTree_APIError` - handle API failures gracefully

- **Integration test**:
  - Mock HTTP server returning formula JSONs
  - Mock registry checker
  - Verify full tree discovery and generation order

## Risks and Mitigations

- **API rate limiting**: The design uses sequential API calls; could add jitter/backoff if needed
- **Circular dependencies**: Homebrew formulas shouldn't have cycles, but visited set prevents infinite loops

## Success Criteria

- [x] `DiscoverDependencyTree()` correctly builds tree from Homebrew API
- [x] Diamond dependencies handled (shared deps discovered once)
- [x] `ToGenerationOrder()` returns correct topological order (leaves first)
- [x] Registry check identifies which deps already have recipes
- [x] Progress reporting during tree discovery
- [x] All tests pass
- [x] No regressions in existing homebrew builder tests

## Open Questions

None - design document provides clear specification.
