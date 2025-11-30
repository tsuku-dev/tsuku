# Issue #17 Baseline

## Issue
feat(cli): add --json flag for machine-readable output

## Branch
feature/17-json-flag (from main at 97b39a5)

## Test Results
All 16 packages pass:
- github.com/tsuku-dev/tsuku
- github.com/tsuku-dev/tsuku/cmd/tsuku
- github.com/tsuku-dev/tsuku/internal/actions
- github.com/tsuku-dev/tsuku/internal/builders
- github.com/tsuku-dev/tsuku/internal/buildinfo
- github.com/tsuku-dev/tsuku/internal/config
- github.com/tsuku-dev/tsuku/internal/executor
- github.com/tsuku-dev/tsuku/internal/install
- github.com/tsuku-dev/tsuku/internal/progress
- github.com/tsuku-dev/tsuku/internal/recipe
- github.com/tsuku-dev/tsuku/internal/registry
- github.com/tsuku-dev/tsuku/internal/telemetry
- github.com/tsuku-dev/tsuku/internal/testutil
- github.com/tsuku-dev/tsuku/internal/toolchain
- github.com/tsuku-dev/tsuku/internal/userconfig
- github.com/tsuku-dev/tsuku/internal/version

## Commands Requiring JSON Output
Per issue requirements:
- list
- info
- versions
- outdated
- search
