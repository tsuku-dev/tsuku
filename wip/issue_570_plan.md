# Issue 570 Implementation Plan

## Summary

Implement `SandboxRequirements` struct and `ComputeSandboxRequirements()` function in a new `internal/sandbox/requirements.go` file. The function derives container configuration from installation plans by querying actions' `RequiresNetwork()` method.

## Approach

Follow the design in `docs/DESIGN-install-sandbox.md`. The implementation queries the already-implemented `NetworkValidator` interface (#568, #569) on each action to aggregate network requirements, then selects appropriate container image and resource limits.

### Alternatives Considered
- Embed requirements in plan struct: Rejected - would require plan format version bump and lose ability to work with existing plans
- Compute at plan generation time: Rejected - requirements should be computable from any plan, including hand-written ones

## Files to Create
- `internal/sandbox/requirements.go` - SandboxRequirements struct and computation function
- `internal/sandbox/requirements_test.go` - Unit tests for requirements computation

## Implementation Steps
- [ ] Create `internal/sandbox/` package directory
- [ ] Implement `SandboxRequirements` struct with fields: RequiresNetwork, Image, Resources
- [ ] Implement `ComputeSandboxRequirements()` function that iterates plan steps and queries NetworkValidator
- [ ] Implement `hasBuildActions()` helper for detecting build-intensive plans
- [ ] Define constants for default images (DefaultSandboxImage, SourceBuildSandboxImage)
- [ ] Define `SourceBuildLimits()` function for build resource limits
- [ ] Write comprehensive unit tests covering offline plans, network plans, and build plans

## Testing Strategy
- Unit tests: Test computation with various plan configurations
  - Offline-only plan (download, extract, install_binaries) -> no network, debian image
  - Network-required plan (cargo_build, go_build) -> network, ubuntu image
  - Build plan without network (configure_make with offline deps) -> no network, ubuntu image, build resources
  - Empty plan -> defaults
  - Unknown actions -> safe defaults (no network)

## Risks and Mitigations
- Action not implementing NetworkValidator: Mitigated by BaseAction default (returns false)
- Future actions added without RequiresNetwork: Fail-closed design - unknown actions default to no network

## Success Criteria
- [ ] `SandboxRequirements` struct created in `internal/sandbox/requirements.go`
- [ ] `ComputeSandboxRequirements(plan)` function implemented
- [ ] Function queries actions via NetworkValidator interface
- [ ] Network requirement aggregated correctly (any true = plan needs network)
- [ ] Image selection: debian:bookworm-slim for offline, ubuntu:22.04 for network/build
- [ ] Resource limits selected based on plan complexity
- [ ] Unit tests cover offline, network, and build plan configurations
- [ ] All tests pass
- [ ] No lint warnings

## Open Questions
None - design is clear and dependencies (#568, #569) are already implemented.
