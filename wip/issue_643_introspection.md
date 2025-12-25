# Issue #643 Introspection

## Issue Details
- **Number**: #643
- **Title**: feat(deps): support platform-conditional dependencies in action dependency declarations
- **Created**: 2025-12-20 (5 days ago)
- **Milestone**: Dependency Provisioning: System-Required
- **Status**: OPEN

## Staleness Signals
- **Issue age**: 5 days
- **Sibling issues closed since creation**: 3 (#560, #561, #562)
- **Milestone position**: Middle
- **Files modified since creation**: 2 files
  - `internal/actions/action.go`
  - `internal/actions/homebrew_relocate.go`

## Context Reviewed
- **Design doc**: DESIGN-dependency-provisioning.md (reviewed)
- **Sibling issues reviewed**: 3 (#560, #561, #562)
- **Files checked**: 10+ action files
- **Related commits**: 4 key commits examined

## Findings

### What Has Changed Since Issue Creation

1. **Issue #560 (require_system action) - COMPLETED**
   - Implemented `require_system` action in `internal/actions/require_system.go`
   - Action detects command presence, validates versions, provides installation guidance
   - Registered as primitive action in action registry
   - Full implementation matches design spec

2. **Issue #561 (docker recipe) - COMPLETED**
   - Created `internal/recipe/recipes/d/docker.toml` using `require_system` action
   - Platform-specific install guides for darwin, linux, fallback
   - Version regex parsing working
   - Recipe successfully validates docker presence

3. **Issue #562 (cuda recipe) - COMPLETED**
   - Created `internal/recipe/recipes/c/cuda.toml` using `require_system` action
   - Platform-specific guidance (Linux only, macOS unsupported message)
   - Minimum version constraint (11.0+) implemented
   - Recipe successfully validates cuda/nvcc presence

4. **File Modifications Analysis**:

   **`internal/actions/action.go`**:
   - Added `require_system` to registered actions (commit 2d0b5c8)
   - No changes to `ActionDeps` struct definition
   - ActionDeps remains unchanged: `InstallTime []string`, `Runtime []string`, `EvalTime []string`

   **`internal/actions/homebrew_relocate.go`**:
   - Updated TODO comment (commit 43e40d0) to note that dependency aggregation from #644 is working
   - **Still contains TODO(#643)**: "Use platform-conditional dependencies to only install patchelf on Linux"
   - patchelf dependency declaration remains unconditional
   - Currently installs patchelf on **all platforms** (macOS, Linux) even though only used on Linux

### Current State Assessment

**Problem Still Exists**: The original problem described in #643 is still present:
- `patchelf` is declared as an unconditional dependency in `HomebrewRelocateAction.Dependencies()`
- It gets installed on macOS even though macOS uses `install_name_tool` (system tool)
- patchelf is only needed on Linux for ELF RPATH fixup

**Other Actions with Platform-Specific Tool Usage**:
- `meson_build`: Declares `patchelf` unconditionally (line 20)
- `homebrew`: Declares `patchelf` unconditionally (line 37)
- All build actions that might use RPATH fixup have the same issue

**Current Workaround**:
- Actions install patchelf everywhere "for consistency"
- Code detects platform at runtime and only uses patchelf on Linux (lines 188-208 in homebrew_relocate.go)
- Wastes disk space and installation time on macOS

### No Blockers or Design Changes

The design document (DESIGN-dependency-provisioning.md) still includes platform-conditional dependencies:
- Section discusses the need for platform-conditional deps
- Presents two options:
  1. **Option 1**: More complex `PlatformDeps map[string]PlatformDep` structure
  2. **Option 2**: Simpler OS-specific fields like `LinuxInstallTime []string`, `DarwinInstallTime []string`

The issue description proposes both approaches and asks which to implement.

### Issue Validity

**The issue spec remains FULLY VALID**:
1. Problem statement is accurate (patchelf installed on all platforms)
2. Proposed solutions are still relevant
3. No sibling work has addressed platform-conditional dependencies
4. Design doc supports this feature
5. Multiple actions would benefit (homebrew_relocate, meson_build, homebrew, set_rpath)

### Implementation Readiness

**Ready to implement**:
- Clear problem statement
- Two well-defined solution options
- Multiple use cases identified
- No dependencies on other work
- Test infrastructure exists (validated by sibling issues)

## Recommendation

**PROCEED AS SPECIFIED**

### Rationale

1. **Issue is current and valid**: The problem still exists exactly as described
2. **No design drift**: Sibling work (#560, #561, #562) focused on system dependencies, not platform-conditional deps
3. **Clear specification**: Issue presents two concrete solution options
4. **Multiple beneficiaries**: At least 4 actions would immediately benefit
5. **User value**: Reduces unnecessary installations on macOS (patchelf ~2MB, only needed on Linux)

### Suggested Action

1. Review the two proposed solution options with the user/team:
   - **Option A** (complex): `PlatformDeps map[string]PlatformDep` - more flexible, supports complex conditions
   - **Option B** (simple): `LinuxInstallTime`, `DarwinInstallTime` fields - simpler, covers 95% of cases

2. Implement chosen solution in `internal/actions/action.go`

3. Update affected actions to use platform-conditional syntax:
   - `homebrew_relocate.go`
   - `homebrew.go`
   - `meson_build.go`
   - Any other actions using platform-specific tools

4. Add tests for platform-conditional dependency resolution

5. Verify patchelf only installed on Linux in CI matrix

### No Blocking Concerns

- Sibling work complements this issue (system deps) rather than conflicts
- ActionDeps struct is extensible (can add new fields)
- Backward compatibility maintained (unconditional deps still work)
- Design doc explicitly covers this scenario
