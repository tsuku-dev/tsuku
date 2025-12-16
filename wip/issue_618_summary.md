# Issue #618 Implementation Summary

**Issue**: fix(builders): wire Orchestrator into tsuku create command
**Branch**: fix/618-orchestrator-wiring
**PR**: Pending

## Changes Made

### 1. internal/builders/builder.go
- Added `Is()` method to `RepairNotSupportedError` to enable `errors.Is()` matching

### 2. internal/builders/orchestrator.go
- Updated repair loop to handle `RepairNotSupportedError` gracefully
- When ecosystem builders return this error, the Orchestrator returns `ValidationFailedError` instead of wrapping it as a generic error
- This provides clear feedback when sandbox validation fails for deterministic recipes

### 3. cmd/tsuku/create.go
- Removed dead code (`forceInit` and `skipSandbox` unused variable assignments)
- Added imports for `sandbox`, `validate`, and `errors` packages
- Created sandbox executor when `--skip-sandbox` is not set
- Created Orchestrator with sandbox executor and config
- Replaced direct `session.Generate()` call with `orchestrator.Create()`
- Added detailed error handling for `ValidationFailedError`
- Updated `SandboxSkipped` check to use `orchResult.SandboxSkipped`

## Behavior Changes

### Before
- `tsuku create foo --from homebrew:bar` called `session.Generate()` directly
- No sandbox testing occurred
- `--skip-sandbox` flag was dead code

### After
- `tsuku create foo --from homebrew:bar` uses Orchestrator for generate → sandbox → repair cycle
- Sandbox testing runs when Docker/Podman is available
- `--skip-sandbox` flag properly skips sandbox testing
- LLM builders (homebrew, github) get repair cycles on validation failure
- Ecosystem builders (cargo, pypi, npm, etc.) get clear error on validation failure

## Test Results

- All unit tests pass: `go test ./...`
- Static analysis clean: `go vet ./...`
- Build successful: `go build ./cmd/tsuku`

## Files Changed

| File | Lines Changed | Description |
|------|---------------|-------------|
| internal/builders/builder.go | +6 | Add Is() method to RepairNotSupportedError |
| internal/builders/orchestrator.go | +8 | Handle RepairNotSupportedError in repair loop |
| cmd/tsuku/create.go | +40, -20 | Wire Orchestrator, remove dead code |
