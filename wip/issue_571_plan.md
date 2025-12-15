# Issue 571 Implementation Plan

## Summary

Unify `Validate()` and `ValidateSourceBuild()` into a single `Sandbox()` method that takes `SandboxRequirements`. The new method uses requirements-driven configuration instead of hardcoded decisions per builder type.

## Approach

Move sandbox execution logic from `internal/validate/` to `internal/sandbox/` package. The unified method:
1. Takes a plan and computed requirements (from #570)
2. Configures container based on requirements (image, network, resources)
3. Generates a simplified script (no more `detectRequiredBuildTools`)
4. Runs container and checks verification

The `internal/validate/` package will be deprecated in favor of `internal/sandbox/`.

### Key Design Decision

Per the design doc, build tools are NOT installed via apt-get in the sandbox script. Instead, tsuku's normal dependency resolution handles them via `ActionDependencies.InstallTime`. The sandbox script simply runs `tsuku install --plan`.

## Files to Modify/Create

### Create
- `internal/sandbox/executor.go` - Unified Sandbox executor with Sandbox() method
- `internal/sandbox/executor_test.go` - Tests for the executor

### Modify
- `internal/sandbox/requirements.go` - May need minor adjustments for integration

### Keep for backward compatibility (no changes for now)
- `internal/validate/*.go` - Will be deprecated in #572 when builders migrate

## Implementation Steps
- [ ] Create `internal/sandbox/executor.go` with Executor struct and Sandbox() method
- [ ] Copy and adapt runtime detection, predownloader types from validate package
- [ ] Implement buildSandboxScript() that uses SandboxRequirements for network/image selection
- [ ] Implement checkVerification() method (similar to validate package)
- [ ] Add SandboxResult struct (similar to ValidationResult)
- [ ] Write unit tests for the new executor
- [ ] Verify all existing tests still pass

## Testing Strategy
- Unit tests for Sandbox() with mocked runtime
- Test with offline requirements (no network, debian image)
- Test with network requirements (host network, ubuntu image)
- Test verification pattern matching

## Risks and Mitigations
- Duplicated code between validate and sandbox packages: Acceptable during transition, will be cleaned in #572
- Breaking existing builders: Not a concern since this issue just adds the new method, builders migrate in #572

## Success Criteria
- [ ] `Executor.Sandbox()` method accepts `*executor.InstallationPlan` and `*SandboxRequirements`
- [ ] Method configures container based on requirements (image, network, resources)
- [ ] Sandbox script uses simplified approach (no detectRequiredBuildTools)
- [ ] All existing tests pass
- [ ] New tests cover different requirement combinations
