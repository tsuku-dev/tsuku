# Issue 339 Summary

## What Was Implemented

Extended the `generate-registry.py` script to extract dependency fields from recipe TOML metadata and include them in the generated `recipes.json` file. The schema version was bumped to 1.1.0 to indicate this backwards-compatible change.

## Changes Made

- `scripts/generate-registry.py`:
  - Bumped `SCHEMA_VERSION` from "1.0.0" to "1.1.0"
  - Added validation for `dependencies` and `runtime_dependencies` arrays in `validate_metadata()`
  - Added cross-recipe validation to ensure referenced dependencies exist
  - Extended `parse_recipe()` to extract dependency fields (defaulting to empty arrays)
- `internal/recipe/recipes/l/libyaml.toml`:
  - Added missing `homepage` field (pre-existing issue)

## Key Decisions

- **Validate dependency names against NAME_PATTERN**: Ensures consistency with recipe names and prevents injection attacks
- **Cross-recipe validation**: Each dependency must reference an existing recipe to prevent broken links in the UI
- **Empty arrays as default**: Recipes without dependencies get empty arrays (not omitted fields) for consistent schema

## Trade-offs Accepted

- **No separate unit test file**: Python script validation was tested manually since the script is simple and self-contained

## Test Coverage

- Manual verification performed:
  - Recipes with no dependencies: empty arrays present
  - Recipes with `dependencies` only: correctly extracted
  - Recipes with `runtime_dependencies` only: correctly extracted
  - Invalid dependency name (uppercase): validation error
  - Non-existent dependency: validation error

## Known Limitations

- None

## Future Improvements

- Could add Python unit tests if the script grows more complex
