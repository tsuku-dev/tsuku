# Issue 610 Implementation Summary

## Overview
Successfully implemented gem_install evaluable decomposition following the pattern established in issue #609 (pipx_install). The gem_install composite action now decomposes into gem_exec primitive steps with bundler lockfile data.

## Changes Made

### 1. gem_exec.go - Added lock_data mode
**File**: `internal/actions/gem_exec.go`

**Changes**:
- Added lock_data mode for installing from lockfile data (similar to pip_exec)
- Added parameters: `gem`, `version`, `lock_data`, `executables`
- Added `IsDeterministic()` method returning false (residual non-determinism)
- Added `RequiresNetwork()` method returning true
- Implemented `executeLockDataMode()` for lockfile-based installation
- Added `countLockfileGems()` helper for progress reporting
- Writes Gemfile and Gemfile.lock from parameters
- Creates symlinks for executables after installation

**Lines added**: ~200 lines

### 2. gem_install.go - Implemented Decomposable interface
**File**: `internal/actions/gem_install.go`

**Changes**:
- Added Decomposable interface assertion
- Updated Dependencies() to include `EvalTime: ["ruby"]` (bundler comes with ruby)
- Implemented `Decompose()` method:
  - Creates temp directory with Gemfile
  - Runs `bundle lock --add-checksums`
  - Returns gem_exec step with lock_data
- Added helper functions:
  - `findBundlerForEval()` - locates bundler for eval-time decomposition
  - `generateGemfileLock()` - generates lockfile using bundle lock
  - `getRubyVersionForGem()` - extracts Ruby version for metadata

**Lines added**: ~190 lines

### 3. Test Coverage
**Files**: `internal/actions/gem_exec_test.go`, `internal/actions/gem_install_test.go`

**New tests**:
- `TestGemExecAction_LockDataMode_Validation` - validates lock_data parameters
- `TestGemExecAction_LockDataMode_BundlerNotFound` - error handling
- `TestGemExecAction_LockDataMode_WithMockBundler` - successful installation
- `TestGemExecAction_IsDeterministic` - verifies non-deterministic flag
- `TestGemExecAction_RequiresNetwork` - verifies network requirement
- `TestCountLockfileGems_EdgeCases` - lockfile parsing edge cases
- `TestGemInstallAction_Decompose_Validation` - validates Decompose parameters
- `TestGemInstallAction_ImplementsDecomposable` - interface assertion
- `TestGemInstallAction_Dependencies` - verifies EvalTime dependency
- `TestCountLockfileGems` - gem counting in lockfiles

**Test results**: All tests pass (41 gem-related tests)

### 4. Test Matrix Updates
**Files**: `test-matrix.json`, `.github/workflows/sandbox-tests.yml`

**Changes**:
- Added T30 to `ci.linux` array
- Removed T30 from `blocked` section
- Removed T30 exclusion from sandbox-tests.yml

## Testing

### Local Testing
✅ **Build**: `go build -o tsuku ./cmd/tsuku` - successful
✅ **Unit tests**: `go test ./internal/actions/...` - all pass
✅ **Jekyll eval**: `./tsuku eval jekyll --yes` - generates correct plan with gem_exec and lock_data

### Example Output (Jekyll)
```json
{
  "action": "gem_exec",
  "params": {
    "gem": "jekyll",
    "version": "4.4.1",
    "executables": ["jekyll"],
    "lock_data": "GEM\n  remote: https://rubygems.org/\n  specs:\n...",
    "ruby_version": "3.4.7"
  },
  "evaluable": false,
  "deterministic": false
}
```

The lock_data contains the complete Gemfile.lock with:
- Full dependency graph
- Platform specifications
- SHA256 checksums for all gems
- Bundler version used

### Known Limitation: Bundler Self-Installation
**Issue**: Bundler 4.0.1 cannot be installed using bundler 2.6.9 (from Ruby 3.4.7)

**Root cause**: Circular dependency - bundler requires itself to create lockfiles, and version 2.6.9 cannot resolve dependencies for version 4.0.1.

**Error message**:
```
Because the current Bundler version (2.6.9) does not satisfy bundler = 4.0.1
  and Gemfile depends on bundler = 4.0.1,
  version solving has failed.
```

**Impact**: T30 (bundler) test will fail in sandbox tests

**Workaround**: All other gems (jekyll, fpm, etc.) work correctly. Bundler is a unique edge case.

## Files Modified
1. `internal/actions/gem_exec.go` (+200 lines)
2. `internal/actions/gem_install.go` (+190 lines)
3. `internal/actions/gem_exec_test.go` (+250 lines)
4. `internal/actions/gem_install_test.go` (+165 lines)
5. `test-matrix.json` (T30 enabled)
6. `.github/workflows/sandbox-tests.yml` (T30 exclusion removed)

## Acceptance Criteria Status

✅ gem_install.Decompose() generates Gemfile.lock via `bundle lock --add-checksums`
✅ Complete dependency graph with checksums stored in plan
✅ Exec phase uses BUNDLE_FROZEN=true for strict lockfile enforcement
✅ Platform-specific gems handled (captured in lockfile PLATFORMS section)
✅ Ruby version constraint captured in plan (ruby_version parameter)
✅ Pre-compiled gems preferred when available (bundler handles this)
⚠️ T30 (bundler) passes sandbox tests - **Known limitation: bundler self-installation fails**

## Pattern Consistency with #609 (pipx_install)

| Aspect | pipx_install (#609) | gem_install (#610) |
|--------|---------------------|-------------------|
| Primitive | pip_exec | gem_exec |
| Lock mechanism | pip download + requirements.txt | bundle lock + Gemfile.lock |
| Lock parameter | locked_requirements | lock_data |
| Checksum verification | --require-hashes | Bundler CHECKSUMS section |
| Determinism flag | IsDeterministic() = false | IsDeterministic() = false |
| Network requirement | RequiresNetwork() = true | RequiresNetwork() = true |
| EvalTime dependency | python-standalone | ruby |
| Environment variable | PYTHONHASHSEED=0 | SOURCE_DATE_EPOCH=315619200 |

## Next Steps
1. Create PR with implementation
2. Monitor CI checks (expect T30 to fail with known bundler limitation)
3. Consider follow-up issue for bundler edge case if needed

## Commits
- `e473d56` - test: increase code coverage for nix actions
- `193655d` - test: add coverage tests for nix_install and nix_portable
- `f4dc956` - chore: clean up temporary artifacts
- `7bb63a9` - feat(actions): implement nix_realize ecosystem primitive
- `9bc7d49` - docs: establish baseline for nix_realize primitive
- `40a4fcd` - feat(actions): implement pipx_install decomposition (base commit)
- `f0b54d6` - docs: add implementation plan for gem_install decomposition
- `2e2ab73` - feat(actions): implement gem_install evaluable decomposition
