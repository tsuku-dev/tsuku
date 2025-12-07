# Issue 219 Implementation Plan

## Summary

Implement the `install_libraries` action that copies library files matching glob patterns to the shared libs directory while preserving symlink structure.

## Approach

Create a new action following existing action patterns (like `install_binaries`). The action will:
1. Accept glob patterns for library files (e.g., `lib/*.so*`, `lib/*.dylib`)
2. Copy matching files from WorkDir to the libs destination
3. Preserve symlinks as symlinks (not dereferenced) using existing `CopySymlink` utility

### Alternatives Considered
- **Inline in homebrew_bottle**: Rejected - design doc specifies separate action for composability
- **Use CopyDirectory for entire lib/**: Rejected - need pattern matching, not full directory copy

## Files to Create
- `internal/actions/install_libraries.go` - Action implementation
- `internal/actions/install_libraries_test.go` - Unit tests

## Files to Modify
- `internal/actions/action.go` - Register the new action

## Implementation Steps
- [ ] Create InstallLibrariesAction struct with Name() method
- [ ] Implement Execute() with pattern parsing and glob matching
- [ ] Copy matched files preserving symlinks
- [ ] Register action in action.go
- [ ] Add unit tests for basic file copying
- [ ] Add unit tests for symlink preservation
- [ ] Add unit tests for glob pattern matching

## Testing Strategy
- Unit tests:
  - Copy regular files matching patterns
  - Preserve symlink structure (e.g., libyaml.so.2 -> libyaml.so.2.0.9)
  - Handle multiple patterns (both .so* and .dylib)
  - Reject path traversal attempts
  - Handle missing source files gracefully

## Risks and Mitigations
- **Path traversal attacks**: Validate patterns don't contain `..`
- **Symlink resolution**: Use `os.Lstat` not `os.Stat` to detect symlinks

## Success Criteria
- [ ] Action copies files matching glob patterns to destination
- [ ] Symlinks copied as symlinks (not dereferenced)
- [ ] Supports Linux (.so) and macOS (.dylib) patterns
- [ ] Unit tests for symlink preservation pass
- [ ] All existing tests continue to pass
