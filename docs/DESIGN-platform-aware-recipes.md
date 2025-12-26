# Platform-Aware Recipe Support

**Status:** Proposed

## Context and Problem Statement

Tsuku currently lacks a standard mechanism for recipes to declare platform compatibility, leading to poor user experience when attempting to install tools on unsupported platforms.

When a user tries to install a recipe that doesn't support their operating system or architecture, they encounter cryptic errors rather than clear, actionable messages. For example:

- **btop** on macOS fails with a 404 error because the upstream project doesn't publish darwin binaries
- **hello-nix** on macOS fails because nix-portable (the underlying action) only supports Linux

These failures happen late in the installation process after downloading dependencies and attempting installation. Users receive error messages like "404 Not Found" or "command failed" with no upfront indication that the tool isn't available for their platform.

The current codebase has partial infrastructure for conditional execution:
- `Step.When` field in `internal/recipe/types.go:175` allows steps to specify `os` and `arch` conditions
- `executor.shouldExecute()` in `internal/executor/executor.go:115` checks these conditions at runtime

However, this step-level mechanism is:
- Undocumented and unused by existing recipes
- Not exposed to consumers (CLI, website, test matrix)
- Insufficient for recipe-level platform constraints (a recipe that doesn't work on macOS needs all steps skipped, not just some)

Without recipe-level platform metadata, there's no way to:
- **Fail fast**: Detect incompatibility before attempting installation
- **Provide clear UX**: Show users which platforms are supported
- **Test efficiently**: Skip unsupported platform combinations in CI
- **Inform discovery**: Display platform badges on the website

### Scope

**In scope:**
- Recipe-level platform constraint declaration
- CLI enforcement with clear error messages before installation
- Integration with `tsuku info` to display platform support
- Test matrix filtering to skip unsupported combinations
- Relationship between recipe-level constraints and existing step-level `when` conditions

**Out of scope:**
- Cross-compilation or providing alternative binaries for unsupported platforms
- Runtime platform detection for conditional feature enablement
- Automatic platform fallback mechanisms (e.g., building from source if binary unavailable)
- Website integration (deferred to separate issue after CLI support is established)

## Decision Drivers

### Priority 0 (Must Have)
- **User experience**: Users should know upfront whether a tool supports their platform, not after installation fails
- **Backwards compatibility**: Existing recipes without platform constraints must continue working (missing metadata = supports all platforms)

### Priority 1 (Should Have)
- **Ecosystem integration**: Platform metadata must be consumable by CLI commands (`tsuku info`), test matrix, and eventually the website
- **Implementation cost**: Minimize complexity to ship quickly and reduce maintenance burden
- **Recipe simplicity**: Recipe authors shouldn't need to manually add `when` clauses to every step when the entire recipe is platform-specific

### Priority 2 (Nice to Have)
- **Existing infrastructure**: Leverage the existing `Step.When` mechanism where appropriate rather than duplicating logic
- **Clear semantics**: The distinction between recipe-level platform support and step-level conditional execution must be well-defined
- **Extensibility**: Design should accommodate future platform dimensions (libc, ABI compatibility) without breaking changes

## Implementation Context

### Existing Patterns in Tsuku

The codebase has several mechanisms for platform-specific behavior:

**1. Step-level conditional execution** (`internal/recipe/types.go:175`, `executor.go:115`)
- Steps have a `When` map with `os` and `arch` keys
- Executor checks conditions before running each step
- Used sparingly in existing recipes (e.g., `gcc-libs.toml`, `nodejs.toml`)

**2. Action-level OS/arch mapping** (`composites.go`, `download.go`)
- Actions accept `os_mapping` and `arch_mapping` parameters
- Transform Go's `runtime.GOOS/GOARCH` to upstream naming conventions
- Example: `nodejs.toml` maps `amd64` → `x64` for download URLs

**3. Platform-specific dependencies** (`actions/action.go:42-64`)
- `ActionDeps` supports `LinuxInstallTime`, `DarwinInstallTime`, etc.
- Actions declare dependencies only needed on certain platforms
- Automatically filtered during dependency resolution

**4. Preflight validation** (`actions/preflight.go`)
- Actions implement `Preflight` interface for static validation
- Warnings about unused `os_mapping`/`arch_mapping` in URLs
- Recipe validator runs preflights during CI

**5. System requirements** (`actions/require_system.go`)
- `require_system` action validates command availability
- Provides platform-specific installation guidance via `install_guide` map
- Error messages include platform-aware instructions

### How Other Package Managers Handle Platform Constraints

| Aspect | Homebrew | asdf | Nix |
|--------|----------|------|-----|
| **Schema** | Ruby DSL blocks: `on_macos`, `on_linux`, `on_arm` | Shell scripts with `uname` detection (plugin-defined) | TOML-like: `meta.platforms` list |
| **Constraint Level** | Dependency/resource level (fine-grained) | Plugin script level (plugin handles internally) | Package level (entire derivation) |
| **Platform Values** | Implicit via blocks; version-aware (`on_monterey`) | Freeform strings (plugin-defined) | Standardized: `x86_64-linux`, `aarch64-darwin` |
| **Error Messaging** | Implicit (missing dependency → build fails) | Plugin-driven (inconsistent) | Explicit: "package not supported on 'system-X'" |
| **Override** | None (use separate formulae) | Plugin exit code | `NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1` |

**Key takeaways:**
- Declarative constraints (Homebrew, Nix) are easier to query and display than procedural (asdf)
- Nix's explicit error messages with platform lists provide best UX
- Homebrew's fine-grained control allows platform-specific dependencies; Nix applies constraints at package level
- All systems distinguish between build-time and runtime platform requirements

### Applicable Specifications

**TOML Specification**
- Tsuku recipes use TOML format
- Platform constraints should follow TOML table/array conventions
- Reference: [TOML v1.0.0 spec](https://toml.io/en/v1.0.0)

**Go Platform Constants**
- Platform values should align with Go's `runtime.GOOS` and `runtime.GOARCH`
- Standard values: `linux`, `darwin`, `windows` (OS); `amd64`, `arm64`, `386`, `arm` (arch)
- Reference: [Go runtime package](https://pkg.go.dev/runtime)

### Research Summary

**Upstream constraints:**
- None (this is standalone tactical work, not implementing a strategic design)

**Patterns to follow:**
- Use TOML tables for structured metadata (consistent with recipe schema)
- Align platform values with Go runtime constants (already used throughout codebase)
- Implement preflight validation for early error detection (existing pattern)
- Provide platform-specific guidance in error messages (`install_guide` pattern)

**Specifications to comply with:**
- TOML v1.0.0 for recipe schema extensions
- Go runtime platform constants for portability

**Implementation approach:**
- Add `[metadata.platforms]` table to recipe schema
- Check constraints in CLI before executor runs (fail fast)
- Expose platform info via `tsuku info` command
- Maintain backwards compatibility (missing platforms = all platforms supported)

## Considered Options

This design addresses multiple independent questions about platform support. Options are grouped by the decision they address.

### Decision 1: Schema Design - How should recipes declare platform constraints?

#### Option 1A: Allowlist via `supported` Fields

Add `supported_os` and `supported_arch` arrays to `[metadata]`:

```toml
[metadata]
name = "hello-nix"
supported_os = ["linux"]  # Only Linux
```

If fields are missing, recipe supports all platforms (backwards compatible).

**Pros:**
- Simple mental model: "this recipe works on these platforms"
- Easy to query: check if user's platform is in the list
- Backwards compatible: missing fields = universal support
- Clear semantics for CI: skip tests when platform not in list

**Cons:**
- Requires explicit listing even when only one platform is excluded (e.g., if tool works on linux/darwin but not windows, must list both)
- Doesn't distinguish between "never tested" and "known to fail"
- No way to express "all except X" compactly
- Allowlists can become stale when new platforms are added (e.g., a new OS or arch might work but isn't in the list)

#### Option 1B: Denylist via `unsupported` Fields

Add `unsupported_os` and `unsupported_arch` arrays to `[metadata]`:

```toml
[metadata]
name = "some-tool"
unsupported_os = ["windows"]  # Works everywhere except Windows
```

Missing fields = no restrictions (backwards compatible).

**Pros:**
- Compact for "all except" cases
- Backwards compatible: missing fields = universal support
- Natural for tools that work widely with known exceptions
- Most recipes won't need the field (universal support is the common case)

**Cons:**
- Inverted logic can be confusing ("not in unsupported list" = supported)
- Doesn't help with "only works on X" cases (e.g., hello-nix Linux-only)
- Harder to generate "supported platforms" list for display

#### Option 1C: Combined Allowlist/Denylist

Support both `supported_*` and `unsupported_*` with validation rules:

```toml
# Allowlist form
[metadata]
supported_os = ["linux", "darwin"]

# Denylist form (mutually exclusive)
[metadata]
unsupported_os = ["windows"]

# Error: cannot specify both
```

Recipe validator ensures only one form is used.

**Pros:**
- Flexibility: recipe author chooses natural expression
- Supports both "only on X" and "all except Y" use cases
- Backwards compatible: missing fields = universal support

**Cons:**
- Two ways to express same concept increases complexity
- Validation logic needed to prevent conflicts
- Consumers (CLI, website) must handle both forms
- Documentation burden: explain when to use which form

### Decision 2: Granularity - What level of platform constraints?

#### Option 2A: OS and Architecture Separate

Separate `os` and `arch` constraints (following existing `Step.When` pattern):

```toml
[metadata]
supported_os = ["linux", "darwin"]
supported_arch = ["amd64", "arm64"]
```

Platform is supported if `(os in supported_os) AND (arch in supported_arch)`.

**Pros:**
- Follows existing `Step.When` pattern for consistency
- Natural for constraints that apply to OS or arch independently
- Matches how actions currently handle platform mappings
- Easy to extend (could add `supported_libc`, etc.)

**Cons:**
- Can't express "darwin/amd64 supported but not darwin/arm64" (OS-arch specific combinations)
- Cartesian product may over-specify (linux+darwin × amd64+arm64 = 4 platforms when only 2 are tested)
- Verbose for single-platform recipes (must specify both OS and arch)
- No wildcard support (can't express "linux/*, darwin/amd64" compactly)

#### Option 2B: Combined Platform Tuples

Use Go-style platform tuples:

```toml
[metadata]
supported_platforms = ["linux/amd64", "linux/arm64", "darwin/amd64"]
```

Platform tuple format: `{os}/{arch}`.

**Pros:**
- Precise: explicitly lists supported combinations
- Matches Go's `GOOS/GOARCH` convention (familiar to Go developers)
- No ambiguity about unsupported combinations
- Easy to generate from `go tool dist list` for Go-based tools

**Cons:**
- Verbose: must list all combinations even if arch doesn't matter
- Harder to query "is this OS supported?" without parsing tuples
- Doesn't align with existing `Step.When` pattern (os/arch separate)
- Less human-friendly for non-Go users

#### Option 2C: Hybrid - OS/Arch Separate + Optional Platform Tuples

Separate OS/arch fields with optional override via platform tuples:

```toml
# Case 1: OS constraint only
[metadata]
supported_os = ["linux"]  # All Linux archs

# Case 2: Specific combinations
[metadata]
supported_platforms = ["linux/amd64", "darwin/amd64"]
```

Validator ensures only one form is used.

**Pros:**
- Flexibility: simple cases use os/arch, complex cases use tuples
- Handles both "OS doesn't matter, arch doesn't matter" and "specific combinations only"
- Can evolve over time (start with os/arch, add tuples when needed)

**Cons:**
- Two schema patterns increases complexity
- Validation logic needed to prevent conflicts
- Documentation burden: explain when to use which form
- Risk of inconsistent usage across recipes

### Decision 3: Enforcement - When should platform constraints be checked?

#### Option 3A: Preflight Check Before Execution

Check platform constraints in `install` command before creating executor:

```go
// In cmd/tsuku/install.go
if !recipe.SupportsPlatform(runtime.GOOS, runtime.GOARCH) {
    return fmt.Errorf("recipe '%s' does not support %s/%s", name, runtime.GOOS, runtime.GOARCH)
}
```

Fail immediately without downloading dependencies or creating work directory.

**Pros:**
- Fail fast: no wasted work (downloads, directory creation)
- Clean error path: no cleanup needed
- Consistent with preflight pattern for action validation
- Easy to test: no executor mocking required

**Cons:**
- Adds another validation layer in install command
- Doesn't help with cross-compilation or platform-specific build plans
- Validation logic duplicated if other commands need it (validate, plan)

#### Option 3B: Executor Validation During Initialization

Check constraints when executor is created:

```go
// In executor.New()
if !recipe.SupportsPlatform(runtime.GOOS, runtime.GOARCH) {
    return nil, &UnsupportedPlatformError{...}
}
```

Executor initialization fails with specific error type.

**Pros:**
- Centralized: all execution paths validate automatically
- Executor contract includes platform compatibility
- Supports future cross-compilation (executor could take target platform)
- Natural extension of executor's existing validation

**Cons:**
- Fails after work directory creation (requires cleanup)
- Later in the flow than preflight (some work already done)
- Couples executor to platform detection (harder to mock in tests)

#### Option 3C: Dual Validation - Preflight + Executor

Validate in both install command (preflight) and executor (enforcement):

- Install command checks before starting (UX)
- Executor validates on initialization (safety)

**Pros:**
- Fail fast for CLI users (install command catches it early)
- Safe even if executor used directly (library usage)
- Defense in depth: multiple validation layers
- Clear separation of concerns (UX vs enforcement)

**Cons:**
- Validation logic duplicated in two places
- Potential for inconsistency if one check has bugs
- Slightly more code to maintain
- Overhead of two checks (negligible but exists)

### Decision 4: Error Messaging - What should users see when platform is unsupported?

#### Option 4A: Simple Error Message

Show supported platforms and current platform:

```
Error: hello-nix is not available for darwin/arm64

Supported platforms:
  - linux/amd64
  - linux/arm64
```

**Pros:**
- Clear and concise
- Easy to implement (unblocks other work quickly)
- Consistent with Nix's approach
- Sufficient for most use cases
- Fast implementation path

**Cons:**
- No guidance on alternatives
- Users must manually search for similar tools
- Doesn't explain why it's unsupported

#### Option 4B: Error Message with Alternatives

Suggest similar recipes that support user's platform:

```
Error: hello-nix is not available for darwin/arm64

Supported platforms: linux/*

Alternatives for darwin:
  - gnu-hello (darwin/amd64, darwin/arm64)
  - hello-cpp (darwin/arm64)

Use 'tsuku search hello' to find more options.
```

**Pros:**
- Helpful: guides users to working alternatives
- Leverages existing recipe metadata
- Reduces friction (users don't have to search manually)
- Shows value of platform metadata ecosystem

**Cons:**
- Requires recipe similarity/tagging system
- Slower (must query registry for alternatives)
- May suggest poor matches if metadata is incomplete
- More complex implementation

#### Option 4C: Error with Upstream Issue Link

Include link to upstream project's platform support:

```
Error: hello-nix is not available for darwin/arm64

Supported platforms: linux/*

Reason: nix-portable only supports Linux
See: https://github.com/DavHau/nix-portable/issues/123
```

**Pros:**
- Educational: explains the underlying limitation
- Directs users to authoritative source
- Helps users decide whether to wait for platform support or switch tools
- Low implementation cost (just embed URL in recipe)

**Cons:**
- Requires recipe authors to track upstream issues
- Links can go stale (issues closed, repos moved)
- Not actionable for users (can't use the tool now)
- Adds maintenance burden for recipe authors

## Decision Outcome

**Chosen: 1A (Allowlist Schema) + 2A (Separate OS/Arch) + 3A (Preflight Check) + 4A (Simple Error Messages)**

### Summary

We'll add `supported_os` and `supported_arch` arrays to recipe metadata, validate platform compatibility in the install command before execution begins, and show users a clear error message listing supported platforms when installation is attempted on an unsupported platform.

### Rationale

This combination best serves our Priority 0 goals (user experience and backwards compatibility) with minimal implementation complexity:

**Decision 1 - Allowlist (1A):**
- Explicit "this works on X, Y, Z" is clearer than inverted "doesn't work on A" logic
- Aligns with user mental model: "show me what platforms are supported"
- Easy for consumers (CLI, website, CI) to check: `if current_platform in supported_platforms`
- Backwards compatible: missing fields means universal support
- Rejected 1B (denylist): Inverted logic is confusing for querying and display
- Rejected 1C (combined): Adds schema complexity with minimal benefit given current use cases

**Decision 2 - Separate OS/Arch (2A):**
- Follows existing `Step.When` pattern (`when.os`, `when.arch`)
- Matches how actions currently handle platform mappings (`os_mapping`, `arch_mapping`)
- Natural for current use cases: btop (OS-specific), hello-nix (OS-specific), most tools don't care about arch
- Simple to implement and document
- Rejected 2B (tuples): Over-specifies for common cases (most recipes care about OS, not OS/arch combination)
- Rejected 2C (hybrid): Unnecessary complexity when 2A handles known use cases

**Decision 3 - Preflight Check (3A):**
- Fail fast: no wasted downloads, directory creation, or dependency resolution
- Clean error path: no executor state to clean up
- Easy to test: no mocking required
- Consistent with existing preflight validation pattern
- Rejected 3B (executor validation): Fails after work directory created, requires cleanup
- Rejected 3C (dual validation): Unnecessary duplication when 3A provides both UX and safety

**Decision 4 - Simple Error Messages (4A):**
- Fastest to implement, unblocks ecosystem integration work
- Sufficient information: users know current platform and supported platforms
- Consistent with Nix's approach (proven UX pattern)
- Can be enhanced later without breaking changes (alternative suggestions, upstream links are additive)
- Rejected 4B (alternatives): Requires similarity/tagging system not yet built
- Rejected 4C (upstream links): Adds maintenance burden, links go stale

### Trade-offs Accepted

By choosing this approach, we accept:

**1. Allowlists can become stale when new platforms are added**
- **Mitigation**: This is acceptable because:
  - Recipe authors already update recipes when upstream adds new platform binaries
  - CI tests will catch omissions (new platform added to test matrix, recipe doesn't declare support)
  - Most tools don't frequently add new platform support
  - Allowlist makes staleness visible (recipe says "only linux" when darwin now works) vs hidden (recipe says "all except windows" but freebsd doesn't work)

**2. Can't express OS/arch-specific combinations**
- **Mitigation**: This is acceptable because:
  - Current failing recipes (btop, hello-nix) are OS-specific, not combination-specific
  - Can add tuple support later (additive, backwards compatible) if use cases emerge
  - Most binaries are either OS-specific (nix-portable Linux-only) or arch-agnostic (Go binaries work on all archs)
  - Worst case: recipe declares `supported_os = ["darwin"]` and `supported_arch = ["amd64"]`, allowing darwin/arm64 to attempt install and fail - graceful degradation to current behavior

**3. No alternative suggestions in error messages**
- **Mitigation**: This is acceptable because:
  - Users can run `tsuku search <query>` manually (documented in error message)
  - Keeps initial implementation simple and shippable
  - Can be added later without breaking changes (just enhance the error formatter)
  - Alternative suggestion quality depends on recipe metadata completeness (not yet mature enough)

**4. Platform check happens at install time, not earlier (e.g., search/info)**
- **Mitigation**: This is acceptable because:
  - `tsuku info <tool>` will display platform support (separate work item)
  - Search results don't currently filter by platform (future enhancement)
  - Install-time check is sufficient to prevent the actual problem (installing incompatible tools)

## Solution Architecture

### Overview

Platform awareness is implemented as declarative metadata in recipe TOML files, validated at install-time before execution begins. The implementation extends the existing recipe schema with optional platform constraint fields, adds a validation function to check compatibility, and enhances error messages to communicate platform requirements.

### Components

**1. Recipe Schema Extension** (`internal/recipe/types.go`)

Add two optional fields to the `Metadata` struct:

```go
type Metadata struct {
    Name         string   `toml:"name"`
    Description  string   `toml:"description"`
    Homepage     string   `toml:"homepage,omitempty"`
    VersionFormat string  `toml:"version_format"`
    Tier         int      `toml:"tier,omitempty"`

    // Platform constraints (optional)
    SupportedOS   []string `toml:"supported_os,omitempty"`
    SupportedArch []string `toml:"supported_arch,omitempty"`
}
```

**Semantics:**
- Missing fields = universal support (all platforms)
- Empty array `[]` = no platforms supported (effectively disables recipe)
- Non-empty array = only listed platforms supported

**2. Platform Validation** (`internal/recipe/validation.go` or new `platform.go`)

Add a method to check if recipe supports current platform:

```go
// SupportsPlatform returns true if the recipe supports the given OS and architecture.
// Missing or empty constraint fields indicate universal support.
func (r *Recipe) SupportsPlatform(targetOS, targetArch string) bool {
    // Check OS constraint
    if len(r.Metadata.SupportedOS) > 0 {
        if !contains(r.Metadata.SupportedOS, targetOS) {
            return false
        }
    }

    // Check arch constraint
    if len(r.Metadata.SupportedArch) > 0 {
        if !contains(r.Metadata.SupportedArch, targetArch) {
            return false
        }
    }

    return true
}
```

**3. Install Command Preflight** (`cmd/tsuku/install.go`)

Add platform check before creating executor:

```go
// In installTool() function, after loading recipe
if !recipe.SupportsPlatform(runtime.GOOS, runtime.GOARCH) {
    return &UnsupportedPlatformError{
        RecipeName: recipeName,
        CurrentOS:  runtime.GOOS,
        CurrentArch: runtime.GOARCH,
        SupportedOS: recipe.Metadata.SupportedOS,
        SupportedArch: recipe.Metadata.SupportedArch,
    }
}
```

**4. Error Type** (`internal/errors/errors.go` or in `cmd/tsuku`)

Define structured error for unsupported platforms:

```go
type UnsupportedPlatformError struct {
    RecipeName   string
    CurrentOS    string
    CurrentArch  string
    SupportedOS  []string
    SupportedArch []string
}

func (e *UnsupportedPlatformError) Error() string {
    var msg strings.Builder
    fmt.Fprintf(&msg, "Error: %s is not available for %s/%s\n\n",
        e.RecipeName, e.CurrentOS, e.CurrentArch)

    // Show supported platforms
    if len(e.SupportedOS) > 0 || len(e.SupportedArch) > 0 {
        msg.WriteString("Supported platforms:\n")

        osStr := "any"
        if len(e.SupportedOS) > 0 {
            osStr = strings.Join(e.SupportedOS, ", ")
        }

        archStr := "any"
        if len(e.SupportedArch) > 0 {
            archStr = strings.Join(e.SupportedArch, ", ")
        }

        fmt.Fprintf(&msg, "  OS: %s\n", osStr)
        fmt.Fprintf(&msg, "  Architecture: %s\n", archStr)
    }

    return msg.String()
}
```

**5. Info Command Integration** (`cmd/tsuku/info.go`)

Display platform support in `tsuku info <tool>` output:

```go
// Add to printRecipeInfo() function
if len(recipe.Metadata.SupportedOS) > 0 || len(recipe.Metadata.SupportedArch) > 0 {
    fmt.Println("\nPlatform Support:")

    if len(recipe.Metadata.SupportedOS) > 0 {
        fmt.Printf("  OS: %s\n", strings.Join(recipe.Metadata.SupportedOS, ", "))
    } else {
        fmt.Println("  OS: all")
    }

    if len(recipe.Metadata.SupportedArch) > 0 {
        fmt.Printf("  Architecture: %s\n", strings.Join(recipe.Metadata.SupportedArch, ", "))
    } else {
        fmt.Println("  Architecture: all")
    }
}
```

**6. Recipe Updates**

Update existing platform-specific recipes:

- `btop.toml`: Add `supported_os = ["linux"]`
- `hello-nix.toml`: Add `supported_os = ["linux"]`

### Key Interfaces

**Recipe Metadata API:**
```go
type Metadata struct {
    // ... existing fields ...
    SupportedOS   []string `toml:"supported_os,omitempty"`
    SupportedArch []string `toml:"supported_arch,omitempty"`
}

func (r *Recipe) SupportsPlatform(targetOS, targetArch string) bool
```

**Error Handling:**
```go
type UnsupportedPlatformError struct {
    RecipeName   string
    CurrentOS    string
    CurrentArch  string
    SupportedOS  []string
    SupportedArch []string
}
```

### Data Flow

1. **Recipe Loading:**
   - User runs `tsuku install <tool>`
   - CLI loads recipe TOML from registry
   - TOML parser populates `Metadata.SupportedOS` and `Metadata.SupportedArch` (empty slices if fields missing)

2. **Platform Validation:**
   - Install command calls `recipe.SupportsPlatform(runtime.GOOS, runtime.GOARCH)`
   - Method checks if current platform matches constraints
   - Returns boolean (true = supported, false = unsupported)

3. **Error Path (Unsupported):**
   - Install command creates `UnsupportedPlatformError`
   - Error's `Error()` method formats message showing current vs supported platforms
   - User sees clear error before any installation work begins

4. **Success Path (Supported):**
   - Validation passes
   - Install command proceeds to create executor and run installation steps
   - No behavior change from current flow

5. **Info Display:**
   - User runs `tsuku info <tool>`
   - Info command reads `SupportedOS` and `SupportedArch`
   - Displays platform constraints or "all" if unrestricted

## Implementation Approach

### Phase 1: Schema and Validation (Core Functionality)

**Deliverables:**
- Extend `Metadata` struct with `SupportedOS` and `SupportedArch` fields
- Implement `Recipe.SupportsPlatform()` method
- Define `UnsupportedPlatformError` type with formatted output
- Add unit tests for validation logic (various constraint combinations)

**Testing:**
- Test missing fields (should return true for all platforms)
- Test empty arrays (should return false for all platforms)
- Test OS-only constraints
- Test arch-only constraints
- Test combined OS+arch constraints
- Test error message formatting

**Dependencies:** None (standalone addition to recipe package)

**Success Criteria:**
- Recipe TOML parsing handles new fields
- `SupportsPlatform()` correctly evaluates all combinations
- Error messages are clear and actionable

### Phase 2: CLI Integration (Preflight and Error Handling)

**Deliverables:**
- Add preflight check to `install` command
- Return `UnsupportedPlatformError` when platform check fails
- Add integration test: attempt to install platform-restricted recipe on wrong platform
- Update `info` command to display platform constraints

**Testing:**
- Integration test: create test recipe with `supported_os = ["linux"]`, attempt install on darwin
- Verify error message shows current platform and supported platforms
- Verify `tsuku info` displays platform constraints
- Verify backwards compatibility: existing recipes without constraints install normally

**Dependencies:** Phase 1 complete

**Success Criteria:**
- Install fails fast (before executor creation) when platform unsupported
- Error message guides users to understand incompatibility
- `tsuku info` shows platform support for all recipes

### Phase 3: Recipe Ecosystem Rollout (Documentation and Migration)

**Deliverables:**
- Update `btop.toml` with `supported_os = ["linux"]`
- Update `hello-nix.toml` with `supported_os = ["linux"]`
- Add recipe authoring documentation (when to use platform constraints, syntax examples)
- Update CI test matrix to skip platform-incompatible recipes

**Testing:**
- Verify btop and hello-nix show platform errors on macOS
- Verify CI skips linux-only recipes in darwin test jobs
- Manual testing: confirm error messages are helpful for real use cases

**Dependencies:** Phase 2 complete

**Success Criteria:**
- Known platform-specific recipes correctly declare constraints
- CI no longer fails on expected platform incompatibilities
- Documentation clearly explains when and how to use platform fields

## Consequences

### Positive

**User Experience:**
- Users get immediate, clear feedback when attempting to install unsupported tools
- No wasted time downloading dependencies or running installation steps
- Error messages explicitly state what platforms are supported
- `tsuku info` provides upfront visibility into platform compatibility

**Recipe Ecosystem:**
- Recipe authors can express platform constraints declaratively
- CI can intelligently skip platform-specific recipes in test matrix
- Future website integration can filter/badge recipes by platform
- Clear convention reduces ambiguity (no need to document platform support in description field)

**Implementation Quality:**
- Backwards compatible: existing recipes without constraints work unchanged
- Fail-fast validation prevents partial installation states
- Simple schema (two optional fields) keeps recipe authoring straightforward
- Extensible: can add more granular constraints (libc, ABI) as separate fields later

### Negative

**Maintenance Burden:**
- Recipe authors must update `supported_os`/`supported_arch` when upstream adds platform support
- Allowlists can become stale if recipe authors don't track upstream releases
- No automatic detection of platform support (authors must manually specify)

**Functional Limitations:**
- Cannot express complex constraints (e.g., "darwin/amd64 yes, darwin/arm64 no")
- Cannot express minimum OS versions (e.g., "macOS 12+")
- No override mechanism if user wants to force-install on unsupported platform

**User Impact:**
- Recipes without platform metadata will allow installation attempts on all platforms (gradual rollout means gaps in coverage)
- Users migrating from another package manager may expect more sophisticated platform filtering

### Mitigations

**For maintenance burden:**
- Document platform field as optional (only add when needed)
- CI recipe validation can warn when `os_mapping` is used but `supported_os` is missing (suggests platform may be restricted)
- Website can eventually show recipes missing platform metadata as "unknown support"

**For functional limitations:**
- Phase 1 design is extensible: can add `supported_platforms` array for tuples in future (backwards compatible)
- Can add `min_os_version` field if use cases emerge (separate feature)
- Can add `--force` flag to bypass check (separate feature, low priority)

**For gradual rollout:**
- Start with known-failing recipes (btop, hello-nix) to prove value
- Document pattern so community contributions include platform metadata
- Eventual CI check can flag recipes that fail only on certain platforms but lack metadata

## Security Considerations

### Download Verification

**Not applicable** - Platform-aware recipe support does not introduce new download mechanisms. This feature only adds metadata validation before existing download/installation steps run.

**Impact on existing security:** Platform validation *improves* security by preventing installation attempts on unsupported platforms before any downloads occur. This reduces the attack surface by failing fast when a recipe can't work on the current platform, avoiding potentially buggy code paths in platform-incompatible installation steps.

### Execution Isolation

**No new permissions required** - This feature adds read-only metadata parsing and boolean validation logic. It does not:
- Execute external commands
- Write to the filesystem (beyond existing recipe installation paths)
- Require elevated privileges
- Open network connections

**Runtime validation scope:**
- Reads `runtime.GOOS` and `runtime.GOARCH` (standard Go runtime constants, safe)
- Compares strings from recipe TOML against runtime values
- Returns boolean or error (no side effects)

**Existing isolation preserved:** Platform checks happen in the install command *before* executor creation, so they run in the same security context as current recipe loading and validation.

### Supply Chain Risks

**Recipe metadata trust model:**

Platform constraints are declared in recipe TOML files, which are part of the tsuku-maintained recipe registry. The trust model for platform metadata is the same as for all other recipe fields:

- **Source**: Recipes come from tsuku's GitHub repository (`internal/recipe/recipes/`)
- **Authenticity**: Recipe changes require PR review and approval by tsuku maintainers
- **Integrity**: Recipes are distributed as part of the tsuku CLI binary or fetched from GitHub (HTTPS)

**Potential risks:**

1. **False negatives (recipe omits platform constraint):**
   - Recipe works on Linux but author forgets to add `supported_os = ["linux"]`
   - User on macOS attempts install, fails with cryptic error (current behavior)
   - **Mitigation**: Not a security risk, just UX degradation. Same as current state before this feature.

2. **False positives (recipe incorrectly restricts platform):**
   - Recipe works on macOS but incorrectly declares `supported_os = ["linux"]`
   - User on macOS is blocked from installing a working tool
   - **Mitigation**: Recipe validation in CI tests installation on multiple platforms. Incorrect constraints will cause CI failures on platforms that should work.

3. **Malicious platform constraints:**
   - Attacker submits PR adding `supported_os = ["windows"]` to popular tool, blocking Linux/macOS users
   - **Mitigation**: PR review process catches suspicious changes to platform metadata. Reviewers verify constraints match upstream project's published platform support.

**Supply chain position:** Platform metadata is *less* security-critical than other recipe fields (download URLs, checksums, executable paths) because it only affects whether installation is attempted, not what is downloaded or executed.

### User Data Exposure

**No user data accessed or transmitted** - This feature does not:
- Read user files or environment variables (beyond standard `runtime.GOOS`/`runtime.GOARCH`)
- Transmit any data externally (no network calls)
- Log platform information to external services
- Collect or report platform statistics

**Runtime platform detection:** The feature reads `runtime.GOOS` and `runtime.GOARCH`, which are compile-time constants in the tsuku binary. These values are determined when tsuku itself was compiled, not derived from user environment or system files.

**Error messages:** The `UnsupportedPlatformError` includes the user's platform (`darwin/arm64`) in the error message, but this is displayed locally in the terminal, not transmitted anywhere.

### Mitigations Summary

| Risk | Mitigation | Residual Risk |
|------|------------|---------------|
| Recipe with missing constraint allows unsupported install | CI tests across platforms; false negatives degrade to current behavior | Users may attempt install on unsupported platform (same as current state) |
| Recipe with incorrect constraint blocks working install | PR review validates constraints against upstream docs; CI tests confirm | Human error in review could let incorrect constraint merge |
| Malicious platform constraint DoS (blocking legitimate users) | PR review catches suspicious constraint changes; constraint matches upstream docs | Reviewer must verify upstream platform support claims |
| Platform metadata becomes stale (upstream adds platform support) | Documentation encourages updating constraints when upstream changes; CI can eventually warn about suspected staleness | Recipe authors may not track upstream platform changes promptly |

**Overall security posture:** This feature maintains tsuku's existing security model and adds a fail-fast validation layer that *reduces* attack surface by preventing execution of platform-incompatible code paths. No new external dependencies, network calls, or privilege requirements are introduced.
