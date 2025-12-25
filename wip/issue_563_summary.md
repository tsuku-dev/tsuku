# Issue 563 Summary

## What Was Implemented

Added `tsuku check-deps <recipe>` command that validates dependencies before installation. The command resolves all direct and transitive dependencies, classifies each as provisionable (managed by tsuku) or system-required (external), checks their installation status, and outputs a colorized report.

## Changes Made

- `cmd/tsuku/check_deps.go`: New command implementation (~350 lines)
  - `checkDepsCmd` cobra command with `--json` flag
  - `DepStatus` and `CheckDepsOutput` structs for status tracking
  - `runCheckDeps()` main function using existing resolver patterns
  - `checkDependency()` for type classification and status detection
  - `isSystemRequiredRecipe()` to classify recipes by step types
  - `checkSystemDependency()` for exec.LookPath + version validation
  - `checkProvisionableDependency()` for installed tool detection
  - Colorized output functions with ANSI codes

- `cmd/tsuku/check_deps_test.go`: Unit tests (~200 lines)
  - Tests for `isSystemRequiredRecipe()` classification
  - Tests for `mergeDeps()` function
  - Tests for `getInstallGuide()` extraction
  - Tests for struct creation

## Key Decisions

- **Reused existing patterns**: Leveraged `actions.ResolveDependencies()` and `actions.ResolveTransitive()` from `info.go` for consistency
- **Named function `isSystemRequiredRecipe()`**: Avoids collision with existing `isSystemDependencyRecipe()` in `plan_install.go` which takes `*executor.InstallationPlan`
- **Exit code on system deps only**: Only exits non-zero when system-required dependencies have issues; missing provisionable deps are informational since they can be auto-installed
- **ANSI codes over library**: Used inline ANSI escape codes rather than adding a color library dependency

## Trade-offs Accepted

- **No TTY detection**: Colors are always output; future enhancement could add `--no-color` flag
- **Simple version comparison**: Uses existing `version.CompareVersions()` which handles semver; complex version ranges not supported

## Test Coverage

- New tests added: 5 test functions with 16 test cases
- Tests cover core classification and utility functions
- Integration testing via manual `./tsuku check-deps <recipe>`

## Known Limitations

- Requires recipes to be in registry for type classification
- Unknown dependencies (recipe not found) show as "unknown" type with "missing" status
- Color output in non-TTY environments (can be addressed with future `--no-color` flag)

## Future Improvements

- Add `--no-color` flag for non-interactive use
- Add `--installed-only` to filter to installed dependencies
- Consider caching classification results for performance on large dependency trees
