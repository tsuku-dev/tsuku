# Issue 612 Baseline

## Environment
- Date: 2025-12-17
- Branch: feature/612-go-install-deterministic
- Base commit: 3aa20f7038c0b079828f55cadce544fbda027446

## Test Results
- Command: `go test -test.short ./...`
- Result: FAIL (2 pre-existing failures)

### Failed Tests
1. `github.com/tsukumogami/tsuku/internal/sandbox.TestSandboxIntegration/simple_binary_install`
   - Error: unknown flag: --plan
   - Cause: Sandbox test expects --plan flag that doesn't exist yet

2. `github.com/tsukumogami/tsuku/internal/validate.TestEvalPlanCacheFlow`
   - Error: failed to download for checksum computation: bad status: 404 Not Found
   - Cause: Test depends on external resource that's unavailable

### Passed Packages
All other packages passed:
- github.com/tsukumogami/tsuku
- github.com/tsukumogami/tsuku/cmd/tsuku
- github.com/tsukumogami/tsuku/internal/actions (51.685s)
- github.com/tsukumogami/tsuku/internal/builders
- github.com/tsukumogami/tsuku/internal/buildinfo
- github.com/tsukumogami/tsuku/internal/config
- github.com/tsukumogami/tsuku/internal/errmsg
- github.com/tsukumogami/tsuku/internal/executor (22.201s)
- github.com/tsukumogami/tsuku/internal/httputil
- github.com/tsukumogami/tsuku/internal/install
- github.com/tsukumogami/tsuku/internal/llm
- github.com/tsukumogami/tsuku/internal/log
- github.com/tsukumogami/tsuku/internal/progress
- github.com/tsukumogami/tsuku/internal/recipe
- github.com/tsukumogami/tsuku/internal/registry
- github.com/tsukumogami/tsuku/internal/telemetry
- github.com/tsukumogami/tsuku/internal/testutil
- github.com/tsukumogami/tsuku/internal/toolchain
- github.com/tsukumogami/tsuku/internal/userconfig
- github.com/tsukumogami/tsuku/internal/version (22.518s)

## Build Status
- Command: `go build -o tsuku ./cmd/tsuku`
- Result: PASS (no warnings)

## Coverage
Not measured in baseline (will track changes during implementation)

## Pre-existing Issues
The two failing tests are unrelated to go_install work:
1. Sandbox test failure is due to missing --plan flag in install command
2. Validate test failure is due to external resource unavailability (404 error)

These failures exist before any changes for issue 612.
