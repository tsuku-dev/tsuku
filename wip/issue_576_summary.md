# Issue 576 Summary

## What Was Implemented
Added validation to require SHA256 checksums for URL-based patches, ensuring patch integrity can be verified before installation. The implementation prevents MITM attacks and compromised upstream sources by enforcing checksum verification at recipe parse time.

## Changes Made
- `internal/recipe/types.go`: Added `Checksum` field to `Patch` struct and updated `ToTOML()` serialization
- `internal/recipe/validator.go`: Implemented `validatePatches()` function with checksum requirements and format validation
- `internal/recipe/validator_test.go`: Added 8 comprehensive test cases covering all validation scenarios
- `testdata/recipes/bash-source.toml`: Added checksums to 9 patches
- `testdata/recipes/readline-source.toml`: Added checksums to 3 patches

## Key Decisions
- **Parse-time validation**: Enforced checksum requirement during recipe validation rather than at execution time, providing immediate feedback and preventing invalid recipes from being processed
- **SHA256 format validation**: Validated checksum format (64 hex characters) to catch typos and formatting errors early
- **Inline patches exempt**: Allowed inline patches (with `data` field) to omit checksums since they're embedded in the recipe itself and don't require external download verification

## Trade-offs Accepted
- **Breaking change for existing recipes**: Recipes with URL patches without checksums will now fail validation. This is acceptable because it improves security and existing test recipes were updated as part of this change.
- **No automatic checksum computation**: Recipe authors must manually compute and specify checksums. This ensures intentionality and prevents automatic trust of potentially compromised sources.

## Test Coverage
- New tests added: 8
- All tests passing (23 packages, full test suite)
- Test scenarios covered:
  - URL patch without checksum (validation error)
  - URL patch with valid checksum (pass)
  - Inline patch without checksum (pass)
  - Invalid checksum length (validation error)
  - Non-hexadecimal checksum (validation error)
  - Mutual exclusivity violations (validation error)
  - Missing url/data (validation error)
  - Multiple patches with mixed validation results

## Known Limitations
None - feature is complete and meets all acceptance criteria

## Future Improvements
- Could add helper command to compute patch checksums automatically
- Could support additional hash algorithms (SHA512) if needed
