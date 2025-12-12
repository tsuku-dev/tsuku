# Issue 403 Summary

## What Was Implemented

Added the `tsuku eval <tool>[@version]` command that generates deterministic installation plans as JSON. The command supports `--os` and `--arch` flags for cross-platform plan generation with strict whitelist validation for security.

## Changes Made

- `cmd/tsuku/eval.go`: New command implementation with:
  - Platform validation functions (ValidateOS, ValidateArch)
  - Recipe loading and executor creation
  - Plan generation via Executor.GeneratePlan()
  - JSON output to stdout, warnings to stderr
- `cmd/tsuku/eval_test.go`: Unit tests for platform validation
- `cmd/tsuku/main.go`: Registered evalCmd with root command

## Key Decisions

- **Whitelist validation for platform flags**: Validates os/arch values against explicit whitelists to prevent path traversal injection via template variables
- **Warnings to stderr**: Non-evaluable action warnings go to stderr to avoid mixing with JSON output on stdout
- **Exit code 2 for usage errors**: Follows existing ExitUsage convention for invalid arguments

## Trade-offs Accepted

- **Network required for checksums**: Plan generation downloads files to compute checksums (per design decision 1A in DESIGN-installation-plans-eval.md)

## Test Coverage

- New tests added: 2 (TestValidateOS, TestValidateArch with 22 sub-tests total)
- Tests cover all valid values and various invalid inputs including path traversal attempts

## Known Limitations

- Cross-platform plans download files but cannot verify binaries work on target platform
- Homebrew bottles skip checksum computation (complex URL resolution)

## Future Improvements

- Issue #404: Store plans in state.json after installation
- Issue #405: Add `tsuku plan show` command
- Issue #406: Add `tsuku plan export` command
