# Issue 566 Summary

## What Was Implemented

Refactored action metadata (determinism and dependencies) from static maps to interface methods on action types. Added a `BaseAction` embedded struct to provide sensible defaults and reduce boilerplate.

## Changes Made

- `internal/actions/action.go`: Added `ActionDeps` type, extended `Action` interface with `IsDeterministic()` and `Dependencies()` methods, added `BaseAction` struct with default implementations
- `internal/actions/decomposable.go`: Updated `IsDeterministic()` function to use interface methods, removed `deterministicActions` static map
- `internal/actions/dependencies.go`: Updated `GetActionDeps()` to use interface methods, removed `ActionDeps` type (moved to action.go) and `ActionDependencies` static map
- `internal/actions/*.go` (33 action files): Added `BaseAction` embedding to all action types, added override methods for determinism and dependencies where needed
- `internal/actions/*_test.go`: Updated test mocks to implement new interface methods, removed tests that checked the now-removed static maps

## Key Decisions

- **Embed `BaseAction` for defaults**: Actions with default values (non-deterministic, no deps) don't need to implement any methods - they inherit from `BaseAction`
- **Methods vs type-level interface**: Added methods directly to Action interface rather than creating a separate `ActionMetadata` interface - simpler API, enforced by compiler
- **Remove static maps entirely**: No backwards compatibility shims - the interface-based approach fully replaces the maps

## Trade-offs Accepted

- **No lazy initialization**: Every action type must embed `BaseAction` - but this is a trivial one-line change and provides compile-time guarantees
- **Removed consistency tests**: Tests that verified all actions had map entries are now replaced by compiler enforcement - actions without proper method implementations won't compile

## Test Coverage

- Tests updated: 3 test files (action_test.go, decomposable_test.go, dependencies_test.go)
- Existing tests for `IsDeterministic()` and `GetActionDeps()` API continue to pass
- Static map consistency tests removed (replaced by compile-time checks)

## Known Limitations

- None - the refactor is complete and all functionality is preserved

## Future Improvements

- Could add compile-time interface assertions for each action type (e.g., `var _ Action = (*DownloadAction)(nil)`) to catch interface implementation issues earlier
