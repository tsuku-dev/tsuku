# Issue 235 Summary

## What Was Implemented

Dependency resolution algorithm that collects install-time and runtime dependencies from recipes by examining each step's action type and merging with step-level extensions.

## Changes Made

- `internal/actions/resolver.go`: New file with:
  - `ResolvedDeps` struct holding install-time and runtime deps as maps
  - `ResolveDependencies(recipe)` function that walks recipe steps
  - `parseDependency(dep)` helper for parsing "name@version" syntax
  - `getStringSliceParam()` helper for extracting string slices from step params

- `internal/actions/resolver_test.go`: Comprehensive tests covering:
  - npm_install → nodejs in both install and runtime
  - go_install → go in install only
  - download → empty deps
  - Multiple steps merging correctly
  - extra_dependencies and extra_runtime_dependencies handling

## Key Decisions

- **Placed in actions package**: Avoids import cycle with recipe package; resolver needs ActionDependencies from actions
- **Map[string]string for deps**: Keys are dep names, values are versions; supports deduplication naturally
- **"latest" as default version**: When no version specified, use "latest" as the version value

## Trade-offs Accepted

- **No recipe-level overrides yet**: Issue #236 will add recipe-level dependencies and runtime_dependencies fields
- **No transitive resolution**: Issue #237 will add recursive dependency resolution

## Test Coverage

- New tests added: 10 tests in resolver_test.go
- All 17 packages pass

## Known Limitations

- Only handles step-level extensions; recipe-level fields handled in #236
- No transitive resolution; handled in #237
- No version constraint solving; dependencies use simple "name@version" or "latest"

## Future Improvements

- Issue #236: Recipe-level dependency overrides
- Issue #237: Transitive resolution with cycle detection
