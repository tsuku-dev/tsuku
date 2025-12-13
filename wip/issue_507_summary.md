# Issue 507 Summary

## What Was Implemented

Added `--plan` flag to `tsuku install` command to enable installation from pre-computed plan files. This completes the deterministic recipe execution milestone by allowing the canonical workflow: `tsuku eval tool | tsuku install --plan -`.

## Changes Made

- `cmd/tsuku/install.go`: Added `--plan` flag, modified command to route plan-based installs, changed Args to ArbitraryArgs to allow zero args with `--plan`
- `cmd/tsuku/install_test.go`: Updated TestInstallCmdUsage to reflect new usage syntax and verify plan examples
- `cmd/tsuku/plan_install.go`: New file with `runPlanBasedInstall()` function that loads, validates, and executes external plans
- `cmd/tsuku/plan_install_test.go`: New tests for plan-based installation (valid plan loading, tool name defaulting, mismatch errors, flag registration)

## Key Decisions

- **Args validation**: Changed from `MinimumNArgs(1)` to `ArbitraryArgs` with manual validation to allow `--plan` without tool name
- **Error codes**: Used `ExitUsage` for invalid argument combinations consistent with other CLI commands
- **Minimal recipe**: Created minimal recipe for executor context (name only) since plan contains all execution details

## Trade-offs Accepted

- **No telemetry for plan-based installs**: Plan-based installs don't emit telemetry events since there's no version resolution happening
- **Simplified state update**: Plan-based installs mark as explicit but don't track dependencies (plans are self-contained)

## Test Coverage

- New tests added: 4 (TestRunPlanBasedInstall_ValidPlan, TestRunPlanBasedInstall_ToolNameFromPlan, TestRunPlanBasedInstall_ToolNameMismatch, TestInstallPlanFlag)
- Uses existing plan_utils tests from #506 for loading/validation coverage

## Known Limitations

- `--dry-run` not supported with `--plan` (plan already shows what will happen)
- Dependencies must be pre-installed for plan-based installation (plan-based install doesn't resolve dependencies)

## Future Improvements

- Plan signing for organizational trust (noted in design doc as future security enhancement)
- Lock file support for team coordination (tracked separately)
