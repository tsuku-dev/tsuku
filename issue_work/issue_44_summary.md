# Issue 44 Summary

## What Was Implemented

Enhanced recipe management UX with source indicators, filtering, and improved error messages. The `tsuku recipes` command now shows whether recipes are from local or registry sources, supports `--local` filtering, and provides helpful suggestions when a tool is not found.

## Changes Made

- `internal/recipe/loader.go`: Added `ListAllWithSource()`, `ListLocal()`, `RecipesDir()` methods and `RecipeSource`/`RecipeInfo` types
- `internal/recipe/loader_test.go`: Added tests for new listing functionality (7 new test functions)
- `internal/registry/registry.go`: Added `ListCached()` method to enumerate cached recipes
- `internal/registry/registry_test.go`: Added tests for `ListCached()` (3 new test functions)
- `cmd/tsuku/recipes.go`: Rewrote command to show source indicators and support `--local` flag
- `cmd/tsuku/install.go`: Enhanced error message when tool not found with ecosystem suggestions

## Key Decisions

- **Local recipes listed first**: When a recipe exists in both local and registry, only the local version is shown (since local takes precedence during loading)
- **Parse recipes for descriptions**: We parse the TOML to get descriptions, which has performance implications but provides a better UX
- **Show count by source**: Header shows breakdown (e.g., "2 total: 1 local, 1 registry") for visibility

## Trade-offs Accepted

- **Parsing all recipes**: For large recipe sets this could be slow, but the expected number of recipes is small enough that this is acceptable
- **No registry-only filter**: Only `--local` flag was implemented; no `--registry` flag since the default shows both

## Test Coverage

- New tests added: 10 (7 in loader_test.go, 3 in registry_test.go)
- All existing tests continue to pass

## Known Limitations

- The recipes list only shows locally cached registry recipes, not all available remote recipes
- No pagination for very long recipe lists

## Future Improvements

- Add registry search to discover uncached recipes
- Add `--registry` filter if needed
- Consider lazy loading of descriptions for performance optimization
