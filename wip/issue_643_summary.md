# Issue 643 Summary

## What Was Implemented

Added platform-conditional dependency support to the action dependency system, allowing actions to declare dependencies that only apply to specific operating systems. This enables patchelf to be installed only on Linux where it's needed for ELF RPATH fixup, while macOS users avoid unnecessary installations (since macOS uses the system-provided install_name_tool).

## Changes Made

- `internal/actions/action.go`: Added four new fields to ActionDeps struct (LinuxInstallTime, DarwinInstallTime, LinuxRuntime, DarwinRuntime) with documentation
- `internal/actions/resolver.go`: Added ResolveDependenciesForPlatform function for testable platform-specific resolution, plus getPlatformInstallDeps and getPlatformRuntimeDeps helpers
- `internal/actions/resolver_test.go`: Added comprehensive tests for platform-specific dependency resolution (11 new test cases)
- `internal/actions/dependencies_test.go`: Updated meson_build test and added TestActionDependencies_PlatformSpecific
- `internal/actions/homebrew_relocate.go`: Changed patchelf from InstallTime to LinuxInstallTime
- `internal/actions/homebrew.go`: Changed patchelf from InstallTime to LinuxInstallTime
- `internal/actions/meson_build.go`: Separated cross-platform deps from Linux-only patchelf

## Key Decisions

- **Chose simple field-based approach over map-based**: The issue proposed two options. We chose dedicated OS-specific fields (LinuxInstallTime, DarwinInstallTime) over a more complex PlatformDeps map because:
  - All current use cases only need OS-level distinction, not architecture-specific
  - Simpler implementation, easier to understand and maintain
  - Type-safe at compile time
  - Follows Go's preference for explicit, simple code

- **Added ResolveDependenciesForPlatform for testability**: Created a separate function that takes targetOS as a parameter, allowing unit tests to verify behavior for all platforms without mocking runtime.GOOS

## Trade-offs Accepted

- **Limited to Linux/Darwin platforms**: Other platforms (Windows, FreeBSD) get only cross-platform deps. This is acceptable because tsuku primarily targets Linux and macOS.
- **No architecture-specific deps yet**: Can be added later (e.g., LinuxAmd64InstallTime) if the need arises.

## Test Coverage

- New tests added: 13 test functions (11 for platform resolution, 2 for helper functions)
- All existing tests continue to pass
- Tests verify both Linux and Darwin scenarios, plus unknown OS handling

## Known Limitations

- Platform-specific fields are only for Linux and Darwin. Other platforms will only get cross-platform dependencies.
- Step-level "dependencies" override replaces all deps including platform-specific (this is intentional for consistency with existing replace behavior).

## Future Improvements

- Add architecture-specific dependency fields if needed (e.g., LinuxAmd64InstallTime, DarwinArm64InstallTime)
- Consider migrating to map-based PlatformDeps if use cases become more complex
- Potential for recipe-level platform conditionals (currently only action-level)
