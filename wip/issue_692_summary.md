# Issue 692 Summary

## What Was Implemented

Added `--recipe` flag to the `tsuku eval` command, allowing users to evaluate local recipe files without adding them to the registry first. This matches the UX pattern established by `tsuku install --recipe`.

## Changes Made

- `internal/recipe/loader.go`: Added `ParseFile` function for loading and validating recipes from file paths
- `internal/recipe/loader_test.go`: Added unit tests for `ParseFile` covering valid recipes, missing files, invalid TOML, and missing required fields
- `cmd/tsuku/eval.go`: Added `--recipe` flag and updated `runEval` to handle both registry and file-based recipe loading
- `cmd/tsuku/install_sandbox.go`: Updated `loadLocalRecipe` to use shared `recipe.ParseFile` function

## Key Decisions

- **Flag-based approach over positional detection**: Changed from detecting file paths in the positional argument (per original issue suggestion) to using `--recipe` flag, matching the `install` command's UX for consistency
- **Shared recipe loading**: Both `eval` and `install` commands now use the same `recipe.ParseFile` function, reducing code duplication and ensuring consistent behavior

## Trade-offs Accepted

- **Version specification not supported with --recipe**: The `@version` syntax doesn't work with file paths since the recipe file defines version configuration. This is consistent with `install --recipe`.

## Test Coverage

- New tests added: 4 (for `ParseFile` function)
- Unit tests verify: valid recipe parsing, file not found, invalid TOML, missing required fields

## Known Limitations

- `--recipe` and tool name are mutually exclusive (clear error message provided)
- No `@version` support with `--recipe` (recipe file defines version behavior)

## Future Improvements

- Could add validation warnings when using `--recipe` to help catch common recipe issues during development
