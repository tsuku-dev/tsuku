# Issue 559 Baseline

## Environment
- Date: 2025-12-23 15:26:00
- Branch: feature/559-git-recipe
- Base commit: 45d8d4c (feat(recipes): add readline and sqlite recipes with dependency chain validation)

## Test Results
- Total: All packages tested
- Passed: All except 1 pre-existing failure
- Failed: 1 pre-existing failure
  - TestCargoInstallAction_Decompose (cargo not found - documented pre-existing issue)

## Build Status
✓ Build passes successfully
- Command: `go build -o tsuku ./cmd/tsuku`
- No warnings or errors

## Coverage
Not tracked for baseline (will add tests during implementation)

## Pre-existing Issues
- TestCargoInstallAction_Decompose fails consistently (requires Rust/cargo installed)
- This is a known baseline failure, not related to our work

## Scope
Working on issue #559: Add git recipe to validate complete toolchain

Following the established pattern:
- Production recipe (git.toml): Use homebrew bottles for fast user installs
- Test recipe (git-source.toml): Build from source to validate toolchain
- Comprehensive testing and CI integration
- Update design doc mermaid diagrams

## Dependencies Status
- #554 (curl recipe): CLOSED ✓
- git can proceed
