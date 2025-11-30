# Issue 14 Implementation Plan

## Overview
Add progress bars for downloads showing percentage, size, speed, and ETA.

## Analysis

### Download Locations
1. **Primary**: `internal/actions/download.go` - DownloadAction.downloadFile()
2. **Secondary**: `internal/actions/nix_portable.go` - downloadFile() (private function)

### Requirements
- Display: percentage, downloaded/total size, speed, ETA
- Suppress in non-TTY environments (CI)
- Suppress when --quiet flag is set (future integration)

### Dependencies
Need to add `golang.org/x/term` for TTY detection.

## Implementation

### Step 1: Create progress writer package
Create `internal/progress/progress.go`:
- `Writer` struct wrapping io.Writer with progress tracking
- `NewWriter(w io.Writer, total int64, output io.Writer) *Writer`
- Write method that updates progress display
- Format: `[=================>            ] 52% (44MB/85MB) 2.3MB/s ETA: 18s`
- `IsTerminal(fd int) bool` for TTY detection using golang.org/x/term

### Step 2: Update download.go
Modify `downloadFile()` to:
1. Get Content-Length from response headers
2. Check if stdout is a TTY
3. Wrap response.Body with progress writer if TTY
4. Fall back to silent copy if not TTY or Content-Length unknown

### Step 3: Update nix_portable.go
Apply same pattern to the private downloadFile() function.

### Step 4: Add tests
- Test progress calculation logic
- Test format output
- Test TTY detection (via dependency injection)

## File Changes

| File | Change |
|------|--------|
| `go.mod` | Add golang.org/x/term |
| `internal/progress/progress.go` | New - progress writer |
| `internal/progress/progress_test.go` | New - tests |
| `internal/actions/download.go` | Use progress writer |
| `internal/actions/nix_portable.go` | Use progress writer |

## Testing Strategy
- Unit tests for progress formatting
- Integration test via manual verification (progress bars are visual)
- CI tests should pass (progress suppressed in non-TTY)

## Success Criteria
1. Progress bar displays during downloads on TTY
2. Shows percentage, size, speed, ETA
3. Suppressed in non-TTY (CI passes)
4. All existing tests pass
