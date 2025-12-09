# Issue 281 Summary

## What Was Implemented

Implemented the `GitHubReleaseBuilder` that uses the LLM client to analyze GitHub release assets and generate installation recipes for tsuku.

## Changes Made

- `internal/builders/github_release.go`: New file with the builder implementation
  - `GitHubReleaseBuilder` struct with HTTP and LLM clients
  - `fetchReleases()` fetches last 5 releases from GitHub API
  - `fetchRepoMeta()` fetches repository description and homepage
  - `fetchREADME()` fetches README.md from raw.githubusercontent.com
  - `generateRecipe()` transforms LLM pattern into `recipe.Recipe`
  - `deriveAssetPattern()` infers asset pattern from concrete mappings
- `internal/builders/github_release_test.go`: Comprehensive unit tests

## Key Decisions

- **Separate `CanBuild()` always returns false**: This builder requires `SourceArg`, not just package name, so it cannot auto-detect packages like other builders
- **README fetch is non-fatal**: Returns empty string on failure rather than erroring the build
- **Pattern derivation from first mapping**: Uses first mapping as template for asset pattern, replacing OS/arch with placeholders

## Trade-offs Accepted

- **Simple asset pattern derivation**: The `deriveAssetPattern` function uses simple string replacement which may not handle all edge cases (e.g., version appearing multiple times in filename). This is acceptable since the LLM provides the mappings and patterns can be manually adjusted.
- **No integration test**: Skipped integration test since it requires ANTHROPIC_API_KEY and would incur costs. Unit tests provide good coverage.

## Test Coverage

- New tests added: 12 test functions
- Tests cover: parseRepo, deriveAssetPattern, Name, CanBuild, fetchReleases, fetchRepoMeta, generateRecipe variants

## Known Limitations

- `deriveAssetPattern` does simple string replacement, may not handle complex cases
- Only supports archives (tar.gz, zip) and binary formats
- No caching of GitHub API responses

## Future Improvements

- Add caching for GitHub API responses to reduce rate limit impact
- Support additional archive formats (tar.xz, tar.bz2)
- Integration test with API key in CI secrets
