# Issue 24 Summary

## What Was Implemented

Added a `tsuku validate` command that validates recipe files without attempting to install them. The command checks for TOML syntax errors, missing required fields, invalid actions, and security issues.

## Changes Made

- `internal/recipe/validator.go`: New comprehensive validation module with:
  - ValidationResult type with errors and warnings
  - ValidateFile and ValidateBytes functions
  - Action type validation with typo suggestions
  - Action parameter validation (required params per action)
  - URL scheme validation (http/https only)
  - Path traversal detection
  - Version source validation

- `internal/recipe/validator_test.go`: Unit tests covering:
  - Valid recipe validation
  - Missing required fields
  - Unknown actions with typo suggestions
  - Missing action parameters
  - Security checks (URL schemes, path traversal)
  - Warning generation

- `cmd/tsuku/validate.go`: New CLI command with:
  - File path argument
  - --json flag for structured output
  - Human-readable output with error/warning formatting
  - Exit code handling

- `cmd/tsuku/main.go`: Registered the new command

## Key Decisions

- **Separate validator module**: Created `validator.go` instead of extending the existing `validate()` function in loader.go to keep validation logic modular and testable.

- **Static action list**: Used a hardcoded list of known actions instead of querying the registry at runtime. This avoids circular dependencies and provides instant validation.

- **Warnings vs errors**: Implemented both categories - errors block validation (invalid recipe), warnings are informational (missing description, missing {version} placeholder).

- **Typo suggestions**: Added Levenshtein distance-based suggestions for unknown actions (e.g., "did you mean 'download_archive'?").

## Trade-offs Accepted

- Action list may get out of sync if new actions are added without updating validator. Acceptable since new actions are rare and the validator provides clear "unknown action" errors.

## Test Coverage

- New tests added: 15 test functions in validator_test.go
- Coverage for recipe package: Maintained at 91%

## Known Limitations

- Security checks are pattern-based, not exhaustive (e.g., can detect "../" but not encoded variants).
- No validation of URL accessibility (by design - validation is offline).

## Future Improvements

- Consider loading action registry dynamically for more accurate action validation.
- Add schema validation for complex parameter structures.
