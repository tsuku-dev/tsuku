# Platform Tuple Support for install_guide and when Clauses

## Status

Proposed

## Context and Problem Statement

PR #685 implemented platform-aware recipes with OS and architecture constraints using the `supported_os` and `supported_arch` arrays. While this provides recipe-level platform filtering, it operates at OS-level granularity only for step-level features like `install_guide` and the `when` clause.

The `install_guide` field in the `require_system` action currently maps OS strings (`darwin`, `linux`) to installation instructions. However, some system dependencies have different installation methods or package names based on both OS _and_ architecture. For example:

- **darwin/arm64** (Apple Silicon): Homebrew installed at `/opt/homebrew`
- **darwin/amd64** (Intel Mac): Homebrew installed at `/usr/local`

These architectural differences can require different install commands or package managers, but the current implementation only allows OS-level keys like `darwin` or `linux`.

Similarly, the step-level `when` clause (defined in `Step.When` but currently undocumented and unused) only supports `os` and `arch` arrays for independent filtering, not platform tuples that combine both dimensions.

### Current Limitation

Given this recipe fragment:

```toml
[[steps]]
action = "require_system"
command = "gcc"
install_guide = { darwin = "brew install gcc", linux = "apt install gcc" }
when = { os = ["darwin", "linux"] }
```

This assumes all darwin architectures use the same Homebrew location and command, which is not true for Apple Silicon vs Intel Macs.

### Why This Matters Now

With platform-aware recipe support (#228) merged, recipes can now declare precise platform constraints. However, the step-level execution features have not been updated to match this precision. This creates an architectural inconsistency:

- **Recipe-level**: Can express "linux/arm64 only" via `supported_os` + `supported_arch`
- **Step-level**: Can only express "darwin" or "linux" in `install_guide` and `when`

This limitation prevents recipes from providing architecture-specific installation guidance or conditionally executing steps based on exact platform tuples.

### Scope

**In scope:**
- Support platform tuple format (`os/arch`) as keys in `install_guide` maps
- Fallback behavior: exact tuple match first, then OS-only match, then fallback key
- Validation: ensure `install_guide` has either complete tuple coverage or OS-level fallbacks for all supported platforms

**Out of scope:**
- Step-level `when` clause support (currently undocumented and unused; should be addressed in separate design)
- Changes to recipe-level platform constraints (`supported_os`, `supported_arch` work as-is)
- New action types or execution semantics
- Website or CLI display changes (already handled by existing platform support)

**Current Usage:**
Only 2 recipes currently use `install_guide` (docker.toml and cuda.toml), both with OS-level keys only. This is a forward-looking enhancement to support architecture-specific system dependencies as they become more common.

## Decision Drivers

- **Consistency**: Step-level features should match the precision of recipe-level constraints
- **Backwards compatibility**: Existing OS-only keys in `install_guide` must continue to work
- **Simplicity**: Fallback logic should be intuitive (try exact match, then OS match, then fallback)
- **Validation**: Recipe authors should get clear errors if coverage is incomplete
- **Implementation scope**: Reuse existing platform computation logic from PR #685

## Implementation Context

### Existing Platform Support (PR #685)

PR #685 introduced comprehensive platform-aware recipe support:

**Platform Computation:**
- `Recipe.GetSupportedPlatforms()` returns platform tuples (`[]string` like `["linux/amd64", "darwin/arm64"]`)
- `Recipe.SupportsPlatform(os, arch)` checks if a specific platform tuple is supported
- Uses complementary hybrid model: `(supported_os × supported_arch) - unsupported_platforms`

**Key Files:**
- `internal/recipe/platform.go:194-212` - `GetSupportedPlatforms()` returns final platform tuples
- `internal/recipe/platform.go:86-111` - `SupportsPlatform(os, arch)` validates platform support
- `internal/recipe/platform.go:267-335` - `ValidateStepsAgainstPlatforms()` validates step mappings

**Current Validation (line 318-331):**
```go
// Check install_guide coverage for require_system steps
// TODO: Update this validation when issue #686 is resolved (support platform tuples in install_guide)
if step.Action == "require_system" {
    if installGuide, ok := step.Params["install_guide"].(map[string]interface{}); ok {
        for os := range supportedOS {
            if _, hasGuide := installGuide[os]; !hasGuide {
                errors = append(errors, &StepValidationError{
                    StepIndex: i,
                    Message:   fmt.Sprintf("install_guide missing entry for supported OS '%s' (see issue #686 for platform tuple support)", os),
                })
            }
        }
    }
}
```

This validation currently only checks OS-level coverage, with an explicit TODO referencing this issue.

### Existing install_guide Pattern

The `require_system` action (internal/actions/require_system.go:182-199) already has fallback logic:

```go
func getPlatformGuide(installGuide map[string]string, platform string) string {
    if installGuide == nil {
        return ""
    }

    // Try platform-specific guide (e.g., "darwin", "linux")
    if guide, ok := installGuide[platform]; ok {
        return guide
    }

    // Try fallback
    if guide, ok := installGuide["fallback"]; ok {
        return guide
    }

    return ""
}
```

This function is called with `runtime.GOOS`, so it currently only supports OS-level keys. The fallback pattern is already established.

### Step When Clause (Undocumented)

The `Step.When` field exists in types.go but is not currently used:
- Defined in recipe types but no runtime enforcement
- Would need executor changes to support conditional step execution
- Out of scope for install_guide support (can be addressed separately)

### Patterns to Follow

**Validation consistency:**
- Use `Recipe.GetSupportedPlatforms()` to get final platform tuples
- Extract unique OS/arch sets for validation (platform.go:284-290 pattern)
- Return `[]error` with descriptive messages for validation failures

**Fallback hierarchy:**
- Match exact platform tuple first (`darwin/arm64`)
- Fall back to OS-only key (`darwin`)
- Fall back to `fallback` key if present
- This mirrors the existing `getPlatformGuide` pattern

**Test coverage:**
- Add tests to `internal/recipe/platform_test.go` (TestValidateStepsAgainstPlatforms pattern)
- Add tests to `internal/actions/require_system_test.go` (TestGetPlatformGuide pattern)

## Considered Options

### Option 1: Extend install_guide with Platform Tuple Keys

Allow `install_guide` to use platform tuple keys (`os/arch` format) alongside existing OS-only keys, with hierarchical fallback.

**Example:**
```toml
[[steps]]
action = "require_system"
command = "brew"
install_guide = {
  "darwin/arm64" = "/opt/homebrew/bin/brew install gcc",
  "darwin/amd64" = "/usr/local/bin/brew install gcc",
  "darwin" = "brew install gcc",  # fallback for any other darwin arch
  "linux" = "apt install gcc"
}
```

**Lookup logic:**
1. Try exact platform tuple (`runtime.GOOS/runtime.GOARCH`)
2. Fall back to OS-only key (`runtime.GOOS`)
3. Fall back to `fallback` key if present

**Implementation:**
- Update `getPlatformGuide()` in `require_system.go` to accept both OS and arch
- Check tuple key first, then OS key, then fallback
- Update validation in `ValidateStepsAgainstPlatforms()` to allow tuple keys

**Pros:**
- **Maximum flexibility**: Recipe authors can provide architecture-specific guidance where needed
- **Backwards compatible**: Existing OS-only keys continue to work
- **Intuitive fallback**: Hierarchical lookup is easy to understand
- **Minimal API surface**: No new fields or structures, just enhanced key format

**Cons:**
- **Validation complexity**: Must validate that either OS-level or tuple-level coverage is complete
- **Mixed granularity**: Same map can have both OS and tuple keys, which may be confusing
- **Partial coverage ambiguity**: What if only `darwin/arm64` is specified but not `darwin/amd64`? Validation must catch this or fall back gracefully.

### Option 2: Separate arch_install_guide Field

Introduce a new `arch_install_guide` field that uses `os/arch` tuple keys only, keeping `install_guide` for OS-only keys.

**Example:**
```toml
[[steps]]
action = "require_system"
command = "brew"
install_guide = {
  "darwin" = "brew install gcc",  # Generic darwin fallback
  "linux" = "apt install gcc"
}
arch_install_guide = {
  "darwin/arm64" = "/opt/homebrew/bin/brew install gcc",
  "darwin/amd64" = "/usr/local/bin/brew install gcc"
}
```

**Lookup logic:**
1. Try `arch_install_guide[os/arch]`
2. Fall back to `install_guide[os]`
3. Fall back to `install_guide[fallback]`

**Pros:**
- **Clear separation**: No mixed granularity confusion - each field has one purpose
- **Explicit opt-in**: Recipe authors only use `arch_install_guide` when architecture differences matter
- **Consistent with existing patterns**: Mirrors the `os_mapping` and `arch_mapping` separation used in download actions
- **Simpler validation**: Each field validates independently against one granularity level
- **Clear override semantics**: `arch_install_guide` always takes precedence when present

**Cons:**
- **Override behavior ambiguity**: If `arch_install_guide` has darwin/arm64 but `install_guide` has darwin, which wins? (Answer: arch_install_guide, but this must be documented)
- **Potential for conflicting guidance**: Recipe authors could accidentally provide incompatible instructions across both fields
- **Additional schema field**: Adds another optional field to the recipe format
- **Learning curve**: Recipe authors must understand the precedence rules between the two fields

### Option 3: Status Quo with Documentation

Document the current limitation and recommend recipe authors use separate recipes for architecture-specific cases.

**Example:**
Create `homebrew-arm64.toml` and `homebrew-amd64.toml` recipes with different `install_guide` values and platform constraints.

**Pros:**
- **Zero implementation cost**: No code changes required
- **Maximum clarity**: Each recipe has one clear purpose
- **Leverages existing platform constraints**: Uses `supported_arch` effectively

**Cons:**
- **Recipe proliferation**: More recipe files for what is essentially the same dependency
- **Maintenance burden**: Changes to one recipe must be duplicated to others
- **User confusion**: Multiple recipes for the same tool with overlapping names
- **Not scalable**: Doesn't solve the problem, just works around it

## Alternatives Considered and Rejected

### Template-based Variable Substitution

Allow template variables in `install_guide` values (e.g., `{HOMEBREW_PREFIX}/bin/brew install gcc`).

**Rejected because:**
- Requires defining and documenting variable naming conventions
- Adds templating engine complexity to recipe parsing
- Doesn't solve the fundamental problem (still OS-level keys)
- Overkill for the specific Homebrew path use case

### Tuple-only Keys (No Mixed Granularity)

Require all `install_guide` keys to be platform tuples, deprecating OS-only keys.

**Rejected because:**
- Breaks backwards compatibility (violates decision driver)
- Forces unnecessary verbosity when OS-level guidance is sufficient
- Migration burden for existing recipes

### Callable Installation Scripts

Allow `install_guide` to reference executable scripts that detect platform dynamically.

**Rejected because:**
- Out of scope (requires action execution model changes)
- Security implications of arbitrary script execution
- Overkill for static installation instructions
- Inconsistent with tsuku's declarative recipe philosophy

## Decision Outcome

**Chosen option: Option 1 - Extend install_guide with Platform Tuple Keys**

This option provides the best balance of flexibility, backwards compatibility, and implementation simplicity while directly addressing the architectural inconsistency between recipe-level and step-level platform precision.

### Rationale

This option was chosen because:

- **Consistency (primary driver)**: Allows step-level `install_guide` to match the precision of recipe-level platform constraints. A recipe can declare `supported_os = ["darwin"]` with `unsupported_platforms = ["darwin/amd64"]` and provide matching guidance for `darwin/arm64` only.

- **Backwards compatibility**: Existing recipes with OS-only keys (`darwin`, `linux`) continue to work without modification. The two recipes currently using `install_guide` (docker.toml, cuda.toml) require no changes.

- **Simplicity**: Hierarchical fallback (tuple → OS → fallback) is intuitive and mirrors common configuration patterns. Recipe authors can incrementally add architecture-specific guidance where needed.

- **Minimal API surface**: No new fields or structures. The enhancement is purely in key format and lookup logic. The existing `getPlatformGuide()` function signature changes from accepting one platform string to accepting OS and architecture separately, but the fallback pattern remains the same.

- **Implementation scope**: Reuses existing platform computation from `Recipe.GetSupportedPlatforms()`. Validation can leverage the same platform tuple extraction pattern used elsewhere.

### Rejected Alternatives

**Option 2 (Separate arch_install_guide field)** was rejected because:
- Adds unnecessary API complexity (new field) when existing field can be enhanced
- Introduces precedence rules that must be learned and documented
- Creates potential for conflicting guidance across two fields
- Does not align with the "extend existing patterns" principle

**Option 3 (Status quo with documentation)** was rejected because:
- Does not solve the architectural inconsistency
- Forces recipe proliferation for a capability that should be intrinsic
- Maintenance burden scales poorly as architecture-specific dependencies become more common

### Trade-offs Accepted

By choosing this option, we accept:

**Mixed granularity in a single map**: The same `install_guide` can contain both OS-only keys (`darwin`) and platform tuple keys (`darwin/arm64`). This could initially be confusing to recipe authors.

**Validation complexity**: The validation logic must handle multiple valid coverage patterns:
- Complete OS-level coverage (current behavior)
- Complete tuple-level coverage for all supported platforms
- Mixed coverage where some OSes have tuple-level entries and others have OS-level entries

**Partial coverage edge cases**: If a recipe specifies `darwin/arm64` but not `darwin/amd64` and has no `darwin` fallback, validation must catch this as an error.

These trade-offs are acceptable because:
- Mixed granularity only appears when needed (most recipes will use OS-only keys)
- Validation complexity is confined to one function (`ValidateStepsAgainstPlatforms`) which already exists
- Clear validation error messages will guide recipe authors to correct patterns
- The flexibility gained (architecture-specific guidance) justifies the modest increase in complexity

## Solution Architecture

### Overview

The solution extends the existing `install_guide` field to accept platform tuple keys (`os/arch` format) alongside OS-only keys. Lookup uses hierarchical fallback: exact tuple match → OS match → fallback key. Validation ensures coverage for all supported platforms, accepting either OS-level or tuple-level entries.

### Components

**Modified Components:**

1. **`internal/actions/require_system.go:getPlatformGuide()`**
   - Current: `func getPlatformGuide(installGuide map[string]string, platform string) string`
   - New: `func getPlatformGuide(installGuide map[string]string, os, arch string) string`
   - Implements hierarchical fallback logic

2. **`internal/recipe/platform.go:ValidateStepsAgainstPlatforms()`**
   - Current: Validates OS-level coverage only
   - New: Validates coverage using tuple-aware algorithm (see below)

3. **`internal/actions/require_system.go:Execute()`**
   - Current: Calls `getPlatformGuide(installGuide, runtime.GOOS)`
   - New: Calls `getPlatformGuide(installGuide, runtime.GOOS, runtime.GOARCH)`

**No New Components Required**

### Key Interfaces

**getPlatformGuide Lookup Algorithm:**

```go
func getPlatformGuide(installGuide map[string]string, os, arch string) string {
    if installGuide == nil {
        return ""
    }

    // Try exact platform tuple first
    tuple := fmt.Sprintf("%s/%s", os, arch)
    if guide, ok := installGuide[tuple]; ok {
        return guide
    }

    // Fall back to OS-only key
    if guide, ok := installGuide[os]; ok {
        return guide
    }

    // Fall back to generic fallback key
    if guide, ok := installGuide["fallback"]; ok {
        return guide
    }

    return ""
}
```

**ValidateStepsAgainstPlatforms Coverage Algorithm:**

```
For each require_system step with install_guide:
  1. Get recipe's supported platforms (from Recipe.GetSupportedPlatforms())
  2. Extract unique OS/arch sets from platform tuples

  3. For each supported platform tuple (e.g., "darwin/arm64"):
     - Check if install_guide has exact tuple key
     - If not, check if install_guide has OS key for this tuple's OS
     - If not, check if install_guide has "fallback" key
     - If none found, report validation error for this platform

  4. For each install_guide key:
     - If key contains "/", validate it's a platform tuple in format "os/arch"
     - If tuple key, verify it exists in supported platforms
     - If OS-only key (no "/"), verify OS exists in supported OS set
```

**Validation Error Cases:**

- Tuple key not in supported platforms: `"install_guide contains 'linux/riscv64' which is not in the recipe's supported platforms"`
- Missing coverage: `"install_guide missing entry for supported platform 'darwin/amd64' (no tuple key, no OS fallback, no generic fallback)"`
- Invalid tuple format: `"install_guide key 'darwin/' is invalid (must be 'os/arch' format)"`

### Data Flow

**Runtime (Normal Execution):**
```
1. require_system.Execute() called
2. Reads runtime.GOOS and runtime.GOARCH
3. Calls getPlatformGuide(installGuide, os, arch)
4. getPlatformGuide checks: tuple key → OS key → fallback key
5. Returns guidance string or "" if no match
6. Error message includes guidance if command not found
```

**Validation (Recipe Load Time):**
```
1. Recipe.ValidateStepsAgainstPlatforms() called
2. Gets supported platform tuples from Recipe.GetSupportedPlatforms()
3. For each require_system step:
   a. Extracts install_guide map
   b. Iterates over supported platforms
   c. For each platform, verifies coverage (tuple OR OS OR fallback)
   d. Collects validation errors for uncovered platforms
4. For each install_guide key:
   a. If tuple format, validates against supported platforms
   b. If OS format, validates against supported OS set
5. Returns []error with specific messages
```

## Implementation Approach

### Prerequisites

Before implementation:
1. **Verify TOML compatibility**: Confirm that TOML parsers correctly handle keys with slashes (e.g., `"darwin/arm64" = "..."`). Test with BurntSushi/toml library used by tsuku.
2. **Review test infrastructure**: Ensure existing test helpers can construct recipes with tuple keys.

### Implementation Phases

**Phase 1: Update getPlatformGuide() Function**
- Modify signature to accept `os` and `arch` separately
- Implement hierarchical fallback logic (tuple → OS → fallback)
- Update all call sites in `require_system.go:Execute()` to pass `runtime.GOOS, runtime.GOARCH`
- Add unit tests in `require_system_test.go:TestGetPlatformGuide` for:
  - Exact tuple match
  - OS fallback
  - Generic fallback
  - Mixed keys (tuple + OS)
  - No match

**Phase 2: Update Validation Logic**
- Modify `platform.go:ValidateStepsAgainstPlatforms()` to implement tuple-aware validation
- Add helper to check if key is tuple format (contains `/`)
- Iterate over supported platforms and validate coverage
- Validate tuple keys are in supported platforms
- Validate OS keys are in supported OS set
- Add unit tests in `platform_test.go:TestValidateStepsAgainstPlatforms` for:
  - Complete tuple coverage
  - Complete OS coverage
  - Mixed tuple and OS coverage
  - Missing coverage error cases
  - Invalid tuple format error cases
  - Tuple key not in supported platforms

**Phase 3: Update Documentation**
- Update recipe format documentation to show tuple key examples
- Update require_system action documentation
- Add migration guide showing when to use tuple keys vs OS keys
- Document validation error messages

**Phase 4: Integration Testing**
- Create test recipe with tuple keys
- Validate strict mode catches edge cases
- Test on actual darwin/arm64 and darwin/amd64 platforms
- Verify backwards compatibility with existing docker.toml and cuda.toml recipes

### Dependencies

- Phase 2 depends on Phase 1 (validation needs to understand fallback logic)
- Phase 3 can proceed in parallel with Phase 1/2
- Phase 4 depends on Phase 1 and 2 completion

## Consequences

### Positive

- **Architectural consistency**: Step-level install_guide now matches recipe-level platform precision
- **Backwards compatible**: Zero breaking changes for existing recipes
- **Incremental adoption**: Recipe authors can add tuple keys only where needed
- **Clear fallback semantics**: Hierarchical lookup is intuitive and predictable
- **Reuses existing infrastructure**: Leverages `Recipe.GetSupportedPlatforms()` without modification

### Negative

- **Validation complexity**: Validation logic must handle multiple valid coverage patterns
- **Potential confusion**: Mixed OS and tuple keys in same map could initially confuse recipe authors
- **Error message verbosity**: Validation errors for missing coverage must explain fallback logic
- **Testing surface**: More edge cases to cover in unit and integration tests

### Mitigations

- **Clear validation errors**: Error messages explain what's missing and suggest fixes
  - Example: `"install_guide missing entry for supported platform 'darwin/amd64' (no tuple key 'darwin/amd64', no OS fallback 'darwin', no generic 'fallback')"`
- **Documentation and examples**: Recipe format docs show common patterns (OS-only, tuple-only, mixed)
- **Strict mode catches edge cases**: Running `tsuku validate` in strict mode catches incomplete coverage before runtime
- **Test coverage**: Comprehensive unit tests document expected behavior for all patterns

## Security Considerations

### Download Verification

**Not applicable** - This feature does not modify how tsuku downloads or verifies external artifacts. The `install_guide` field provides human-readable installation instructions for system dependencies that tsuku cannot provision (e.g., Docker, CUDA). Platform tuple support only affects which instruction text is shown to users based on their platform. No binaries are downloaded or executed by this feature.

### Execution Isolation

**Not applicable** - This feature does not execute code or modify tsuku's execution model. The `install_guide` field contains static text displayed to users when a system dependency is missing. The only code execution is the existing validation logic at recipe load time, which operates on the recipe TOML structure in memory. No additional permissions or isolation changes are required.

### Supply Chain Risks

**Input validation for recipe authors**: The platform tuple format (`os/arch`) could potentially be exploited if TOML parsing is vulnerable to crafted key names. However:

**Mitigations:**
- TOML key validation is handled by the BurntSushi/toml library, which has been extensively tested
- Slash character in keys is valid TOML syntax (must be quoted: `"darwin/arm64" = "..."`)
- Validation logic verifies tuple format matches `os/arch` pattern (rejects malformed keys like `darwin/`, `/amd64`, `darwin/amd64/extra`)
- Tuple OS and arch components are validated against known supported values from `TsukuSupportedOS()` and `TsukuSupportedArch()`

**Residual risk:** If an attacker could modify a recipe file, they could provide misleading installation instructions (e.g., directing users to malicious package sources). This is not new - the existing `install_guide` already has this property. Recipe trust is handled at a higher level (recipe provenance, repository access controls).

### User Data Exposure

**Not applicable** - This feature does not access, transmit, or store user data. The `install_guide` text is static recipe content defined by recipe authors. No user-specific information is involved in lookup or validation. The only runtime operation is string comparison against platform keys, which uses publicly known OS/arch values from `runtime` package.

### Security Summary

This feature has minimal security impact because it operates entirely on static recipe data. The primary security consideration is ensuring tuple key validation cannot be exploited through crafted TOML keys, which is mitigated through TOML library parsing and explicit format validation.

**No new security vectors introduced** beyond what existing `install_guide` already provides (recipe authors can write arbitrary installation instructions).
