# Issue 296 Implementation Plan

## Summary

Implement atomic symlink updates to prevent race conditions during version switching. The atomic pattern is: create temp symlink, then rename over existing.

## Approach

Create a new `AtomicSymlink` function in the install package that:
1. Creates a temporary symlink in the same directory as the target
2. Uses `os.Rename` to atomically replace the existing symlink
3. Validates that the target is within `$TSUKU_HOME/tools/`

### Why Atomic Updates Matter

The current implementation in `createBinarySymlink` does:
```go
os.Remove(symlinkPath)  // Window of time with no symlink
os.Symlink(targetPath, symlinkPath)
```

If a crash occurs between Remove and Symlink, the user has a broken PATH. The atomic pattern:
```go
os.Symlink(targetPath, tmpPath)  // Create temp symlink
os.Rename(tmpPath, symlinkPath)  // Atomic replacement
```

## Files to Modify

- `internal/install/manager.go` - Update `createBinarySymlink` to use atomic updates

## Files to Create

- `internal/install/symlink.go` - New file with `AtomicSymlink` function and validation
- `internal/install/symlink_test.go` - Unit tests for atomic symlink creation

## Implementation Steps

- [ ] Create `internal/install/symlink.go` with:
  - `AtomicSymlink(target, linkPath string) error` - creates symlink atomically
  - `ValidateSymlinkTarget(target, tsukuHome string) error` - validates target is within `$TSUKU_HOME/tools/`
- [ ] Add unit tests in `internal/install/symlink_test.go`
- [ ] Update `createBinarySymlink` in `manager.go` to use `AtomicSymlink`
- [ ] Verify existing tests still pass

## Testing Strategy

- Unit tests for `AtomicSymlink`:
  - Creates symlink when target doesn't exist
  - Atomically replaces existing symlink
  - Atomically replaces existing regular file (edge case)
  - Works with relative and absolute targets
- Unit tests for `ValidateSymlinkTarget`:
  - Allows targets within `$TSUKU_HOME/tools/`
  - Rejects targets outside `$TSUKU_HOME/tools/`
  - Rejects path traversal attempts

## Success Criteria

- [ ] `AtomicSymlink` function implemented
- [ ] Target validation prevents path traversal
- [ ] Unit tests pass
- [ ] Existing manager tests pass
- [ ] Build passes
