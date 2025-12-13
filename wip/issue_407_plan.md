# Issue 407 Implementation Plan

## Summary

Harden download cache against symlink attacks and improve permissions security.

## Approach

Add security helpers to detect symlinks and validate permissions, then integrate them into cache operations. Fail operations if security checks detect issues.

### Alternatives Considered

- **Ignore symlinks silently**: Rejected - security issues should be explicit errors
- **Follow symlinks if safe**: Rejected - adds complexity, better to reject all symlinks

## Files to Modify

- `internal/actions/download_cache.go` - Add security checks and change directory permissions

## Files to Create

None - all changes in existing file

## Implementation Steps

- [ ] Add `isSymlink(path)` helper to check if a path is or contains a symlink
- [ ] Add `validateCacheDir(path)` to validate permissions (mode 0700)
- [ ] Update `Save()` to:
  - [ ] Create cache directory with mode 0700 (not 0755)
  - [ ] Check for symlinks before writing
  - [ ] Validate existing directory permissions
- [ ] Update `Check()` to validate directory before operations
- [ ] Add unit tests for symlink detection
- [ ] Add unit tests for permission validation
- [ ] Run `go vet`, `go test`, and `go build` to verify

## Testing Strategy

- Unit tests: Create symlink and verify operations are rejected
- Unit tests: Create directory with wrong permissions, verify validation fails
- Existing tests should continue to pass

## Success Criteria

- [x] `os.MkdirAll` uses mode 0700
- [ ] Symlink detection prevents cache writes to symlinked paths
- [ ] Permission validation rejects directories with wrong permissions
- [ ] Clear error messages for security check failures
- [ ] All tests pass, no lint errors

## Open Questions

None.
