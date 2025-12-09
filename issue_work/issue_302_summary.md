# Issue 302 Summary

## What Was Implemented

Container runtime abstraction and detection for the `internal/validate` package. This provides the foundation for container-based recipe validation by detecting available container runtimes (Podman/Docker) and providing a unified interface for running containers.

## Changes Made

- `internal/validate/runtime.go`: Core package with Runtime interface, RuntimeDetector, and implementations for Podman and Docker runtimes
- `internal/validate/runtime_test.go`: 11 unit tests covering detection scenarios with mocked exec calls
- `go.mod`: Updated Go version from 1.24.11 to 1.25.5 to fix crypto/x509 vulnerability

## Key Decisions

1. **Detection caching**: RuntimeDetector caches the detected runtime to avoid repeated detection overhead. Use Reset() to force re-detection.

2. **Preference order**: Podman rootless > Docker rootless > Docker group. This prioritizes more secure options.

3. **Hybrid detection**: Check binary availability first (fast), then verify functionality (accurate). This avoids false positives from stale binaries.

4. **Mockable design**: lookPath and cmdRun are injected dependencies, enabling thorough unit testing without actual containers.

## Trade-offs Accepted

- **No actual container test in detection**: We verify runtime info commands work rather than running a test container. This is faster but could miss some edge cases.

- **Docker group detection is permissive**: If `docker info` succeeds, we assume docker group access. This could include sudo-based access.

## Test Coverage

- New tests added: 11
- Coverage: New package, comprehensive unit test coverage for detection logic

## Known Limitations

1. Detection relies on `podman info` and `docker info` output format - could break with major version changes
2. No integration tests with actual containers (would require container runtime in CI)
3. `hasSubuidEntry()` helper is defined but not currently used in detection (prepared for future use)

## Future Improvements

- Issues #305 and #306 will add the full Podman and Docker runtime implementations with actual container execution
- Issue #308 will integrate this with the container executor
