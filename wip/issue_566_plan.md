# Issue 566 Implementation Plan

## Summary

Refactor action metadata (determinism and dependencies) from static maps to interface methods on action types, using a BaseAction embedded struct to provide defaults and reduce boilerplate.

## Approach

The issue proposes using interface methods instead of static maps to co-locate metadata with action implementations. This is Go-idiomatic and provides compile-time safety via interface implementation checks.

Key design decisions:
1. Create `ActionMetadata` interface with `IsDeterministic()` and `Dependencies()` methods
2. Create `BaseAction` struct with default implementations (non-deterministic, no deps)
3. Embed `BaseAction` in all action types
4. Actions override methods only when needed (non-default values)
5. Update callers (`IsDeterministic()`, `GetActionDeps()`) to use interface methods
6. Remove static maps after migration

### Alternatives Considered

- **Keep static maps alongside interface**: Rejected - duplicates information, doesn't solve the sync problem
- **Require all methods on all actions**: Rejected - excessive boilerplate for actions with default values

## Files to Modify

- `internal/actions/action.go` - Add `ActionMetadata` interface and `BaseAction` struct
- `internal/actions/decomposable.go` - Update `IsDeterministic()` to use interface; remove `deterministicActions` map
- `internal/actions/dependencies.go` - Update `GetActionDeps()` to use interface; remove `ActionDependencies` map
- `internal/actions/*.go` (33 action files) - Embed `BaseAction`, add override methods where needed

## Files to Create

None - all changes are in existing files.

## Implementation Steps

- [x] Add `ActionMetadata` interface and `BaseAction` struct to `action.go`
- [ ] Update action types to embed `BaseAction` and implement metadata methods
- [ ] Update `IsDeterministic()` in `decomposable.go` to use interface methods
- [ ] Update `GetActionDeps()` in `dependencies.go` to use interface methods
- [ ] Remove `deterministicActions` map from `decomposable.go`
- [ ] Remove `ActionDependencies` map from `dependencies.go`
- [ ] Update tests to verify interface-based behavior

Mark each step [x] after it is implemented and committed. This enables clear resume detection.

## Testing Strategy

- Unit tests: Existing tests in `decomposable_test.go` and `dependencies_test.go` verify the API still works
- The tests that check all registered actions have entries will be replaced with compile-time checks
- Add test to verify BaseAction defaults

## Risks and Mitigations

- **Risk**: Forgetting to embed BaseAction in new actions
  - **Mitigation**: The Action interface now includes metadata methods; compile fails without them

- **Risk**: Performance regression from interface method calls vs map lookup
  - **Mitigation**: Negligible - both are O(1) and these methods are called infrequently

## Success Criteria

- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds (`go build ./cmd/tsuku`)
- [ ] No static maps remain for action metadata
- [ ] Each action's metadata is co-located with its implementation
- [ ] Compiler enforces that all actions have metadata (via interface)

## Open Questions

None - the approach is clear from the issue description.
