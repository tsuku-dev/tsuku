# Issue 300 Summary

## What Was Implemented

Updated `tsuku list` to display all installed versions of each tool (one line per version) with an `(active)` indicator showing which version is currently symlinked. Output is sorted by tool name, then by version.

## Changes Made

- `internal/install/list.go`:
  - Added `IsActive` field to `InstalledTool` struct
  - Rewrote `ListWithOptions()` to iterate over state.json Versions map instead of directory scanning
  - Added sorting by tool name then version
  - Added directory existence check to filter stale state entries

- `cmd/tsuku/list.go`:
  - Added `(active)` indicator to text output
  - Added `is_active` field to JSON output

- `internal/install/list_test.go`: New test file with multi-version tests

- `internal/install/state_test.go`: Updated existing tests to add state entries (List now uses state.json)

## Key Decisions

- **Use state.json as source of truth**: The list now iterates over state.json's Versions map rather than scanning directories. This is more reliable and provides access to metadata like active version.
- **Filter stale entries**: Versions in state that don't have corresponding directories are filtered out, providing resilience against state/filesystem inconsistencies.

## Trade-offs Accepted

- **Changed output format**: The output now shows one line per version instead of one line per tool. This may break user scripts, but the JSON output provides a stable machine-readable format.

## Test Coverage

- New tests added: 4 (TestListWithOptions_MultiVersion, TestListWithOptions_StaleStateEntries, TestListWithOptions_EmptyVersionsMap, TestListWithOptions_HiddenToolFiltering)
- Existing tests updated: 1 (TestManager_ListWithOptions_WithTools)

## Known Limitations

- None identified

## Future Improvements

- Could add `--versions` flag to show only versions for a specific tool
- Could add version sorting that understands semver (currently lexicographic)
