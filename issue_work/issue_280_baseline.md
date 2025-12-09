# Issue 280 Baseline

## Environment
- Date: 2025-12-08
- Branch: feature/280-inspect-archive-tool
- Base commit: 48343d0e228f68c7c9b7590ec7e83f9946fb60bc

## Test Results
- Total: 18 packages
- Passed: 17
- Failed: 1 (TestGovulncheck - pre-existing)

## Build Status
PASS - Build succeeds without errors

## Pre-existing Issues
- TestGovulncheck fails due to Go standard library vulnerabilities (GO-2025-4175, GO-2025-4155) in crypto/x509
- This is unrelated to issue #280 and affects the base Go version, not our code
