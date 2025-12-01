# Issue 144 Baseline

## Environment
- Date: 2025-11-30
- Branch: feature/144-perl-integration-tests
- Base commit: a0cb9ec (main)

## Test Results
- Total: 17 packages
- Passed: 17
- Failed: 0

## Build Status
Pass

## Notes
- Fixed pre-existing test failure in TestCpanInstallAction_Execute_PerlNotFound
  (same issue as in PR #154, test failed when perl was installed in ~/.tsuku/tools)
