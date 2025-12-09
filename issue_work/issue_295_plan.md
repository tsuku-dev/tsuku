# Issue 295 Implementation Plan

## Summary

Add file locking to StateManager to prevent state.json corruption from concurrent tsuku processes, using flock(2) for cross-process locking and atomic writes for crash safety.

## Approach

Use Go's `syscall.Flock` for advisory file locking on a dedicated lock file (`state.json.lock`), combined with atomic write pattern (write to temp, rename). The lock protects the entire read-modify-write cycle.

### Alternatives Considered
- **In-memory mutex only**: Already exists (`sync.RWMutex`), but doesn't protect against concurrent processes
- **Database (SQLite/BoltDB)**: Overkill for simple JSON state, adds dependencies
- **Lock-free optimistic concurrency**: Complex conflict resolution, not worth it for this use case

## Files to Modify
- `internal/install/state.go` - Add file locking and atomic write to StateManager

## Files to Create
- `internal/install/filelock.go` - File locking abstraction (platform-specific)
- `internal/install/filelock_unix.go` - Unix flock implementation
- `internal/install/filelock_windows.go` - Windows LockFileEx implementation

## Implementation Steps
- [x] Add FileLock type with Lock/Unlock methods in `filelock.go`
- [x] Implement Unix flock in `filelock_unix.go` (build tag: `//go:build !windows`)
- [x] Implement Windows LockFileEx in `filelock_windows.go` (build tag: `//go:build windows`)
- [x] Add `lockPath()` method to StateManager returning `state.json.lock`
- [x] Modify `Load()` to acquire shared lock, release after read
- [x] Modify `Save()` to acquire exclusive lock, write atomically, release lock
- [x] Modify read-modify-write methods (UpdateTool, etc.) to hold exclusive lock for entire operation
- [x] Add unit tests for concurrent access
- [x] Add unit tests for atomic write (crash safety)

Mark each step [x] after it is implemented and committed. This enables clear resume detection.

## Testing Strategy
- Unit tests: Concurrent goroutines modifying state (verify no corruption)
- Unit tests: Simulate crash mid-write (verify atomic recovery)
- Manual verification: Run two tsuku install commands simultaneously

## Risks and Mitigations
- **Windows compatibility**: Use build tags with Windows-specific LockFileEx implementation
- **Stale locks**: Use advisory locking (automatically released on process exit)
- **Deadlock**: Keep locks short-lived, no nested locking

## Success Criteria
- [x] File locking implemented with flock(2) on Unix
- [x] Atomic write pattern (temp file + rename)
- [x] Concurrent access tests pass
- [x] Build passes on Linux (Windows support as best-effort)
- [x] No deadlocks in normal operation
