# Issue 45 Implementation Plan

## Summary

Add toolchain availability checks to the `tsuku create` command before making API calls, providing clear error messages when required toolchains (cargo, gem, pip/pipx, npm) are missing.

## Approach

Add a new `internal/toolchain` package with functions to check if toolchains are available in PATH, then integrate these checks into `cmd/tsuku/create.go` before calling builder methods. Each ecosystem maps to a specific toolchain binary that must be present.

### Alternatives Considered
- **Add checks inside builders**: Rejected because checks should happen early to avoid wasted API calls, and the create command is the proper place for this validation
- **Add checks in actions**: Rejected because this issue is specifically about `tsuku create`, not `tsuku install` (actions already fail with errors when toolchains are missing)

## Files to Modify
- `cmd/tsuku/create.go` - Add toolchain check before builder.CanBuild()

## Files to Create
- `internal/toolchain/toolchain.go` - Toolchain availability checking
- `internal/toolchain/toolchain_test.go` - Tests for toolchain checking

## Implementation Steps
- [x] Create internal/toolchain package with CheckAvailable function
- [x] Add ecosystem-to-toolchain mapping
- [x] Add toolchain check to create.go before API calls
- [x] Add tests for toolchain checking
- [x] Verify all tests pass

## Testing Strategy
- Unit tests: Mock exec.LookPath behavior via interface/function injection
- Manual verification: Run `tsuku create foo --from crates.io` without cargo installed

## Risks and Mitigations
- **PATH changes during execution**: Low risk, check is done once at start
- **Multiple toolchain binaries per ecosystem**: For now, check primary binary only (cargo, gem, pipx, npm)

## Success Criteria
- [ ] `tsuku create foo --from crates.io` checks for `cargo` availability before API call
- [ ] Clear error message when toolchain missing: "Cargo is required. Install Rust or run: tsuku install rust"
- [ ] Same detection for gem (Ruby), pipx (Python), npm (Node.js)
- [ ] Check happens before API calls to avoid wasted network requests
- [ ] All existing tests continue to pass

## Open Questions
None - requirements are clear from issue and design doc
