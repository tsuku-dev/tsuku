# Issue 226 Summary

## What Was Implemented
Added `--all` flag to `tsuku list` command that includes libraries (from $TSUKU_HOME/libs/) in the output, displayed in a separate section with [lib] marker.

## Changes Made
- `internal/install/library.go`: Added `InstalledLibrary` type and `ListLibraries()` method
- `internal/install/library_test.go`: Added 3 unit tests for ListLibraries
- `cmd/tsuku/list.go`: Added `--all` flag, library listing logic, and JSON output support

## Key Decisions
- Libraries shown in separate section rather than mixed with tools for clarity
- Used [lib] marker consistent with existing output style
- JSON output includes both tools and libraries arrays when `--all` is used

## Trade-offs Accepted
- None significant

## Test Coverage
- New tests added: 3 (TestManager_ListLibraries_Empty, TestManager_ListLibraries, TestManager_ListLibraries_IgnoresFiles)

## Known Limitations
- Libraries already hidden by default since they're in separate `libs/` directory

## Future Improvements
- Could show which tools depend on each library
