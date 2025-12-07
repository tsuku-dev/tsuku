# Issue 217 Implementation Plan

## Summary

Implement the `set_rpath` action for modifying binary RPATH to enable relocatable library loading, using patchelf on Linux and install_name_tool on macOS.

## Approach

Create a new action that detects the binary format (ELF/Mach-O) and uses the appropriate platform tool to modify RPATH. The action will strip existing RPATH values before setting new ones (security requirement) and use the `$ORIGIN/../lib` pattern. A wrapper script fallback will handle cases where RPATH modification fails (e.g., signed binaries on macOS).

### Alternatives Considered

- **Shell script wrapper only**: Rejected - RPATH provides better security (can't be overridden by environment) and performance
- **Single cross-platform binary patching library**: Rejected - complexity of implementing ELF/Mach-O parsing; patchelf and install_name_tool are well-tested

## Files to Modify

- `internal/actions/action.go` - Register the new SetRpathAction

## Files to Create

- `internal/actions/set_rpath.go` - Main action implementation
- `internal/actions/set_rpath_test.go` - Unit tests

## Implementation Steps

- [ ] Create `set_rpath.go` with SetRpathAction struct and Name() method
- [ ] Implement binary format detection (ELF vs Mach-O)
- [ ] Implement Linux RPATH modification using patchelf
- [ ] Implement macOS RPATH modification using install_name_tool
- [ ] Implement macOS re-signing with codesign (ad-hoc)
- [ ] Implement wrapper script fallback for when RPATH modification fails
- [ ] Register action in action.go
- [ ] Add unit tests for parameter validation
- [ ] Add unit tests for binary format detection
- [ ] Add platform-specific tests (mocked command execution)

## Testing Strategy

- **Unit tests**: Test parameter validation, binary format detection, command construction
- **Integration tests**: Not feasible without actual binaries; rely on platform-specific CI runs
- **Manual verification**: Test with actual ruby binary from Ruby recipe

## Risks and Mitigations

- **patchelf not installed**: Mitigated by clear error message pointing user to install patchelf
- **Signed binaries on macOS**: Mitigated by wrapper script fallback and re-signing with ad-hoc signature
- **Cross-platform testing**: Mitigated by mocking command execution in tests; CI runs on both Linux and macOS

## Success Criteria

- [ ] Action modifies RPATH on Linux using patchelf
- [ ] Action modifies RPATH on macOS using install_name_tool (with re-signing)
- [ ] Existing RPATH stripped before setting new value
- [ ] Uses `$ORIGIN/../lib` pattern (not bare `$ORIGIN`)
- [ ] Wrapper script fallback when RPATH modification fails
- [ ] Unit tests for both platforms

## Open Questions

None
