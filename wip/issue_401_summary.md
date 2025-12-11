# Issue 401 Summary

## What Was Implemented

Added foundational data types for installation plans in the executor package. These types enable deterministic recipe resolution by capturing fully-resolved installation specifications with checksums.

## Changes Made

- `internal/executor/plan.go`: New file with InstallationPlan, Platform, ResolvedStep structs, ActionEvaluability map, and IsActionEvaluable helper
- `internal/executor/plan_test.go`: New file with JSON round-trip serialization tests for all types

## Key Decisions

- **Package location**: Types placed in `internal/executor/` rather than a new `internal/plan/` package because the design places PlanGenerator in executor. Can be extracted later if needed.
- **ActionEvaluability as map**: Used a constant map rather than methods on action types to keep the classification centralized and easy to reference. Unknown actions are treated as non-evaluable for safety.
- **Optional fields with omitempty**: URL, Checksum, and Size use `omitempty` JSON tags since they only apply to download-type steps.

## Trade-offs Accepted

- **No validation methods**: The types don't include validation logic (e.g., checking FormatVersion compatibility). This keeps the foundational types simple; validation can be added in issue #402 (plan generator).

## Test Coverage

- New tests added: 7 test functions
- Tests cover:
  - JSON round-trip for InstallationPlan with all fields
  - JSON round-trip for ResolvedStep with and without optional fields
  - Platform JSON serialization
  - Verification of JSON field names (snake_case)
  - Omit-empty behavior for optional fields
  - IsActionEvaluable for all known actions plus unknown action handling
  - FormatVersion constant value

## Known Limitations

- Types are standalone without integration to the rest of the system. Issue #402 will implement the plan generator that uses these types.
- No validation for format version compatibility - consumers must check FormatVersion themselves.

## Future Improvements

- Add validation method on InstallationPlan for format version checking
- Consider extracting to separate package if plan types are needed across multiple packages
