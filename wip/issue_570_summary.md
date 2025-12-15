# Issue 570 Summary

## What Was Implemented

Implemented `SandboxRequirements` struct and `ComputeSandboxRequirements()` function in a new `internal/sandbox/` package. The function derives container configuration (image, network access, resource limits) from installation plans by querying each action's `RequiresNetwork()` method via the `NetworkValidator` interface.

## Changes Made
- `internal/sandbox/requirements.go`: New file with:
  - `SandboxRequirements` struct (RequiresNetwork, Image, Resources)
  - `ComputeSandboxRequirements()` function that iterates plan steps
  - `ResourceLimits` struct for container resource configuration
  - `DefaultLimits()` and `SourceBuildLimits()` helper functions
  - `hasBuildActions()` helper to detect build-intensive plans
  - Constants for container images (DefaultSandboxImage, SourceBuildSandboxImage)
- `internal/sandbox/requirements_test.go`: Comprehensive unit tests

## Key Decisions
- **Separate package**: Created `internal/sandbox/` to keep sandbox testing logic isolated from the existing `internal/validate/` package. This allows future unification work (#571) to cleanly refactor.
- **Fail-closed for unknown actions**: Unknown actions default to no network access for security.
- **Build actions upgrade resources**: Even if configure_make doesn't need network (source is cached), it still gets upgraded resources since compilation is resource-intensive.

## Trade-offs Accepted
- **Redundant ResourceLimits type**: Created new ResourceLimits type in sandbox package rather than reusing validate.ResourceLimits. This avoids a dependency but creates some duplication. Will be unified in #571.

## Test Coverage
- New tests added: 12 test functions with multiple subtests
- Tests cover: nil plan, empty plan, offline plan, network-required actions (8 different actions), build actions, unknown actions, mixed plans

## Known Limitations
- ResourceLimits duplicates the existing type in validate package (intentional - will unify in #571)
- Does not yet integrate with the executor - that's the scope of #571

## Future Improvements
- #571 will unify Sandbox() and SandboxSourceBuild() methods using this requirements computation
- #572 will update builders to use centralized sandbox testing
- #573 will add --sandbox CLI flag
