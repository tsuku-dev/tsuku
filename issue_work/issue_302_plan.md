# Issue 302 Implementation Plan

## Summary
Create the container runtime abstraction and detection system for `internal/validate/runtime.go`, defining the `Runtime` interface and `RuntimeDetector` that identifies available container runtimes (Podman/Docker) in preference order.

## Approach
Implement a clean interface-based design where:
1. `Runtime` interface defines the contract for container runtimes
2. `RuntimeDetector` handles detection with caching to avoid repeated checks
3. Detection uses hybrid approach: check binary availability first, then verify functionality with a simple container test

### Alternatives Considered
- **Pure configuration check**: Only check `/etc/subuid` and binary presence - rejected because it can give false positives; a binary might exist but not work due to permission issues
- **Always run test container**: Run a test container on every detection - rejected because it's slow (~5-10s) and requires network for image pull

## Files to Create
- `internal/validate/runtime.go` - Core interface and detector
- `internal/validate/runtime_test.go` - Unit tests with mocked runtimes

## Implementation Steps
- [x] Create `internal/validate/` package with `runtime.go`
- [x] Define `Runtime` interface with `Name()`, `IsRootless()`, `Run()` methods
- [x] Define `RunOptions` and `RunResult` types for container execution
- [x] Implement `RuntimeDetector` struct with detection logic
- [x] Implement Podman detection (binary check + rootless verification)
- [x] Implement Docker detection (rootless mode vs group membership)
- [x] Add `ErrNoRuntime` sentinel error
- [x] Write unit tests for detection logic
- [x] Verify tests pass

## Testing Strategy
- Unit tests: Mock exec.Command calls to simulate podman/docker availability
- Test cases:
  - Podman available and rootless working
  - Docker rootless available
  - Docker group available (with warning flag)
  - No runtime available -> ErrNoRuntime
  - Detection caching (subsequent calls return cached result)

## Risks and Mitigations
- **Binary detection varies by system**: Use `exec.LookPath` for portable binary detection
- **Container test may fail for unrelated reasons**: Distinguish between "not available" and "available but failed"

## Success Criteria
- [ ] `Runtime` interface defined with `Name()`, `IsRootless()`, `Run()` methods
- [ ] `RuntimeDetector` detects Podman when available
- [ ] `RuntimeDetector` detects Docker (rootless vs group)
- [ ] Returns `ErrNoRuntime` when no runtime available
- [ ] Unit tests pass

## Open Questions
None - design document provides clear guidance.
