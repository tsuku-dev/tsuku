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

- [x] Create `set_rpath.go` with SetRpathAction struct and Name() method
- [x] Implement binary format detection (ELF vs Mach-O)
- [x] Implement Linux RPATH modification using patchelf
- [x] Implement macOS RPATH modification using install_name_tool
- [x] Implement macOS re-signing with codesign (ad-hoc)
- [x] Implement wrapper script fallback for when RPATH modification fails
- [x] Register action in action.go
- [x] Add unit tests for parameter validation
- [x] Add unit tests for binary format detection
- [x] Add platform-specific tests (wrapper fallback tests)

## Testing Strategy

- **Unit tests**: Test parameter validation, binary format detection, command construction
- **Integration tests**: Not feasible without actual binaries; rely on platform-specific CI runs
- **Manual verification**: Test with actual ruby binary from Ruby recipe

## Risks and Mitigations

- **patchelf not installed**: Mitigated by clear error message pointing user to install patchelf
- **Signed binaries on macOS**: Mitigated by wrapper script fallback and re-signing with ad-hoc signature
- **Cross-platform testing**: Mitigated by mocking command execution in tests; CI runs on both Linux and macOS

## Success Criteria

- [x] Action modifies RPATH on Linux using patchelf
- [x] Action modifies RPATH on macOS using install_name_tool (with re-signing)
- [x] Existing RPATH stripped before setting new value
- [x] Uses `$ORIGIN/../lib` pattern (not bare `$ORIGIN`)
- [x] Wrapper script fallback when RPATH modification fails
- [x] Unit tests for both platforms

## Open Questions

None
