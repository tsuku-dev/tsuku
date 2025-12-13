# Issue 472 Summary

## What Was Implemented

Added a public `ResolveVersion` method to the Executor that exposes version resolution as Phase 1 of the two-phase installation model. This enables orchestration code to resolve versions independently from plan generation, which is required for cache key computation.

## Changes Made

- `internal/executor/executor.go`: Added `ResolveVersion(ctx context.Context, constraint string) (string, error)` method that:
  - Creates a version resolver internally
  - Uses the provider factory to get the appropriate version provider from the recipe
  - Returns the resolved version string (e.g., "14.1.0")
  - Handles empty constraint (resolves to latest) and specific version constraints

- `internal/executor/executor_test.go`: Added 4 new unit tests:
  - `TestResolveVersion_EmptyConstraint`: Tests resolving latest version
  - `TestResolveVersion_SpecificConstraint`: Tests resolving a specific version
  - `TestResolveVersion_UnknownSource`: Tests error handling for unknown sources
  - `TestResolveVersion_NoVersionSource`: Tests error when no version source is configured

## Key Decisions

- **Return string, not VersionInfo**: The method returns just the version string rather than the full `VersionInfo` struct, matching the design doc specification and the orchestration code's needs for cache key computation.

- **Create resolver internally**: The method creates a fresh resolver internally rather than accepting one as a parameter, keeping the public API simple and avoiding exposure of internal details.

## Trade-offs Accepted

- **Resolver creation per call**: Each call to `ResolveVersion` creates a new resolver. This is acceptable because version resolution is typically called once per installation, and the resolver is lightweight.

## Test Coverage

- New tests added: 4
- All existing tests continue to pass
- Tests follow existing patterns in executor_test.go

## Known Limitations

- Network-dependent tests may log errors in offline environments (follows existing test patterns)
- Some version sources (e.g., nodejs_dist) don't support resolving specific versions, returning an appropriate error

## Future Improvements

This method is foundational for the following dependent issues:
- #477: getOrGeneratePlan orchestration will use this method for Phase 1
- Cache key computation will be based on the resolved version from this method
