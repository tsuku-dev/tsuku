# Issue 521 Summary

## What Was Implemented
Implemented `meson_build` action as an ecosystem primitive, following the established patterns from `cmake_build` and `configure_make`. The action supports the standard Meson build workflow (setup → compile → install) with proper parameter validation, security checks, and executable verification.

## Changes Made
- `internal/actions/meson_build.go`: Created MesonBuildAction implementing three-phase Meson build workflow
- `internal/actions/meson_build_test.go`: Added comprehensive unit tests covering all requirements
- `internal/actions/action.go`: Registered MesonBuildAction in init function
- `internal/actions/decomposable.go`: Added "meson_build" to primitives map
- `internal/actions/decomposable_test.go`: Updated TestPrimitives to include new primitive

## Key Decisions

### Decision 1: Follow cmake_build pattern exactly
**Rationale**: Meson and CMake are both modern build systems with similar workflows. Using the same structure ensures consistency and makes the codebase easier to maintain. Key similarities:
- Both use isolated build directories
- Both have three-phase execution (setup/configure, build/compile, install)
- Both need compiler environment setup
- Both require security validation

### Decision 2: Validate buildtype against known values
**Rationale**: Unlike cmake where build type can be arbitrary, Meson has well-defined buildtype values (release, debug, plain, debugoptimized). Explicit validation prevents typos and potential issues.

### Decision 3: Use `--wrap-mode=nofallback` as default
**Rationale**: This prevents Meson from automatically downloading dependencies from the internet, which aligns with tsuku's philosophy of reproducible, deterministic builds using pre-downloaded or system-provided dependencies.

## Trade-offs Accepted

### Integration tests deferred
**Why acceptable**: The unit tests thoroughly cover parameter validation, error handling, and execution logic. Integration tests would require:
1. Meson to be available in the CI environment
2. A suitable test project (json-glib or similar)
3. Additional CI time and complexity

The action follows proven patterns from cmake_build (which has no integration tests either), so risk is minimal. Integration can be added later if needed.

## Test Coverage
- New tests added: 11 unit tests
- All tests pass (23/23 packages)
- Build verified successful
- Coverage maintained (no regressions)

Unit tests cover:
- MB-1: Basic meson project with defaults
- MB-2: Custom meson_args
- MB-3: Invalid meson_args (security validation)
- MB-4: Missing meson.build file
- MB-5: Invalid buildtype
- MB-6: Invalid executable names (path traversal prevention)
- MB-7: Relative source_dir resolution
- Plus: Registration, determinism, and validation helper tests

## Known Limitations

### Meson must be pre-installed
Like cmake_build and configure_make, this action assumes Meson is available in PATH. It does not bootstrap Meson. This is consistent with the ecosystem primitive model where the build tool itself is a system dependency.

### Limited cross-compilation support
The action accepts meson_args where `--cross-file` could be passed, but cross-compilation is not explicitly tested or documented. This can be enhanced later if needed.

## Future Improvements

### Integration test with real Meson project
Once CI has Meson pre-installed (or using Docker), add integration test with json-glib or another simple Meson project to verify end-to-end functionality.

### Recipe creation for Meson-based tools
Create recipes for popular Meson-based tools (json-glib, glib, libsoup) to demonstrate the action's real-world utility.

### Cross-compilation documentation
If users need cross-compilation, document how to use `--cross-file` parameter via meson_args.
