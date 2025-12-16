# Issue 610 Implementation Plan

## Summary

Implement gem_install evaluable decomposition following the pattern established in issue #609 (pipx_install).
The gem_install composite action will decompose into gem_exec primitive steps with bundler lockfile data.

## Pattern Reference

Following the pipx_install decomposition from #609:
- pipx_install.Decompose() generates locked requirements via pip download
- Returns pip_exec step with `locked_requirements` parameter
- pip_exec uses PYTHONHASHSEED=0 and --require-hashes for determinism

For gem_install:
- gem_install.Decompose() will generate Gemfile.lock via `bundle lock --add-checksums`
- Returns gem_exec step with `lock_data` parameter
- gem_exec uses BUNDLE_FROZEN=true and SOURCE_DATE_EPOCH for determinism

## Current State Analysis

### gem_exec.go (exists)
- Currently expects `source_dir` with existing Gemfile/Gemfile.lock
- Uses `command` parameter (e.g., "install")
- Already has BUNDLE_FROZEN and SOURCE_DATE_EPOCH support
- **Needs modification**: Add support for `lock_data` parameter to write Gemfile/Gemfile.lock

### gem_install.go (exists)
- Uses direct `gem install` command
- Has ruby as dependency
- **Needs modification**: Implement Decomposable interface with Decompose() method
- **Needs modification**: Add EvalTime dependency on bundler for `bundle lock`

### decomposable.go
- gem_exec already registered as primitive
- No changes needed

## Files to Modify

1. **internal/actions/gem_exec.go**
   - Add `lock_data` parameter support
   - Add `gem` and `version` parameters for Gemfile generation
   - Write Gemfile from gem/version params
   - Write Gemfile.lock from lock_data param
   - Add `executables` parameter for verification

2. **internal/actions/gem_install.go**
   - Implement Decomposable interface
   - Add Decompose() method that:
     - Creates temp directory with Gemfile
     - Runs `bundle lock --add-checksums`
     - Parses Gemfile.lock
     - Returns gem_exec step with lock_data
   - Update Dependencies() to include EvalTime: ["ruby"] (bundler comes with ruby)

3. **internal/actions/gem_exec_test.go**
   - Add tests for lock_data parameter
   - Add tests for gem/version Gemfile generation

4. **internal/actions/gem_install_test.go** (may need to create)
   - Add tests for Decompose() method

5. **test-matrix.json**
   - Move T30 from blocked to linux array
   - Remove T30 from blocked section

6. **.github/workflows/sandbox-tests.yml**
   - Remove T30 from exclude list if present

## Implementation Steps

### Step 1: Update gem_exec to support lock_data
- [ ] Add parameters: gem, version, lock_data, executables
- [ ] When lock_data is provided, write Gemfile and Gemfile.lock to source_dir
- [ ] Generate Gemfile from gem/version: `source 'https://rubygems.org'\ngem '<gem>', '= <version>'`
- [ ] Create symlinks for executables after installation

### Step 2: Implement gem_install.Decompose()
- [ ] Add Decomposable interface assertion
- [ ] Update Dependencies() to add EvalTime: ["ruby"]
- [ ] Implement Decompose():
  - Find bundler (from ruby installation)
  - Create temp directory with Gemfile
  - Run `bundle lock --add-checksums`
  - Read Gemfile.lock content
  - Return gem_exec step with lock_data

### Step 3: Update tests
- [ ] Add gem_exec tests for lock_data mode
- [ ] Add gem_install.Decompose() tests
- [ ] Verify IsPrimitive("gem_exec") still works

### Step 4: Enable T30 in test matrix
- [ ] Update test-matrix.json: add "T30" to ci.linux array
- [ ] Update test-matrix.json: remove T30 from blocked section
- [ ] Remove T30 exclusion from sandbox-tests.yml if present

### Step 5: Local testing
- [ ] Build: `go build -o tsuku ./cmd/tsuku`
- [ ] Run: `go test ./internal/actions/...`
- [ ] Run: `./tsuku eval bundler --yes > bundler-plan.json`
- [ ] Verify plan has gem_exec with lock_data
- [ ] Run: `./tsuku install --plan bundler-plan.json --force`
- [ ] Verify bundler works: `~/.tsuku/bin/bundle --version`

## Success Criteria

1. `gem_install.Decompose()` returns gem_exec step with lock_data
2. `./tsuku eval bundler --yes` produces evaluable plan
3. T30 (bundler) passes sandbox tests
4. All existing gem_install functionality preserved
5. Unit tests pass: `go test ./internal/actions/...`

## Technical Details

### Gemfile.lock Generation
```bash
# Eval phase - generate locked dependencies
cd /tmp/gem-eval
cat > Gemfile <<EOF
source 'https://rubygems.org'
gem 'bundler', '= 4.0.1'
EOF
bundle lock --add-checksums
# Read Gemfile.lock content
```

### gem_exec Parameters (updated)
```go
// When lock_data is provided, gem_exec operates in "install from lock" mode:
// - gem: gem name (for Gemfile generation)
// - version: gem version (for Gemfile generation)
// - lock_data: complete Gemfile.lock content
// - executables: list of executables to verify and symlink
// - ruby_version: optional Ruby version validation
```

### Environment Variables for Determinism
- BUNDLE_FROZEN=true - strict lockfile enforcement
- SOURCE_DATE_EPOCH=315619200 - reproducible timestamps (1980-01-01)
- GEM_HOME/GEM_PATH - isolated installation
