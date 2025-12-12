# Issue 404 Summary

## What Was Implemented

Added installation plan storage to state.json. During `tsuku install`, after successful recipe execution, the plan is generated and stored inline with version metadata. This enables plan inspection and deterministic re-installation.

## Changes Made

- `internal/install/state.go`:
  - Added `Plan` field to `VersionState` struct
  - Added `Plan`, `PlanPlatform`, `PlanStep` types for state storage
  - Added `NewPlanFromExecutor` helper for plan creation
- `internal/install/manager.go`:
  - Added `Plan` field to `InstallOptions` struct
  - Updated `InstallWithOptions` to store plan in version state
- `cmd/tsuku/install_deps.go`:
  - Added plan generation after recipe execution
  - Added `convertExecutorPlan` function to convert executor types to install types
- `internal/install/state_test.go`: Added tests for plan storage and backward compatibility
- `cmd/tsuku/install_test.go`: Added tests for plan conversion

## Key Decisions

- **Separate install.Plan type**: Avoids circular imports between executor and install packages
- **Plan stored in VersionState**: Associates plan with specific version, enables multi-version plan tracking
- **Plan generation after execute**: Ensures plan reflects actual installation, not speculative planning
- **Warning on plan generation failure**: Installation continues without plan rather than failing

## Trade-offs Accepted

- **Type duplication**: install.Plan mirrors executor.InstallationPlan to avoid import cycles
- **State file size increase**: Plans add 1-5KB per tool (acceptable per design doc)

## Test Coverage

- New tests added: 4 (TestVersionState_WithPlan, TestVersionState_WithoutPlan_BackwardCompatible, TestNewPlanFromExecutor, TestConvertExecutorPlan)
- All tests pass

## Known Limitations

- Plan generation requires network access for checksum computation
- Plans not generated for already-installed versions (only new installs)

## Future Improvements

- Issue #405: Add `tsuku plan show` command
- Issue #406: Add `tsuku plan export` command
