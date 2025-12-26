# Issue 681 Summary

## What Was Implemented

Refactored composite actions (`download_archive`, `github_archive`, `github_file`) to delegate URL resolution, OS/arch mapping, and checksum computation to the `download` action's `Decompose()` method. This eliminates code duplication and ensures consistent behavior across all download-based actions.

## Changes Made

- `internal/actions/composites.go`:
  - Added `decomposeDownload()` helper function that delegates to `DownloadAction.Decompose()`
  - Refactored `DownloadArchiveAction.Decompose()` to use delegation (removed ~40 lines)
  - Refactored `GitHubArchiveAction.Decompose()` to use delegation (removed ~50 lines)
  - Extracted `resolveAssetName()` helper for GitHub asset pattern resolution
  - Refactored `GitHubFileAction.Decompose()` to use delegation and reuse `resolveAssetName()` (removed ~50 lines)

## Key Decisions

- **Delegation vs extraction**: Chose to delegate to `download.Decompose()` rather than extract shared helper functions. This leverages existing tested code and ensures future improvements to download logic automatically propagate to composites.
- **Keep Preflight() separate**: Composite actions keep their own `Preflight()` methods because they validate action-specific parameters (repo format, asset_pattern) rather than download-specific ones.
- **Reuse resolveAssetName()**: Made `GitHubArchiveAction.resolveAssetName()` a method that can be reused by `GitHubFileAction`, avoiding duplication of the asset pattern expansion and wildcard resolution logic.

## Trade-offs Accepted

- **Slight indirection**: Composite actions now call through `decomposeDownload()` -> `download.Decompose()`, adding one level of indirection. This is acceptable because it eliminates significant code duplication and the performance impact is negligible.

## Test Coverage

- No new tests added (existing tests cover behavior)
- All 22 packages pass tests
- All recipes pass strict validation

## Known Limitations

- `Execute()` methods still have separate implementations. The issue scope was limited to `Decompose()` methods to minimize risk.

## Future Improvements

- Consider aligning `Execute()` methods with `Decompose()` delegation pattern
- Consider extracting `resolveAssetName()` to a package-level function for better reusability
