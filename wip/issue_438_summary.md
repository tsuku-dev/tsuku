# Issue 438 Summary

## What Was Done

Implemented `Decompose()` method for `GitHubArchiveAction` to enable deterministic plan generation by resolving assets and computing checksums at evaluation time.

## Changes Made

### internal/actions/decomposable.go
- Added `Downloader` interface for checksum computation during decomposition
- Added `DownloadResult` struct containing asset path, checksum, and size
- Extended `EvalContext` with `Context` and `Downloader` fields

### internal/actions/composites.go
- Added compile-time interface check: `var _ Decomposable = (*GitHubArchiveAction)(nil)`
- Implemented `Decompose()` method that:
  - Validates required parameters (repo, asset_pattern, archive_format, binaries)
  - Applies OS/arch mappings for platform-specific asset names
  - Resolves wildcard patterns via GitHub API when Resolver is available
  - Downloads file to compute checksum when Downloader is available
  - Returns 4 primitive steps: download, extract, chmod, install_binaries

### internal/actions/decomposable_test.go
- Updated `TestEvalContextStruct` for new fields
- Added `TestDownloadResultStruct`

### internal/actions/composites_test.go
- Added `TestGitHubArchiveAction_Decompose` (basic decomposition)
- Added `TestGitHubArchiveAction_Decompose_MissingParams` (error cases)
- Added `TestGitHubArchiveAction_Decompose_OSArchMapping` (platform mapping)
- Added `TestGitHubArchiveAction_Decompose_InstallMode` (mode passthrough)
- Added `TestGitHubArchiveAction_Decompose_AllStepsArePrimitives` (primitive verification)
- Added `TestGitHubArchiveAction_Decompose_BinariesFormats` (format handling)

## Testing

- All tests pass: `go test ./internal/actions/... -v`
- No lint issues: `go vet ./internal/actions/...`
- Build succeeds: `go build ./internal/actions/...`

## Design Decisions

1. **Downloader interface in actions package**: Avoided import cycle by defining the interface locally rather than importing from validate package
2. **Optional checksum computation**: Decompose works without Downloader (returns empty checksum/size) for flexibility
3. **Wildcard resolution requires Resolver**: Returns error if wildcards present but no Resolver available
