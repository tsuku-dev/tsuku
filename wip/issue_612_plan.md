# Issue 612 Implementation Plan

## Summary

The go_install action decomposition is already fully implemented in the codebase. The implementation includes the Decompose method that generates go.sum at eval time and produces a go_build step for deterministic execution. All helper functions and action registrations are in place.

## Approach

After thorough code review, I discovered that the implementation is complete:
- `go_install.go` has the `Decompose` method (lines 264-402)
- `go_build.go` exists as the primitive action
- Helper functions (`ResolveGo`, `GetGoVersion`, `ResolveGoVersion`) are implemented in `util.go`
- Both actions are registered in `action.go`
- Comprehensive tests exist for both actions

The failing test mentioned in the baseline (`TestSandboxIntegration`) is expecting a `--plan` flag on the install command, which is part of Milestone 3 of the design doc (DESIGN-deterministic-resolution.md) and is separate from making go_install evaluable.

### Alternatives Considered

- **Implement missing functionality**: Initially thought decomposition was missing
- **Wait for --plan flag**: This would block unrelated to the actual decomposition
- **Just verify and test**: Chosen approach - verify implementation works and add any missing tests

## Files to Modify

### No files need modification for the core requirement

The decomposition implementation is complete. However, to fully validate:

- `internal/actions/go_install_test.go` - May add sandbox-specific test if needed
- `docs/deterministic-builds/ecosystem_go.md` - Already comprehensive

## Files to Create

None - all required files exist

## Implementation Steps

- [x] Research existing go_install and go_build implementations
- [ ] Verify Decompose method matches patterns from cargo_install, npm_install, pip_install, gem_install
- [ ] Run go_install decomposition tests to ensure they pass
- [ ] Create a simple integration test demonstrating eval → exec flow
- [ ] Update documentation if needed

## Testing Strategy

### Unit Tests
- ✅ Existing: `TestGoInstallAction_Decompose_*` tests in `go_install_test.go`
- ✅ Existing: `TestGoBuildAction_Execute_*` tests in `go_build_test.go`
- Add: Test that verifies go.sum content is captured correctly
- Add: Test that verifies Go version is captured

### Integration Tests
- Create a test that demonstrates full eval → exec cycle
- Use a real Go module (e.g., golang.org/x/tools/cmd/goimports)
- Verify:
  1. Decompose captures go.sum
  2. Generated go_build step has all required params
  3. Build succeeds with captured checksums

### Manual Verification
- Install a Go tool using tsuku to verify end-to-end flow
- Check that decomposition works for tools with dependencies

## Risks and Mitigations

### Risk 1: Missing tests for edge cases
**Mitigation**: Add tests for:
- Large dependency trees
- Go modules with replace directives
- Different Go versions

### Risk 2: Integration with eval command
**Mitigation**: The `--plan` flag is tracked separately in the design doc. This issue only covers making go_install evaluable (Decompose method), which is complete.

## Success Criteria

- [x] go_install implements Decomposable interface
- [x] Decompose() captures go.sum content at eval time
- [x] Dependency list extracted (via `go list -m -json all` in the implementation)
- [x] Exec phase uses isolated GOMODCACHE with CGO_ENABLED=0
- [x] Build flags include `-trimpath -buildvcs=false`
- [x] Go toolchain version captured in plan
- [ ] Pure Go packages marked as deterministic (verify this is happening)
- [ ] Test with a real Go tool (e.g., gofumpt if T53 exists, or another Go tool)

## Open Questions

1. Is there a T53 (gofumpt) recipe that should pass sandbox tests?
   - **Action**: Search for existing gofumpt recipe or create test case

2. Should we add more comprehensive integration tests?
   - **Action**: Add at least one end-to-end test

3. Are there any missing edge cases in the implementation?
   - **Action**: Review against ecosystem_go.md recommendations
