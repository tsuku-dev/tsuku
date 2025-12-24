# Dead Code Scanner Report: Milestone M19

**Milestone**: 19 - Dependency Provisioning: Build Environment
**Scan Date**: 2025-12-23
**Issues in M19**: #547-#556

## Summary

Status: **FINDINGS**
Finding Count: **6**

The scan identified 6 instances of dead code artifacts requiring attention:
1. One debug print statement in pip_exec.go
2. Four TODO comments referencing future work (#643, #644, #660)
3. One research artifact file in wip/ directory from M19 validation

No TODOs referencing closed M19 issues (#547-#556) were found, indicating proper cleanup during implementation.

## Scan Methodology

### 1. TODO/FIXME Comments for Closed Issues
Searched for TODO, FIXME, HACK, XXX comments referencing:
- Milestone 19 explicitly (M19, milestone 19)
- Issue numbers #547-#556 (all M19 issues)

**Result**: No matches found ✓

### 2. Debug Code Patterns
Searched for:
- `fmt.Println`, `log.Println` with "debug" messages
- `console.log` patterns
- Commented-out print statements

**Result**: 1 finding (see below)

### 3. Unused Feature Flags
Searched for:
- Feature flag patterns (FEATURE_, EnableFeature, etc.)
- Always-on/off conditions

**Result**: No feature flags found ✓

### 4. Leftover Test Artifacts
Searched for:
- `.only` or `.skip` markers specific to M19 issues
- Disabled tests referencing milestone issues
- Build-related skip conditions

**Result**: No M19-specific test artifacts found ✓

### 5. Temporary Files
Checked for:
- .bak, .orig, .swp files
- Temporary recipe files
- Research artifacts in wip/

**Result**: 1 wip/ file found (see below)

## Detailed Findings

### Finding 1: Debug Print Statement
**File**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/pip_exec.go:219`
**Severity**: Low
**Type**: Debug output

```go
fmt.Printf("   Debug: skipped shebang fix for %s: %v\n", file.Name(), err)
```

**Context**: This debug print appears in the shebang fixing logic for Python virtual environment scripts. The message prefix "Debug:" suggests this was added for development/debugging purposes.

**Recommendation**:
- Convert to proper structured logging using the logger interface
- Or remove if the information is not needed for users
- The error is non-fatal (logged as warning), so proper logging would be appropriate

**Related to M19**: No - this is in pip_exec.go which is Python ecosystem code, not build environment

### Finding 2-5: TODO Comments for Future Work
**Severity**: Informational
**Type**: Future work tracking

These TODOs reference open issues for future improvements, not M19 issues:

#### TODO #643: Platform-Conditional Dependencies
**Files**:
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew.go:33`
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew_relocate.go:23`

```go
// TODO(#643): Use platform-conditional dependencies to only install patchelf on Linux.
// Currently installed on all platforms for consistency, but only used on Linux.
```

**Status**: Issue #643 is open and tracked in milestone 21 (System-Required)
**Action**: No action needed - proper forward reference

#### TODO #644: Aggregate Primitive Dependencies
**Files**:
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew.go:31`
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew_relocate.go:21`
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/resolver.go:63`

```go
// TODO(#644): Remove this method once composite actions automatically aggregate primitive dependencies.
// This is a workaround because dependency resolution happens before decomposition.
```

**Status**: Issue #644 is open and tracked in milestone 21 (System-Required)
**Action**: No action needed - proper forward reference

#### TODO #660: SQLite Version Format
**File**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/testdata/recipes/sqlite-source.toml:15`

```toml
# TODO(#660): SQLite uses non-standard version format (3.51.1 -> 3510100) requiring manual URL updates
```

**Status**: This is a known limitation documented in a test recipe
**Action**: No action needed - proper documentation of limitation

### Finding 6: WIP Research Artifact
**File**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/wip/research/complete-milestone_M19_validate_metadata.md`
**Severity**: Medium
**Type**: Leftover artifact

**Description**: This is a research artifact from a previous M19 validation step (metadata validation).

**Recommendation**: This file should be cleaned up before PR merge as per project conventions:
> The `wip/` directory holds temporary artifacts during workflows and must be cleaned before PR merge. CI enforces this check.

**Action**: Remove this file as part of M19 completion

## Milestone-Specific Analysis

### M19 Issues Coverage
All issues in M19 (Build Environment milestone):
- #547: feat(actions): declare implicit dependencies for build actions - **DONE**
- #548: feat(recipes): add pkg-config recipe using homebrew - **DONE**
- #549: feat(recipes): add cmake recipe using homebrew - **DONE**
- #550: feat(actions): enhance buildAutotoolsEnv with dependency paths - **DONE**
- #551: feat(actions): implement setup_build_env action - **DONE**
- #552: feat(recipes): add openssl recipe using homebrew - **DONE**
- #553: feat(recipes): add ncurses recipe to validate pkg-config - **DONE**
- #554: feat(recipes): add curl recipe to validate openssl - **DONE**
- #555: feat(actions): implement cmake_build action - **DONE**
- #556: feat(recipes): add ninja recipe to validate cmake_build - **DONE**

### Code References Check
Searched for references to M19 issue numbers in:
- TODO/FIXME/HACK/XXX comments: None found ✓
- Code comments: Only in design docs (appropriate) ✓
- Test skip conditions: None found ✓

### Build Environment Artifacts
Checked for leftover development artifacts related to:
- cmake implementation
- pkg-config setup
- openssl/ncurses/curl recipes
- setup_build_env action

**Result**: No build-specific debug code or temporary files found ✓

## Test Skip Analysis

The codebase has 60+ instances of `t.Skip()` calls, but none are related to M19 issues. They fall into legitimate categories:

1. **Short mode tests** (17 instances): Skip slow tests in `-short` mode
2. **Environment-based** (15 instances): Skip when API keys/tools not available
3. **Platform-specific** (20 instances): Skip on incompatible platforms
4. **Integration tests** (8 instances): Skip unless explicitly enabled

All skip conditions are appropriate and not dead code.

## Recommendations

### Immediate Actions (Before M19 PR Merge)
1. **MUST**: Remove wip/research/complete-milestone_M19_validate_metadata.md
2. **SHOULD**: Fix or remove debug print in pip_exec.go:219

### Future Cleanup
3. **Track**: Monitor issues #643, #644, #660 for completion, then clean up TODOs
4. **Consider**: Audit all debug-style print statements and convert to structured logging

## Conclusion

The codebase is in good shape for M19 completion. The only critical finding is one leftover research artifact in wip/ that must be removed before merge. The debug print statement should be addressed but is not blocking. All TODO comments reference open future work and are appropriately documented.

No evidence of dead code related to closed M19 issues was found, indicating proper cleanup during implementation.
