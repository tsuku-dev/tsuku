# Issue 560 Summary

## What Was Implemented
Implemented the `require_system` primitive action that validates system-installed dependencies (like Docker, CUDA) that tsuku cannot provision. The action detects command presence via secure exec.Command calls (no shell), parses versions using configurable regex, validates minimum version requirements, and provides platform-specific installation guidance when dependencies are missing or outdated.

## Changes Made
- `internal/actions/require_system.go`: New primitive action with command detection, version parsing, and error handling
- `internal/actions/require_system_test.go`: Comprehensive unit tests covering command detection, version validation, error handling, and edge cases
- `internal/actions/action.go`: Registered RequireSystemAction in the action registry init() function
- `internal/version/version_utils.go`: Exported CompareVersions() function for use by require_system and other actions
- `internal/version/npm.go`, `resolver.go`, `resolver_test.go`: Updated to use exported CompareVersions()

## Key Decisions
- **No shell execution**: Used exec.LookPath() and exec.Command() directly instead of shell commands for security (prevents command injection)
- **Regex-based version parsing**: Made version_regex configurable per recipe to support different tool output formats
- **Platform-specific guidance via maps**: Used install_guide map with platform keys (darwin, linux, fallback) for maintainability
- **Reused existing version comparison**: Exported CompareVersions() from version package instead of reimplementing
- **Hierarchical validation**: Implemented command exists → version check → min version validation flow as specified in design

## Trade-offs Accepted
- **No runtime validation in initial implementation**: Deferred runtime checks (e.g., `docker info`) to future enhancement to keep this PR focused
- **Simple platform detection**: Used runtime.GOOS directly instead of distro-specific detection (linux.ubuntu, linux.fedora); this can be added later if needed
- **No assisted installation**: Skipped the assisted install feature (user consent → execute install command) for this PR; marked for future enhancement

## Test Coverage
- New tests added: 13 test cases covering:
  - Action name and determinism
  - Command name validation (security)
  - Version detection and parsing
  - Version comparison logic
  - Platform guide selection
  - Error handling (missing command, version mismatch, invalid params)
  - Edge cases (invalid regex, missing guides, empty commands)
- All tests pass: 100% of new code paths tested
- No coverage regression: All existing tests still pass

## Known Limitations
- Platform detection is basic (darwin/linux/windows) without distro-specific matching
- No validation of install_guide URLs (future: enforce HTTPS-only via validation)
- Version comparison uses simple semver logic (may not handle all pre-release formats)
- No caching of command detection results (executes on every install)

## Future Improvements
- Add runtime validation parameter for tools requiring daemon/service checks (Issue #643)
- Implement hierarchical platform matching (linux.ubuntu, linux.fedora, etc.)
- Add assisted installation feature with user consent prompts
- Cache detection results with TTL to avoid repeated checks
- Validate install_guide URLs are HTTPS-only in recipe validation
