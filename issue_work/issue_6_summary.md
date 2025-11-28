# Issue 6 Summary

## What Was Implemented

Go-based integration tests that run tsuku inside Docker containers, using the same test matrix as CI.

## Changes Made

- `Dockerfile.integration`: New Dockerfile for integration testing
  - Ubuntu 22.04 base (matches CI ubuntu-latest)
  - Minimal dependencies (wget, curl, ca-certificates)
  - Non-root test user
  - Pre-built tsuku binary copied in

- `integration_test.go`: Integration test file with `//go:build integration` tag
  - Reads test cases from test-matrix.json
  - Builds tsuku for Linux and creates Docker image
  - Runs each tool installation as a subtest
  - Supports filtering by tool (`-tool=actionlint`) or tier (`-tier=1`)
  - Supports `LIST_TOOLS=1` to show available tests

## Usage

```bash
# Run all integration tests
go test -tags=integration -v ./...

# Run single tool test
go test -tags=integration -v -run TestIntegrationSingle -tool=actionlint ./...

# Skip Docker rebuild (use cached image)
go test -tags=integration -v -run TestIntegrationSingle -tool=boundary -skip-build ./...

# List available tools
LIST_TOOLS=1 go test -tags=integration -v -run TestListTools ./...
```

## Key Decisions

- **Separate Dockerfile**: Created `Dockerfile.integration` rather than modifying existing Vagrant-focused `Dockerfile`
- **Cross-compilation**: Builds Linux binary on any host platform
- **Parallel execution**: Uses `t.Parallel()` for subtests

## Limitations

- macOS tests cannot run in containers (Apple licensing/technical constraints)
- Requires Docker to be installed and running
- Network-dependent (downloads tools from internet)

## Future Work

Created follow-up issue for native macOS testing support.
