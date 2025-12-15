# Issue 543 Baseline

## Environment
- Date: 2025-12-15
- Branch: feature/543-build-validation-scripts
- Base commit: 2ac8bc31771034504293c9a796dbaf06ec8495a0

## Test Results
- Total: 23 packages tested
- Passed: All
- Failed: 0

## Build Status
Build successful (go build -o tsuku ./cmd/tsuku)

## Pre-existing Issues
None - all tests pass, build succeeds.

## Issue Summary
Create scripts to validate bottle relocation and tool functionality:
- scripts/verify-relocation.sh - checks RPATH/install_name for hardcoded paths
- scripts/verify-tool.sh - runs tool-specific functional tests
- scripts/verify-no-system-deps.sh - verifies only tsuku/system libc deps
- Scripts must work on both Linux (readelf/ldd) and macOS (otool)
- CI must use scripts for all build essential validation
