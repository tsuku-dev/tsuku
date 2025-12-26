# Issue #681 Implementation Plan

## Summary

Refactor the composite actions (`download_archive`, `github_archive`, `github_file`) to delegate URL resolution, OS/arch mapping, and checksum computation to the `download` action's `Decompose()` method instead of reimplementing this logic. This eliminates code duplication, centralizes validation, and ensures consistent behavior across all download-based actions.

## Approach

The refactoring will follow a **delegation pattern**: composite actions will construct a `download` action step first, then call `download.Decompose()` to get the resolved `download_file` primitive, before appending their additional steps (extract, chmod, install_binaries).

### Why Delegation Over Other Approaches

1. **Single source of truth**: URL template expansion, OS/arch mapping, and checksum computation are defined once in `download.Decompose()`
2. **Automatic validation inheritance**: Composites automatically benefit from `download.Preflight()` validations
3. **Reduced maintenance burden**: Bug fixes and enhancements to download logic automatically propagate to all composites
4. **Clean layering**: Each action focuses only on its unique concerns

### Key Observations from Code Analysis

1. **`download.Decompose()`** (lines 77-160 of download.go):
   - Expands URL template with variables: `{version}`, `{version_tag}`, `{os}`, `{arch}`
   - Applies `os_mapping` and `arch_mapping` via `ApplyMapping()`
   - Downloads file to compute checksum if `ctx.Downloader` is available
   - Returns single `download_file` step with resolved URL, dest, and checksum

2. **Composite actions' `Decompose()`** (composites.go):
   - `download_archive` (lines 193-323): Manually expands URL, applies mappings, downloads for checksum
   - `github_archive` (lines 522-670): Same pattern + GitHub URL construction + wildcard resolution
   - `github_file` (lines 849-978): Same pattern + single binary handling

3. **Duplicated code patterns**:
   - URL variable expansion using `ExpandVars()` and `vars` map
   - OS/arch mapping with type assertions on `params["os_mapping"]`
   - Checksum computation via `ctx.Downloader.Download()`
   - Cache save and cleanup logic

4. **`Preflight()` validation** (composites.go):
   - Each composite implements its own `os_mapping`/`arch_mapping` warnings
   - `download.Preflight()` already has comprehensive validation

## Alternatives Considered

### Alternative 1: Keep Separate Implementations
- **Pro**: No risk of regression during refactor
- **Con**: Continued code duplication, risk of drift, higher maintenance

### Alternative 2: Extract Shared Helper Function
- **Pro**: Moderate code reduction without changing architecture
- **Con**: Still two places to maintain (helper + download), doesn't leverage existing `download.Decompose()`

### Alternative 3: Delegation to `download.Decompose()` (Chosen)
- **Pro**: Maximum code reuse, automatic validation inheritance, clean architecture
- **Con**: Requires careful handling of GitHub-specific URL construction before delegation

## Files to Modify

| File | Changes |
|------|---------|
| `internal/actions/composites.go` | Refactor `Decompose()` methods for all 3 composites; update `Preflight()` to delegate where appropriate |
| `internal/actions/download.go` | No changes required (already has required functionality) |
| `internal/actions/composites_test.go` | Update tests to verify delegation behavior |
| `internal/actions/composites_decompose_test.go` | Verify step outputs remain identical |
| `internal/actions/preflight_test.go` | Verify delegated validation still works |

## Files to Create

None.

## Implementation Steps

### Step 1: Add Helper Method to Create Download Params
Add a helper function to construct download action params from composite params, extracting the common mapping logic.

```go
// buildDownloadParams creates params suitable for download action from composite params.
// It extracts os_mapping and arch_mapping and constructs the URL parameter.
func buildDownloadParams(url string, dest string, params map[string]interface{}) map[string]interface{} {
    downloadParams := map[string]interface{}{
        "url": url,
    }
    if dest != "" {
        downloadParams["dest"] = dest
    }
    if osMapping, ok := params["os_mapping"]; ok {
        downloadParams["os_mapping"] = osMapping
    }
    if archMapping, ok := params["arch_mapping"]; ok {
        downloadParams["arch_mapping"] = archMapping
    }
    return downloadParams
}
```

### Step 2: Refactor `DownloadArchiveAction.Decompose()`

Current structure:
1. Extract params
2. Build vars map
3. Apply OS/arch mapping manually
4. Expand URL
5. Download for checksum
6. Build 4 steps

New structure:
1. Extract archive-specific params (format, binaries, strip_dirs, install_mode)
2. Build download params with URL and mappings
3. Call `download.Decompose()` to get resolved `download_file` step
4. Build remaining 3 steps (extract, chmod, install_binaries)
5. Return combined steps

### Step 3: Refactor `GitHubArchiveAction.Decompose()`

Current structure:
1. Extract params
2. Build vars map
3. Apply OS/arch mapping manually
4. Expand asset pattern
5. Resolve wildcards via GitHub API (if present)
6. Construct GitHub URL
7. Download for checksum
8. Build 4 steps

New structure:
1. Extract archive-specific params
2. Build vars map with mappings for asset pattern expansion (wildcards need pre-expanded vars)
3. Expand asset pattern and resolve wildcards
4. Construct GitHub URL
5. Build download params (URL already resolved, no {version} etc. - pass empty mappings)
6. Call `download.Decompose()` to get `download_file` step
7. Build remaining 3 steps
8. Return combined steps

**Note**: `github_archive` must pre-resolve the URL because:
- GitHub URL is constructed from `repo + version_tag + assetName`
- Asset pattern may contain wildcards requiring API resolution
- Resulting URL has no template variables (fully resolved)

This means `github_archive` will construct the final URL first, then delegate to `download.Decompose()` primarily for checksum computation. This is still valuable as it centralizes the download/cache logic.

### Step 4: Refactor `GitHubFileAction.Decompose()`

Same approach as `github_archive` but simpler (no extract step):
1. Extract params, resolve asset name
2. Construct GitHub URL
3. Build download params with resolved URL
4. Call `download.Decompose()`
5. Append chmod and install_binaries steps

### Step 5: Refactor Preflight Methods

Update composite `Preflight()` methods to delegate validation where appropriate:

For `download_archive.Preflight()`:
- Keep URL required check
- Delegate os_mapping/arch_mapping warnings to download.Preflight() by building download params and calling it
- Keep archive_format redundancy warning (archive-specific)

For `github_archive.Preflight()` and `github_file.Preflight()`:
- Keep repo format validation (GitHub-specific)
- Keep asset_pattern required check
- Delegate os_mapping/arch_mapping warnings (check against asset_pattern, not URL)
- Note: Cannot fully delegate to download.Preflight() because asset_pattern != URL

**Decision**: Keep os_mapping/arch_mapping validation in composites because:
- They check against `asset_pattern` not `url`
- `download.Preflight()` specifically checks against `url` parameter
- The validation is already minimal and consistent

### Step 6: Update Tests

1. **composites_test.go**:
   - Add tests verifying delegation produces same output as direct implementation
   - Test edge cases (empty mappings, wildcard patterns)

2. **composites_decompose_test.go**:
   - Verify step count and order unchanged
   - Verify download_file params match expected values
   - Verify checksum/size fields populated when downloader available

3. **preflight_test.go**:
   - Verify existing validation tests still pass
   - Add test for validation consistency between download and composites

### Step 7: Run Strict Mode Recipe Validation

After refactoring, run validation on all recipes to ensure no regressions:
```bash
go test ./internal/actions/... -v
./tsuku validate --strict recipes/
```

## Testing Strategy

### Unit Tests
1. Verify `Decompose()` returns correct number of steps (4 for archives, 3 for file)
2. Verify first step is always `download_file` with properly resolved URL
3. Verify os_mapping and arch_mapping are applied correctly
4. Verify checksum and size are populated when downloader is available

### Integration Tests
1. Test `download_archive` with tar.gz archive
2. Test `github_archive` with real GitHub repo pattern
3. Test `github_file` with both binary and binaries params

### Regression Tests
1. Run existing test suite: `go test ./internal/actions/...`
2. Validate all recipes: `./tsuku validate recipes/`
3. Install test with a recipe using each composite action

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Behavioral regression in URL expansion | Keep existing tests, add new tests comparing old vs new behavior |
| GitHub-specific logic doesn't fit delegation model | Pre-resolve GitHub URLs before delegating (documented in Step 3) |
| Performance impact from extra Decompose call | Minimal - one extra function call, no I/O difference |
| Validation messages change | Preserve existing validation by keeping composite-specific checks |

## Success Criteria

1. All existing tests pass
2. `Decompose()` methods reduced by ~50% in LOC
3. URL expansion, OS/arch mapping, and checksum computation code appears only in `download.go`
4. All recipes pass `--strict` validation
5. Behavior identical: same steps, same params, same checksums

## Open Questions

1. **Should Execute() methods also delegate?**
   - Current answer: No - Execute() methods work fine and benefit from caching/progress display
   - Decompose() is the critical path for plan generation; Execute() is only for direct runs
   - Future work could align them but it's out of scope for this issue

2. **Should we add a helper for GitHub URL construction?**
   - Current answer: Not necessary - the pattern `https://github.com/{repo}/releases/download/{tag}/{asset}` is simple
   - Could be considered for future cleanup

3. **How to handle download action's static URL error in delegation?**
   - The `download` action returns an error for URLs without variables
   - Composites constructing fully-resolved URLs (github_*) would trigger this
   - Solution: Skip calling `download.Decompose()` entirely, just reuse the checksum computation pattern directly since the URL is already resolved
