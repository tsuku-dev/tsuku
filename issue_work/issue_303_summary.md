# Issue 303 Summary

## Implementation

Added `PreDownloader` to the validate package for downloading release assets before container execution.

### Files Created

- `internal/validate/predownload.go` - PreDownloader implementation
- `internal/validate/predownload_test.go` - 13 unit tests

### Key Features

1. **Streaming SHA256 Computation**: Uses `io.TeeReader` to compute checksum during download rather than re-reading the file
2. **HTTPS-Only Enforcement**: Rejects non-HTTPS URLs for security
3. **SSRF Protection**: Validates redirect targets to prevent Server-Side Request Forgery attacks
4. **Cleanup on Failure**: Automatically removes temp directories when downloads fail
5. **Configurable**: Supports custom temp directories and HTTP clients for testing

### API

```go
// NewPreDownloader creates a new PreDownloader with sensible defaults
func NewPreDownloader() *PreDownloader

// WithTempDir sets a custom temp directory
func (p *PreDownloader) WithTempDir(dir string) *PreDownloader

// WithHTTPClient sets a custom HTTP client (useful for testing)
func (p *PreDownloader) WithHTTPClient(client *http.Client) *PreDownloader

// Download downloads a file and returns result with path, checksum, and size
func (p *PreDownloader) Download(ctx context.Context, url string) (*DownloadResult, error)

// Cleanup removes downloaded file and parent directory
func (r *DownloadResult) Cleanup() error
```

### Test Coverage

- Success cases (checksum, size, filename extraction)
- HTTP errors (404, 500, 403)
- Non-HTTPS rejection
- Compressed response rejection
- Context cancellation
- Cleanup behavior
- IP validation for SSRF protection
- Large file handling (1MB streaming)

## Acceptance Criteria

- [x] `PreDownloader` downloads assets to temp directory
- [x] Computes SHA256 checksum during download
- [x] Returns `DownloadResult` with path, checksum, size
- [x] Handles download errors gracefully
- [x] Cleans up temp files on failure
