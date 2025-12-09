# Issue 26 Implementation Summary

## Summary

Added a caching layer for version lists with configurable TTL to reduce API calls and improve performance. The cache wraps version providers using a decorator pattern and stores results in `$TSUKU_HOME/cache/versions/`.

## Changes Made

### New Files

1. **`internal/version/cache.go`**
   - `CachedVersionLister` struct implementing `VersionLister` interface
   - Cache file read/write with atomic writes (write temp, then rename)
   - TTL-based expiration checking
   - `ListVersionsWithCacheInfo()` returns cache hit status
   - `Refresh()` bypasses cache and fetches fresh data
   - `GetCacheInfo()` returns cache metadata without fetching

2. **`internal/version/cache_test.go`**
   - Tests for cache miss, cache hit, cache expiration
   - Tests for refresh functionality
   - Tests for source description delegation
   - Tests for corrupt cache handling
   - Tests for per-source cache isolation

### Modified Files

1. **`internal/config/config.go`**
   - Added `EnvVersionCacheTTL` constant (`TSUKU_VERSION_CACHE_TTL`)
   - Added `DefaultVersionCacheTTL` (1 hour)
   - Added `CacheDir` and `VersionCacheDir` fields to Config struct
   - Added `GetVersionCacheTTL()` function with 5m-7d validation range
   - Updated `DefaultConfig()` to set cache directory paths
   - Updated `EnsureDirectories()` to create cache directories

2. **`internal/config/config_test.go`**
   - Updated `TestEnsureDirectories` to include cache directories
   - Added tests for `GetVersionCacheTTL()`: default, custom, invalid, too low, too high

3. **`cmd/tsuku/versions.go`**
   - Imported `internal/config` package
   - Added `--refresh` flag to bypass cache
   - Wrapped version lister with `CachedVersionLister`
   - Display cache status in output (cached vs fresh)
   - Added `from_cache` field to JSON output

4. **`internal/testutil/testutil.go`**
   - Added `CacheDir` and `VersionCacheDir` to test config
   - Create cache directories in test setup

## Features

- Version lists cached to `$TSUKU_HOME/cache/versions/{hash}.json`
- Default TTL: 1 hour
- Configurable via `TSUKU_VERSION_CACHE_TTL` (5m to 7d range)
- `--refresh` flag bypasses cache
- Output shows cache status:
  - "Using cached versions for X (Y) [expires Z]"
  - "Fetching versions for X (Y)..."
- JSON output includes `from_cache` boolean

## Testing

- All unit tests pass
- Cache tests cover: miss, hit, expiration, refresh, corrupt cache, isolation
- Config tests cover: TTL default, custom, invalid, bounds checking

## Verification

```bash
# Run tests
go test ./...

# Build
go build ./...

# Test cache behavior
tsuku versions kubectl        # First call - fetches fresh
tsuku versions kubectl        # Second call - uses cache
tsuku versions kubectl --refresh  # Force refresh
```
