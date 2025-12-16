# Issue #618 Baseline Assessment

**Issue**: fix(builders): wire Orchestrator into tsuku create command
**Branch**: fix/618-orchestrator-wiring
**Base**: main @ 2d95f6b

## Problem Statement

The Orchestrator (which handles generate → sandbox → repair cycles) was implemented in PR #592
but never wired into `tsuku create`. The `--skip-sandbox` flag exists but does nothing.

Current flow in `cmd/tsuku/create.go`:
1. Creates a SessionBuilder via `builders.NewSession()`
2. Calls `session.Generate()` directly
3. Never instantiates or uses the Orchestrator
4. Dead code: `_ = forceInit` and `_ = skipSandbox` at lines 251-253

## Test Baseline

All tests passing:
- `go test ./...` - PASS
- `go build ./cmd/tsuku` - SUCCESS

## Files to Modify

### Primary Changes
- `cmd/tsuku/create.go` - Wire Orchestrator into recipe creation flow

### Dependencies
- `internal/builders/orchestrator.go` - Already implemented, needs integration
- `internal/sandbox/executor.go` - Already implemented
- `internal/validate/runtime.go` - RuntimeDetector for container detection

## Architecture Reference

From `docs/DESIGN-install-sandbox.md`:
- Orchestrator controls the generate → sandbox → repair loop
- SandboxRequirements computed from installation plan
- DeterministicSession used for ecosystem builders (returns RepairNotSupportedError)
- Single flow for all builders, graceful handling when repair not supported

## Success Criteria

1. `tsuku create foo --from homebrew:bar` runs sandbox testing when Docker available
2. `tsuku create foo --from homebrew:bar --skip-sandbox` skips sandbox testing
3. Sandbox failure triggers repair cycle for LLM-based builders
4. Graceful degradation for ecosystem builders (no repair, sandbox result logged)
5. Missing Docker runtime warns but doesn't fail with `--skip-sandbox`
