# Dead Code Scanner Report: Milestone M20

**Milestone:** M20 - Dependency Provisioning: Full Integration
**Date:** 2025-12-23
**Status:** FINDINGS

## Executive Summary

Found 4 TODO comments referencing open issues (#644, #643, #660) that are **not** part of milestone M20. These TODOs are legitimate future work items, not dead code from M20. No debug code patterns, unused feature flags, or leftover test artifacts related to M20 were found.

## Detailed Findings

### 1. TODO Comments (4 instances)

All TODO comments reference open issues that are **outside** milestone M20:

#### Issue #644: Aggregate primitive action dependencies in composite actions (OPEN)
**References:** 3 locations
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew_relocate.go:21`
  ```go
  // TODO(#644): This dependency should be automatically inherited by composite actions like homebrew.
  // Currently duplicated in HomebrewAction due to dependency resolution happening before decomposition.
  ```

- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew.go:31`
  ```go
  // TODO(#644): Remove this method once composite actions automatically aggregate primitive dependencies.
  // This is a workaround because dependency resolution happens before decomposition.
  ```

- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/resolver.go:63`
  ```go
  // TODO(#644): Aggregate dependencies from primitive actions when step.Action is decomposable.
  // Currently only collects dependencies declared directly on the composite action.
  ```

**Analysis:** Issue #644 is open and not part of M20. These TODOs document known technical debt for future enhancement. **Not dead code.**

#### Issue #643: Support platform-conditional dependencies (OPEN)
**References:** 2 locations
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew_relocate.go:23`
  ```go
  // TODO(#643): Use platform-conditional dependencies to only install patchelf on Linux.
  // Currently installed on all platforms for consistency, but only used on Linux.
  ```

- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/homebrew.go:33`
  ```go
  // TODO(#643): Use platform-conditional dependencies to only install patchelf on Linux.
  // Currently installed on all platforms for consistency, but only used on Linux.
  ```

**Analysis:** Issue #643 is open and not part of M20. These TODOs document platform optimization opportunities. **Not dead code.**

#### Issue #660: Support version transformations in recipe URLs (OPEN)
**References:** 1 location
- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/testdata/recipes/sqlite-source.toml:15`
  ```toml
  # TODO(#660): SQLite uses non-standard version format (3.51.1 -> 3510100) requiring manual URL updates
  ```

**Analysis:** Issue #660 is open and not part of M20. This TODO documents a known limitation with SQLite version formatting. **Not dead code.**

### 2. Debug Code Patterns

**Search scope:** All Go files for debug print statements (`fmt.Println`, `log.Println`, `console.log` with "debug")

**Result:** No debug print statements found. ✅

### 3. Commented-Out Code Blocks

**Search scope:** Block comments and commented code in Go files

**Result:** No suspicious commented-out code blocks found. All comments are legitimate documentation. ✅

### 4. Feature Flags

**Search scope:** Feature flag patterns, conditional compilation flags

**Result:** All feature flag usage is legitimate:
- Test framework flags (`-tool`, `-tier`, `LIST_TOOLS`)
- Build flags (`CFLAGS`, `LDFLAGS`, `RUSTFLAGS`) for compilation
- Test metadata features (`patches:ordering`, `platform:linux_only`)

No abandoned or always-on feature flags detected. ✅

### 5. Test Artifacts

**Search scope:** `.only`, `.skip`, `t.Skip()` patterns

**Result:** All `t.Skip()` calls are legitimate conditional test skips:
- Integration tests with environmental prerequisites (API keys, network)
- Platform-specific test guards (`runtime.GOOS != "linux"`)
- Optional test modes (`-short`, `-tool=<name>`)

No leftover test artifacts from M20 development. ✅

## Milestone M20 Context

**M20 Issues (All CLOSED):**
- #559: feat(recipes): add git recipe to validate complete toolchain ✅
- #558: feat(recipes): add sqlite recipe to validate readline integration ✅
- #557: feat(recipes): add readline recipe using homebrew_bottle ✅

All M20 issues are closed and no TODOs reference them.

## Verification

```bash
# Confirmed all TODOs reference open issues outside M20
$ gh issue view 644 --json state,title
{"state":"OPEN","title":"feat: aggregate primitive action dependencies in composite actions"}

$ gh issue view 643 --json state,title
{"state":"OPEN","title":"feat(deps): support platform-conditional dependencies in action dependency declarations"}

$ gh issue view 660 --json state,title
{"state":"OPEN","title":"feat: support version transformations in recipe URLs"}

# Confirmed M20 issues are all closed
$ gh issue list --milestone 20 --json number,title,state --state all
[
  {"number":559,"state":"CLOSED","title":"feat(recipes): add git recipe to validate complete toolchain"},
  {"number":558,"state":"CLOSED","title":"feat(recipes): add sqlite recipe to validate readline integration"},
  {"number":557,"state":"CLOSED","title":"feat(recipes): add readline recipe using homebrew_bottle"}
]
```

## Conclusion

**Status:** FINDINGS (4 TODO comments, but NOT dead code)

All findings are legitimate technical debt markers for future work outside M20 scope. No actual dead code artifacts from M20 development were detected. The TODO comments are well-documented and reference open issues for future milestones.

**Recommendation:** No cleanup required for M20 completion. The TODO comments should remain as they document planned future enhancements.
