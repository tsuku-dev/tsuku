# Issue 166 Summary

## Changes Made

Fixed parameter naming mismatch between validator and installer for `github_archive` and `github_file` actions.

### Files Modified
- `internal/recipe/validator.go` - Changed validation to check for `asset_pattern` instead of `asset`
- `internal/recipe/validator_test.go` - Updated test cases to use `asset_pattern`

## Testing
- All unit tests pass
- Build succeeds

## Impact
Unblocks tsuku-dev/tsuku-registry#25 (adding tsuku validate --strict to CI)
