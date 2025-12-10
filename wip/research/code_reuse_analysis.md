# Code Reuse Analysis: tsuku eval vs LLM Validation

**Date**: 2025-12-09
**Purpose**: Identify how `tsuku eval` can reuse existing infrastructure

## Key Finding: Two Systems with Similar Needs

| Capability | LLM Validation | tsuku eval |
|------------|----------------|------------|
| Pre-download assets | Yes (PreDownloader) | Yes (for checksum capture) |
| Compute checksums | Yes (SHA256 during download) | Yes (for plan generation) |
| Network isolation | Yes (container --network=none) | No (but useful for testing) |
| Execute recipe | Yes (container) | No (eval only, no exec) |
| Produce artifact | ValidationResult | InstallationPlan |

## Reusable Components

### 1. PreDownloader (`internal/validate/predownload.go`)

**Current interface:**
```go
type PreDownloader struct { ... }

func (p *PreDownloader) Download(ctx context.Context, url string) (*DownloadResult, error)

type DownloadResult struct {
    AssetPath  string  // Local path to downloaded file
    Checksum   string  // SHA256 hex-encoded
    Size       int64   // File size in bytes
}
```

**Reuse for eval:**
- Call `PreDownloader.Download()` for each download URL in recipe
- Capture `Checksum` and `Size` for installation plan
- Either keep file (for immediate install) or discard (for plan-only mode)

### 2. Download Action Security (`internal/actions/download.go`)

**Current capabilities:**
- HTTPS enforcement
- SSRF protection (blocks private IPs on redirects)
- Decompression bomb prevention
- Variable expansion ({version}, {os}, {arch})

**Reuse for eval:**
- URL expansion logic already exists
- Security protections should apply to eval downloads too
- Cache integration for efficiency

### 3. Download Cache (`internal/actions/download_cache.go`)

**Current interface:**
```go
type DownloadCache struct { ... }

func (c *DownloadCache) Check(url string, expectedChecksum string) (string, bool, error)
func (c *DownloadCache) Save(url, filePath, checksum string) error
```

**Reuse for eval:**
- Cache can serve eval requests (avoid re-downloading)
- Checksums stored with cached files

## Architecture Options

### Option A: Reuse PreDownloader Directly

```
tsuku eval ripgrep@14.1.0
    │
    ├── Version Resolution (existing)
    │
    ├── URL Expansion (existing in download action)
    │
    └── PreDownloader.Download(url)
        │
        └── Return DownloadResult with checksum
            │
            └── Build InstallationPlan
```

**Pros:**
- Minimal new code
- Security protections included
- Already tested

**Cons:**
- PreDownloader is in validate package (coupling concern)
- Always downloads (no dry-run without network)

### Option B: Extract Shared Download-with-Checksum Module

Create `internal/download/` package with shared components:

```go
// internal/download/downloader.go
type Downloader interface {
    Download(ctx context.Context, url string) (*Result, error)
    DownloadWithChecksum(ctx context.Context, url string) (*Result, error)
}

type Result struct {
    Path     string
    Checksum string
    Size     int64
}
```

Both `validate.PreDownloader` and `actions.Download` use this.

**Pros:**
- Clean separation
- Single source of truth for secure downloads
- Testable in isolation

**Cons:**
- Refactoring effort
- Risk of breaking existing code

### Option C: Extend Download Action with Plan Mode

Add `--capture-checksum` mode to existing download action:

```go
type DownloadResult struct {
    FilePath string
    Checksum string  // NEW: populated when capture enabled
    Size     int64   // NEW: populated when capture enabled
}
```

**Pros:**
- Minimal change to existing code
- Natural integration with executor

**Cons:**
- Mixes concerns (action vs planning)
- Download action already complex

## Recommendation

**Short-term (for this design):** Option A - Reuse PreDownloader directly

The PreDownloader is well-designed and immediately usable. The coupling concern is acceptable because:
1. Both eval and validate are "pre-flight" operations
2. PreDownloader has no validate-specific dependencies
3. Can refactor to Option B later if needed

**Long-term:** Consider Option B when:
- Multiple consumers need download-with-checksum
- PreDownloader needs features that don't apply to eval
- Package structure becomes confusing

## Air-Gapped Installation Alignment

The LLM validation design already supports air-gapped scenarios:

```
Online Environment:
  tsuku eval ripgrep@14.1.0 --output plan.json
      │
      └── Downloads assets, computes checksums
          │
          └── Outputs plan with URLs, checksums, sizes

Air-Gapped Environment:
  # User manually downloads assets to local directory
  tsuku install --plan plan.json --assets ./downloaded/
      │
      └── Reads plan, verifies checksums against local files
          │
          └── Installs from local files (no network)
```

This matches the LLM validation flow:
1. PreDownload assets (network required)
2. Execute in container with --network=none (network isolated)

For `tsuku eval`, we can use the same pattern:
1. Eval phase: Download and compute checksums (network required)
2. Exec phase: Install from plan with local assets (network optional)

## Files to Reference

| File | Relevance |
|------|-----------|
| `internal/validate/predownload.go` | PreDownloader implementation |
| `internal/validate/executor.go` | GetAssetChecksum() method |
| `internal/actions/download.go` | URL expansion, security checks |
| `internal/actions/download_cache.go` | Caching infrastructure |
| `internal/executor/executor.go` | DryRun() for comparison |

## Questions for Design

1. Should `tsuku eval` actually download files, or just resolve URLs?
   - If download: Reuse PreDownloader, get real checksums
   - If no download: Can't verify checksums until install time

2. Should eval results be cached?
   - Plans could be stored in state.json (current design)
   - Or in dedicated plan cache

3. How to handle multi-platform eval?
   - Current platform only? (simpler)
   - All platforms via API? (more useful for teams)
