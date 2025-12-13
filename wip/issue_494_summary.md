# Issue 494 Summary

## What Was Implemented

Container-based validation for source-built Homebrew recipes. When a formula doesn't have pre-built bottles, tsuku now validates the generated recipe by running it in an isolated container.

## Key Changes

### New Files

- `internal/validate/source_build.go`: Source build validation logic
  - `SourceBuildValidationImage`: Ubuntu 22.04 (for build tool support)
  - `SourceBuildLimits()`: 4GB RAM, 4 CPUs, 500 PIDs, 15-minute timeout
  - `ValidateSourceBuild()`: Main validation entry point
  - `buildSourceBuildScript()`: Generates shell script that installs build tools and runs tsuku install
  - `detectRequiredBuildTools()`: Analyzes recipe actions to determine apt packages

- `internal/validate/source_build_test.go`: Comprehensive unit tests (15 test cases)

### Modified Files

- `internal/builders/homebrew.go`: Updated `buildFromSource()` to call `ValidateSourceBuild()` when executor is available

## How It Works

1. When `buildFromSource()` generates a recipe for a bottle-less formula, it now:
2. Creates a workspace with the recipe TOML
3. Generates a validation script that:
   - Updates apt and installs base requirements
   - Installs build tools based on detected actions (autotools, cmake, etc.)
   - Sets up `$TSUKU_HOME` directory structure
   - Runs `tsuku install <package> --force`
4. Runs the script in an Ubuntu 22.04 container
5. Reports validation results to telemetry
6. Returns the recipe (doesn't fail build on validation failure - informational only)

## Build Tool Detection

The validation dynamically installs apt packages based on recipe actions:

| Action | Packages Installed |
|--------|-------------------|
| `configure_make` | autoconf, automake, libtool, pkg-config |
| `cmake_build` | cmake, ninja-build |
| `cargo_build`/`cargo_install` | curl (for rustup) |
| `go_build`/`go_install` | curl (for Go download) |
| `apply_patch` | patch |
| `cpan_install` | perl, cpanminus |

Darwin-only steps (via `when.os = "darwin"`) are skipped during Linux container validation.

## Testing

- 15 new unit tests covering:
  - Limits and constants
  - Build tool detection for all supported build systems
  - Platform conditional handling (darwin vs linux)
  - Script generation structure and content
  - Validation with no runtime (skipped)
  - Successful validation
  - Failed validation (container error)

## Future Work

- LLM repair loop for source build validation failures (similar to bottle validation)
- Integration tests with live LLM for end-to-end source build generation + validation
