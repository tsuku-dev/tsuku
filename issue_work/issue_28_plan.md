# Issue 28 Implementation Plan

## Summary

Add download caching to store downloaded archives in `$TSUKU_HOME/cache/downloads/` and reuse them for reinstalls or retries. Cache entries are verified by checksum before reuse.

## Approach

Integrate caching into the existing `DownloadAction` by checking the cache before downloading and saving to cache after successful downloads. Use URL hash as cache key and store metadata (URL, checksum, size, timestamp) alongside the cached file.

### Alternatives Considered

- **Separate cache layer wrapping DownloadAction**: Would require refactoring action registration and execution. Not chosen because it adds complexity without benefit.
- **HTTP Range requests for resume**: Deferred to future work. Adds significant complexity (partial file tracking, server support detection) and is less impactful than basic caching.
- **LRU eviction with size limits**: Deferred to future work. Basic caching provides immediate value; eviction can be added later.

## Files to Modify

- `internal/config/config.go` - Add `DownloadCacheDir` field to Config struct
- `internal/config/config_test.go` - Update tests for new directory
- `internal/testutil/testutil.go` - Add `DownloadCacheDir` to test config
- `internal/actions/download.go` - Integrate cache check/save into download flow

## Files to Create

- `internal/actions/download_cache.go` - Cache implementation (check, save, clear functions)
- `internal/actions/download_cache_test.go` - Tests for cache functionality
- `cmd/tsuku/cache.go` - New `cache` command with `clear` subcommand

## Implementation Steps

- [ ] Add `DownloadCacheDir` to Config struct and update related code
- [ ] Create download cache implementation with check/save/clear functions
- [ ] Integrate cache into DownloadAction (check before download, save after)
- [ ] Add `tsuku cache clear` command
- [ ] Write tests for cache behavior

## Testing Strategy

- Unit tests: Cache hit/miss, checksum validation, metadata handling, clear functionality
- Unit tests: DownloadAction integration with cache (mock HTTP responses)
- Manual verification: Install tool twice, observe cache usage on second install

## Risks and Mitigations

- **Stale cache with wrong checksum**: Mitigated by always verifying checksum before returning cached file; re-download if mismatch
- **Disk space**: Deferred; basic implementation has no eviction. Future work can add `tsuku cache clear` with size-based eviction
- **Concurrent access**: Use atomic writes (temp file + rename) to prevent corruption

## Success Criteria

- [ ] Downloaded archives cached to `$TSUKU_HOME/cache/downloads/`
- [ ] Cache verified by checksum before reuse
- [ ] `tsuku cache clear` removes cached downloads
- [ ] All tests pass
- [ ] Manual test shows cache hit on second install

## Open Questions

None - requirements are clear from the issue. HTTP Range resume support deferred to separate issue.
