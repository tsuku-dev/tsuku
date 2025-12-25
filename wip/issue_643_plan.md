# Implementation Plan: Issue #643

**Issue:** feat(deps): support platform-conditional dependencies in action dependency declarations

**Type:** Enhancement

**Milestone:** Dependency Provisioning: System-Required

---

## Problem Statement

Actions currently declare dependencies unconditionally via the `Dependencies()` method. This causes platform-specific tools like `patchelf` to be installed on all platforms, even when they're only needed on specific operating systems:

- `patchelf` is only used on Linux for ELF RPATH fixup
- macOS uses `install_name_tool` instead (system-provided)

Current implementation in `homebrew_relocate.go` and `homebrew.go`:

```go
func (HomebrewRelocateAction) Dependencies() ActionDeps {
    return ActionDeps{
        InstallTime: []string{"patchelf"},
    }
}
```

This installs `patchelf` on macOS where it's never used.

## Research Findings

### Current Dependency System Architecture

1. **ActionDeps struct** (`internal/actions/action.go`):
   - Three dependency types: `InstallTime`, `Runtime`, `EvalTime`
   - Simple `[]string` slices
   - No platform awareness

2. **Dependency Resolution** (`internal/actions/resolver.go`):
   - `ResolveDependencies()` collects deps from all recipe steps
   - Calls `GetActionDeps()` for each action
   - Merges with step-level and recipe-level overrides
   - Platform info (`OS`, `Arch`) available in `ExecutionContext` but not used during dep resolution

3. **Actions with Dependencies**:
   Found 20 actions declaring dependencies via `Dependencies()` method:
   - `homebrew_relocate`: patchelf (Linux-only)
   - `homebrew`: patchelf (Linux-only)
   - `meson_build`: meson, make, zig, patchelf (patchelf is Linux-only)
   - `set_rpath`: Uses patchelf at runtime but doesn't declare it (relies on context)
   - Others: npm_install, go_install, pip_install, cargo_build, etc.

4. **Platform Detection Patterns**:
   - `runtime.GOOS` and `runtime.GOARCH` widely used
   - Platform checks exist in Execute methods (e.g., `homebrew_relocate.go` lines 169-180 for binary format detection)
   - Platform-specific code already conditional at execution time

### Actions Benefiting from Platform Conditionals

**Immediate beneficiaries:**
1. `HomebrewRelocateAction`: patchelf (Linux-only)
2. `HomebrewAction`: patchelf (Linux-only)
3. `MesonBuildAction`: patchelf (Linux-only) within broader dep list

**Future potential:**
- Any action that uses platform-specific system tools
- Cross-platform recipes that need different dependencies per OS

## Solution Options Analysis

### Option 1: Platform-Specific Fields (Recommended)

**Design:**
```go
type ActionDeps struct {
    InstallTime []string // Cross-platform deps
    Runtime     []string // Cross-platform deps
    EvalTime    []string // Cross-platform deps

    // Platform-specific install-time dependencies
    LinuxInstallTime  []string
    DarwinInstallTime []string

    // Platform-specific runtime dependencies
    LinuxRuntime      []string
    DarwinRuntime     []string
}
```

**Pros:**
- Simple, explicit, type-safe
- Easy to understand and document
- No new data structures or parsing logic
- Follows Go's simplicity principle
- Clear at a glance which deps apply to which platform

**Cons:**
- Limited to OS-level granularity (no arch-specific deps)
- Fixed set of platforms (can't add new OS without code change)
- Potential field proliferation if we add Windows, FreeBSD, etc.

**Complexity:** Low
**Flexibility:** Medium (sufficient for current needs)

### Option 2: Map-Based Platform Deps

**Design:**
```go
type PlatformDep struct {
    When string   // Platform matcher: "linux", "darwin", "linux/amd64"
    Deps []string  // Dependencies for this platform
}

type ActionDeps struct {
    InstallTime []string // Cross-platform deps
    Runtime     []string // Cross-platform deps
    EvalTime    []string // Cross-platform deps

    PlatformInstallTime []PlatformDep // Platform-specific install deps
    PlatformRuntime     []PlatformDep // Platform-specific runtime deps
}
```

**Pros:**
- More flexible - can add OS/arch combinations
- Extensible to new platforms without code changes
- Could support complex conditions later (e.g., "linux/amd64|linux/arm64")

**Cons:**
- More complex to implement and understand
- Requires parsing logic for platform matchers
- More opportunity for user error in recipes
- Harder to validate at compile time

**Complexity:** High
**Flexibility:** High (perhaps too much for current needs)

### Recommended Approach: Option 1

**Rationale:**
1. **YAGNI Principle:** We only need Linux vs Darwin distinction for patchelf
2. **Simplicity:** Matches Go's preference for explicit, simple code
3. **Type Safety:** Compiler enforces correct usage
4. **Readability:** Intent is immediately clear
5. **Sufficient:** Handles all current use cases

**When to reconsider Option 2:**
- If we need architecture-specific deps (e.g., different tools for arm64 vs amd64)
- If we add 3+ more operating systems
- If we need complex platform matching logic

## Implementation Design

### 1. Update ActionDeps Struct

**File:** `internal/actions/action.go`

```go
type ActionDeps struct {
    // Cross-platform dependencies
    InstallTime []string
    Runtime     []string
    EvalTime    []string

    // Platform-specific install-time dependencies
    // Only applied when OS matches runtime.GOOS
    LinuxInstallTime  []string
    DarwinInstallTime []string

    // Platform-specific runtime dependencies
    // Only applied when OS matches runtime.GOOS
    LinuxRuntime  []string
    DarwinRuntime []string
}
```

### 2. Update Dependency Resolution

**File:** `internal/actions/resolver.go`

Modify `ResolveDependencies()` to merge platform-specific deps based on target OS:

```go
func ResolveDependencies(r *recipe.Recipe) ResolvedDeps {
    return ResolveDependenciesForPlatform(r, runtime.GOOS)
}

// ResolveDependenciesForPlatform allows testing with different platforms
func ResolveDependenciesForPlatform(r *recipe.Recipe, targetOS string) ResolvedDeps {
    result := ResolvedDeps{
        InstallTime: make(map[string]string),
        Runtime:     make(map[string]string),
    }

    // Phase 1: Collect from steps
    for _, step := range r.Steps {
        actionDeps := GetActionDeps(step.Action)

        // Install-time: cross-platform + platform-specific
        if stepDeps := getStringSliceParam(step.Params, "dependencies"); stepDeps != nil {
            // Step override: replace
            for _, dep := range stepDeps {
                name, version := parseDependency(dep)
                result.InstallTime[name] = version
            }
        } else {
            // Action implicit: merge cross-platform + platform-specific
            for _, dep := range actionDeps.InstallTime {
                if dep != r.Metadata.Name {
                    result.InstallTime[dep] = "latest"
                }
            }

            // Add platform-specific deps
            platformDeps := getPlatformDeps(actionDeps, targetOS, true)
            for _, dep := range platformDeps {
                if dep != r.Metadata.Name {
                    result.InstallTime[dep] = "latest"
                }
            }

            // Step extend
            if extraDeps := getStringSliceParam(step.Params, "extra_dependencies"); extraDeps != nil {
                for _, dep := range extraDeps {
                    name, version := parseDependency(dep)
                    result.InstallTime[name] = version
                }
            }
        }

        // Runtime: similar logic
        // [implementation follows same pattern]
    }

    // Phase 2: Recipe-level replace
    // Phase 3: Recipe-level extend
    // [existing logic remains]

    return result
}

// getPlatformDeps returns platform-specific dependencies for the target OS
func getPlatformDeps(deps ActionDeps, targetOS string, installTime bool) []string {
    var result []string

    if installTime {
        switch targetOS {
        case "linux":
            result = deps.LinuxInstallTime
        case "darwin":
            result = deps.DarwinInstallTime
        }
    } else {
        switch targetOS {
        case "linux":
            result = deps.LinuxRuntime
        case "darwin":
            result = deps.DarwinRuntime
        }
    }

    return result
}
```

### 3. Update Actions with Platform-Specific Deps

**Files to modify:**

#### `internal/actions/homebrew_relocate.go`

```go
func (HomebrewRelocateAction) Dependencies() ActionDeps {
    return ActionDeps{
        LinuxInstallTime: []string{"patchelf"},
    }
}
```

Remove TODO comments (lines 23-24 in current code).

#### `internal/actions/homebrew.go`

```go
func (HomebrewAction) Dependencies() ActionDeps {
    return ActionDeps{
        LinuxInstallTime: []string{"patchelf"},
    }
}
```

Remove TODO comments (lines 33-34 in current code).

#### `internal/actions/meson_build.go`

```go
func (MesonBuildAction) Dependencies() ActionDeps {
    return ActionDeps{
        InstallTime: []string{"meson", "make", "zig"},
        LinuxInstallTime: []string{"patchelf"},
    }
}
```

Separates cross-platform deps from Linux-specific patchelf.

### 4. Testing Strategy

**New test file:** `internal/actions/platform_deps_test.go`

Test cases:
1. **Platform-specific install deps are applied correctly**
   - Linux: includes LinuxInstallTime
   - Darwin: includes DarwinInstallTime
   - Other OS: excludes platform-specific deps

2. **Cross-platform deps are always applied**
   - Verify InstallTime deps appear on all platforms

3. **Multiple actions with different platform deps**
   - Recipe with homebrew and meson_build
   - Verify deps are merged correctly per platform

4. **Step-level overrides work with platform deps**
   - Step override replaces both cross-platform and platform-specific
   - Step extend adds to both

5. **Recipe-level overrides work with platform deps**
   - Recipe Dependencies replaces all
   - Recipe ExtraDependencies adds to all

**Update existing tests:**
- `internal/actions/resolver_test.go`: Add platform parameter to test helpers
- Verify existing tests still pass with default platform (runtime.GOOS)

**Integration verification:**
- Build a test recipe using homebrew action
- Install on Linux: verify patchelf is installed
- Install on macOS: verify patchelf is NOT installed

### 5. Documentation Updates

**Code comments:**
- Document new fields in `ActionDeps` struct
- Explain platform matching behavior in `ResolveDependencies`
- Add examples to godoc comments

**No README changes needed:**
- This is an internal API change
- Recipe authors use existing dependency declarations
- Platform conditionals are handled transparently

## Implementation Steps

1. **Update ActionDeps struct** (`internal/actions/action.go`)
   - Add four new fields: LinuxInstallTime, DarwinInstallTime, LinuxRuntime, DarwinRuntime
   - Add documentation comments

2. **Update dependency resolution** (`internal/actions/resolver.go`)
   - Implement `ResolveDependenciesForPlatform` (testable version)
   - Update `ResolveDependencies` to call it with `runtime.GOOS`
   - Implement `getPlatformDeps` helper
   - Update install-time resolution logic
   - Update runtime resolution logic

3. **Write comprehensive tests** (`internal/actions/platform_deps_test.go`)
   - Create new test file
   - Implement all test cases from testing strategy
   - Verify both Linux and Darwin scenarios

4. **Update existing tests** (`internal/actions/resolver_test.go`)
   - Ensure existing tests pass
   - Add platform-specific test coverage

5. **Update actions with platform-specific deps**
   - Update `homebrew_relocate.go`
   - Update `homebrew.go`
   - Update `meson_build.go`
   - Remove TODO comments

6. **Run full test suite**
   - `go test ./...`
   - `go vet ./...`
   - `golangci-lint run --timeout=5m ./...`
   - `go build -o tsuku ./cmd/tsuku`

7. **Manual verification**
   - Test homebrew recipe installation on both platforms (if possible)
   - Or verify via test coverage that correct deps are resolved

## Files to Modify

1. `internal/actions/action.go` - Add fields to ActionDeps
2. `internal/actions/resolver.go` - Update resolution logic
3. `internal/actions/platform_deps_test.go` - New test file
4. `internal/actions/resolver_test.go` - Update existing tests
5. `internal/actions/homebrew_relocate.go` - Use platform-specific deps
6. `internal/actions/homebrew.go` - Use platform-specific deps
7. `internal/actions/meson_build.go` - Use platform-specific deps

**Total:** 7 files (6 modifications + 1 new)

## Risks and Mitigations

### Risk 1: Breaking Changes
**Concern:** Existing code relies on ActionDeps structure
**Mitigation:** New fields are optional; zero values maintain backward compatibility

### Risk 2: Platform Detection Edge Cases
**Concern:** Non-Linux/Darwin platforms (Windows, FreeBSD)
**Mitigation:** Platform-specific fields only apply to their OS; other platforms ignore them and use cross-platform deps

### Risk 3: Test Coverage
**Concern:** Hard to test both platforms in CI
**Mitigation:** Unit tests use `ResolveDependenciesForPlatform` with explicit OS parameter

### Risk 4: Future Extensibility
**Concern:** What if we need arch-specific deps?
**Mitigation:** Can add `Linux64InstallTime`, `LinuxArm64InstallTime` later, or migrate to Option 2 if complexity justifies it

## Success Criteria

1. **Functionality:**
   - patchelf only installed on Linux for homebrew/meson actions
   - All existing tests pass
   - New tests verify platform-specific behavior

2. **Code Quality:**
   - No golangci-lint warnings
   - Test coverage maintained
   - Clear, documented API

3. **Backward Compatibility:**
   - Existing actions without platform-specific deps continue working
   - No changes to recipe format or user-facing APIs

## Future Enhancements

1. **Architecture-specific deps:**
   - Add `Linux64InstallTime`, `LinuxArm64InstallTime` if needed
   - Or implement Option 2 (map-based) if use cases justify complexity

2. **Step-level platform overrides:**
   - Allow recipes to override platform deps at step level
   - Example: `linux_dependencies = ["custom-patchelf"]`

3. **Validation:**
   - Detect unused platform-specific deps (e.g., DarwinInstallTime when action only runs on Linux)
   - Warn if platform deps contradict execution logic

## Notes

- This is a pure enhancement; no deprecations or breaking changes
- Focus on simplicity and maintainability
- The simple field-based approach is sufficient for current needs
- Can evolve to more complex solution if requirements change
