# Platform Tuple Support Architecture Review

## Executive Summary

The solution architecture for platform tuple support in `install_guide` is **fundamentally sound and ready for implementation** with minor clarifications needed. The design successfully extends existing patterns, maintains backwards compatibility, and provides clear implementation phases.

**Overall Assessment: APPROVED with clarifications**

**Key Findings:**
1. Architecture is clear and implementable
2. All necessary components identified correctly
3. Implementation phases are properly sequenced
4. Simpler alternatives exist but were correctly rejected
5. Minor gaps in validation algorithm specification
6. Type conversion handling needs explicit documentation

**Recommendation:** Proceed with implementation after addressing clarifications in Section 7.

---

## 1. Architecture Clarity Assessment

### 1.1 Is the Architecture Clear Enough to Implement?

**Answer: YES** - The architecture provides sufficient detail for implementation.

**Strengths:**

1. **Clear component identification:**
   - Modified components explicitly listed (3 total)
   - No new components required (reduces scope)
   - File paths and line numbers referenced (internal/actions/require_system.go:182-199)

2. **Algorithm specification:**
   - `getPlatformGuide()` lookup algorithm provided with pseudocode
   - Hierarchical fallback clearly defined: tuple → OS → fallback
   - Validation coverage algorithm outlined

3. **Data flow documented:**
   - Runtime execution path specified
   - Validation flow documented
   - Call sites identified (Execute() calls getPlatformGuide)

4. **Interface changes explicit:**
   - Function signature change: `func getPlatformGuide(installGuide map[string]string, platform string)` → `func getPlatformGuide(installGuide map[string]string, os, arch string)`
   - Call site update: `getPlatformGuide(installGuide, runtime.GOOS)` → `getPlatformGuide(installGuide, runtime.GOOS, runtime.GOARCH)`

**Minor Gaps:**

1. **Type conversion handling:** The design shows `install_guide` as `map[string]string` in `getPlatformGuide()` but `map[string]interface{}` in validation. Need explicit conversion logic.

2. **Validation algorithm detail:** The coverage algorithm is outlined but needs step-by-step pseudocode for edge cases (see Section 4.2).

3. **Error message format:** Validation error examples provided but should show exact format with all context.

### 1.2 Code Analysis: Current Implementation

**Current `getPlatformGuide()` (internal/actions/require_system.go:184-199):**
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

**Current call sites (2 locations in require_system.go):**
- Line 88: `guide := getPlatformGuide(installGuide, runtime.GOOS)` (command not found error)
- Line 109: `guide := getPlatformGuide(installGuide, runtime.GOOS)` (version mismatch error)

**Observation:** Only 2 call sites, both in same file. Signature change impact is minimal and localized.

**Current validation (internal/recipe/platform.go:318-331):**
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

**Observation:** Explicit TODO referencing this issue. Validation expects `map[string]interface{}` type.

**Type Conversion Issue Identified:**

The architecture document shows:
- Validation: `step.Params["install_guide"].(map[string]interface{})`
- Runtime: `getPlatformGuide(installGuide map[string]string, ...)`

**Missing:** How to convert from `map[string]interface{}` to `map[string]string`?

**Likely solution:**
```go
// In require_system.go:Execute()
installGuideRaw, _ := params["install_guide"].(map[string]interface{})
installGuide := make(map[string]string)
for k, v := range installGuideRaw {
    if str, ok := v.(string); ok {
        installGuide[k] = str
    }
}
```

This conversion logic exists somewhere (code currently works) but is not documented in architecture.

### 1.3 Implementation Clarity Verdict

**Score: 8.5/10**

**Sufficient for implementation:** YES

**Gaps to address:**
1. Document type conversion pattern
2. Add detailed validation pseudocode for mixed-granularity edge cases
3. Specify exact error message format with examples

---

## 2. Component and Interface Analysis

### 2.1 Are There Missing Components?

**Answer: NO** - All necessary components identified.

**Modified Components Analysis:**

**Component 1: `getPlatformGuide()` function**
- Location: internal/actions/require_system.go:184-199
- Change: Signature modification (add `arch` parameter)
- Impact: 2 call sites in same file
- Complexity: LOW (simple parameter addition)
- Backwards compatibility: N/A (private function)

**Component 2: `ValidateStepsAgainstPlatforms()` function**
- Location: internal/recipe/platform.go:267-335
- Change: Validation algorithm enhancement
- Impact: Called from validator.go:625
- Complexity: MEDIUM (new validation logic)
- Backwards compatibility: Must accept existing OS-level keys

**Component 3: `Execute()` call sites**
- Location: internal/actions/require_system.go:88 and :109
- Change: Add `runtime.GOARCH` parameter
- Impact: 2 lines changed
- Complexity: TRIVIAL
- Backwards compatibility: N/A (internal implementation)

**No New Components Required:** Correctly identified. This is a strength - reusing existing infrastructure.

**Infrastructure Dependencies:**

The architecture correctly identifies reuse of:
- `Recipe.GetSupportedPlatforms()` - returns platform tuples
- `TsukuSupportedOS()` and `TsukuSupportedArch()` - validation against known values
- Existing `Step.Params` map structure
- Existing error types (`StepValidationError`)

**Missing Component Analysis:**

Could additional components improve the design?

**Option: Platform Tuple Utilities**

Create helper functions:
```go
func isPlatformTuple(key string) bool
func splitPlatformTuple(key string) (os, arch string, ok bool)
func formatPlatformTuple(os, arch string) string
```

**Verdict:** Helpful but not necessary. Can be inline in validation logic.

**Option: Dedicated install_guide Type**

Create:
```go
type InstallGuide map[string]string

func (ig InstallGuide) GetGuide(os, arch string) string {
    // Implements hierarchical fallback
}
```

**Verdict:** Over-engineering. Simple function is sufficient.

**Conclusion:** No missing components. Current scope is minimal and appropriate.

### 2.2 Are the Interfaces Well-Defined?

**Answer: YES** with one clarification needed.

**Interface 1: `getPlatformGuide()` Lookup**

```go
func getPlatformGuide(installGuide map[string]string, os, arch string) string
```

**Specification (from design doc):**
1. Try exact platform tuple (`os/arch`)
2. Fall back to OS-only key (`os`)
3. Fall back to `fallback` key
4. Return empty string if no match

**Clarity:** EXCELLENT. Unambiguous behavior.

**Edge cases handled:**
- `installGuide == nil` → return ""
- Empty `os` or `arch` → will construct empty tuple, won't match, falls back to OS key
- Slash in OS name (e.g., "plan9/386") → would create ambiguous tuple "plan9/386/amd64"

**Recommendation:** Add input validation to reject OS/arch values containing "/" character.

**Interface 2: `ValidateStepsAgainstPlatforms()` Coverage**

**Specification (from design doc):**
```
For each require_system step with install_guide:
  1. Get recipe's supported platforms
  2. Extract unique OS/arch sets
  3. For each supported platform tuple:
     - Check tuple key, then OS key, then fallback
     - Report error if none found
  4. For each install_guide key:
     - Validate tuple keys are in supported platforms
     - Validate OS keys are in supported OS set
```

**Clarity:** GOOD but needs detail for step 4.

**Question:** What happens if an install_guide key doesn't match any format?

Example:
```toml
[steps.install_guide]
darwin = "..."
"invalid-key" = "..."  # Not OS, not tuple, not "fallback"
```

**Options:**
1. Validation error (strict)
2. Validation warning (permissive)
3. Silently ignore (lenient)

**Recommendation:** Validation error. Typos should be caught early.

**Missing specification:** Key validation logic.

Add to algorithm:
```
4. For each install_guide key:
   - If key == "fallback": VALID (special key)
   - Else if key contains "/":
     - Parse as "os/arch" tuple
     - Validate OS in TsukuSupportedOS()
     - Validate arch in TsukuSupportedArch()
     - Validate platform in Recipe.GetSupportedPlatforms()
   - Else:
     - Validate key in TsukuSupportedOS()
     - Validate key in recipe's supportedOS set (derived from GetSupportedPlatforms)
   - Any other format: VALIDATION ERROR
```

**Interface 3: Error Messages**

**Specification (from design doc):**
- Tuple key not in supported platforms: `"install_guide contains 'linux/riscv64' which is not in the recipe's supported platforms"`
- Missing coverage: `"install_guide missing entry for supported platform 'darwin/amd64' (no tuple key, no OS fallback, no generic fallback)"`
- Invalid tuple format: `"install_guide key 'darwin/' is invalid (must be 'os/arch' format)"`

**Clarity:** GOOD. Examples show helpful context.

**Missing:** Error messages for:
- Invalid OS in tuple: `"install_guide tuple key 'darwn/arm64' contains unknown OS 'darwn' (must be one of: darwin, linux, ...)"`
- Invalid arch in tuple: `"install_guide tuple key 'darwin/arm' contains unknown architecture 'arm' (must be one of: amd64, arm64, ...)"`
- Unknown key format: `"install_guide key 'windows-amd64' is invalid (must be OS name, 'os/arch' tuple, or 'fallback')"`

**Recommendation:** Document all error message formats in implementation section.

---

## 3. Implementation Phase Sequencing

### 3.1 Are Phases Correctly Sequenced?

**Answer: YES** - Dependencies are correctly identified.

**Phase Analysis:**

**Phase 1: Update getPlatformGuide() Function**
- Tasks: Modify signature, implement fallback, update call sites, add tests
- Dependencies: None (standalone)
- Risk: LOW (isolated function)
- Estimated complexity: 1-2 hours

**Phase 2: Update Validation Logic**
- Tasks: Modify ValidateStepsAgainstPlatforms(), add tuple validation, add tests
- Dependencies: **Depends on Phase 1** (needs to understand fallback logic)
- Risk: MEDIUM (complex edge cases)
- Estimated complexity: 4-6 hours

**Phase 3: Update Documentation**
- Tasks: Recipe format docs, action docs, migration guide, error reference
- Dependencies: **Can run in parallel** with Phase 1/2
- Risk: LOW (documentation only)
- Estimated complexity: 2-3 hours

**Phase 4: Integration Testing**
- Tasks: Test recipe, strict mode validation, platform-specific testing
- Dependencies: **Depends on Phase 1 and 2** completion
- Risk: MEDIUM (requires multiple platforms)
- Estimated complexity: 2-4 hours

**Dependency Graph:**
```
Phase 1 (getPlatformGuide) ─┬─→ Phase 4 (Integration)
                             │
Phase 2 (Validation) ────────┘

Phase 3 (Documentation) ─────→ (Independent)
```

**Critique:** Sequencing is correct but conservative.

**Optimization opportunity:**

Phase 1 could be split:
- Phase 1a: Update getPlatformGuide() signature and logic only
- Phase 1b: Add getPlatformGuide() tests

Phase 2 could start after 1a completes, with 1b running in parallel to 2.

**Verdict:** Current sequencing is safe and correct. Optimization is possible but not necessary.

### 3.2 Missing Implementation Details

**Phase 1 Detail Gaps:**

"Update all call sites" - Should specify:
- How many call sites? (Answer: 2)
- Any call sites outside require_system.go? (Answer: No)
- Need to grep for all uses? (Answer: Yes, as safety check)

**Phase 2 Detail Gaps:**

"Add helper to check if key is tuple format" - Should specify:
- Function signature: `func isTuplePlatform(key string) bool`
- Logic: `strings.Contains(key, "/") && strings.Count(key, "/") == 1`
- Placement: platform.go or require_system.go?

**Phase 4 Detail Gaps:**

"Test on actual darwin/arm64 and darwin/amd64 platforms" - This is challenging if CI only runs on one platform.

**Recommendation:** Add to implementation notes:
- CI testing uses mocked platforms (unit tests with different os/arch values)
- Manual testing on actual platforms recommended but not required for merge
- Platform-specific behavior already tested by existing platform tests

---

## 4. Validation Algorithm Edge Cases

### 4.1 Mixed Granularity Coverage

The design states validation should accept "mixed coverage where some OSes have tuple-level entries and others have OS-level entries."

**Example from design:**
```toml
supported_os = ["darwin", "linux"]
supported_arch = ["arm64", "amd64"]

[steps.install_guide]
"darwin/arm64" = "ARM-specific"
"darwin/amd64" = "Intel-specific"
linux = "Generic Linux"
```

**Coverage check:**
- darwin/arm64: ✓ tuple key
- darwin/amd64: ✓ tuple key
- linux/arm64: ✓ OS fallback
- linux/amd64: ✓ OS fallback

**Verdict:** VALID (all platforms covered)

**Edge case 1: Partial tuple coverage with fallback**
```toml
[steps.install_guide]
"darwin/arm64" = "ARM-specific"
darwin = "Intel fallback"
linux = "Generic Linux"
```

**Coverage:**
- darwin/arm64: ✓ tuple key
- darwin/amd64: ✓ OS fallback (darwin)
- linux/arm64: ✓ OS fallback (linux)
- linux/amd64: ✓ OS fallback (linux)

**Expected:** VALID

**Edge case 2: Partial tuple coverage without fallback**
```toml
[steps.install_guide]
"darwin/arm64" = "ARM-specific"
linux = "Generic Linux"
```

**Coverage:**
- darwin/arm64: ✓ tuple key
- darwin/amd64: ✗ no tuple, no darwin key, no fallback
- linux/arm64: ✓ OS fallback
- linux/amd64: ✓ OS fallback

**Expected:** VALIDATION ERROR

**Error message:** `"install_guide missing entry for supported platform 'darwin/amd64' (no tuple key 'darwin/amd64', no OS fallback 'darwin', no generic 'fallback')"`

**Edge case 3: Fallback-only coverage**
```toml
[steps.install_guide]
fallback = "Generic instructions"
```

**Coverage:**
- All platforms: ✓ fallback

**Expected:** VALID (current behavior preserved)

### 4.2 Validation Algorithm Pseudocode

The design provides high-level coverage algorithm but lacks implementation detail.

**Detailed Pseudocode (missing from design):**

```go
func ValidateInstallGuide(step Step, supportedPlatforms []string) []error {
    var errors []error

    installGuide, ok := step.Params["install_guide"].(map[string]interface{})
    if !ok || installGuide == nil {
        return nil  // No install_guide to validate
    }

    // Build OS/arch sets from supported platforms
    supportedOS := make(map[string]bool)
    supportedArch := make(map[string]bool)
    for _, platform := range supportedPlatforms {
        parts := strings.Split(platform, "/")
        if len(parts) == 2 {
            supportedOS[parts[0]] = true
            supportedArch[parts[1]] = true
        }
    }

    // Step 1: Validate each supported platform has coverage
    for _, platform := range supportedPlatforms {
        parts := strings.Split(platform, "/")
        if len(parts) != 2 {
            continue  // Malformed platform (shouldn't happen)
        }
        os, arch := parts[0], parts[1]

        // Check hierarchical fallback
        platformTuple := fmt.Sprintf("%s/%s", os, arch)
        hasCoverage := false

        // Try tuple key
        if _, ok := installGuide[platformTuple]; ok {
            hasCoverage = true
        }

        // Try OS key
        if !hasCoverage {
            if _, ok := installGuide[os]; ok {
                hasCoverage = true
            }
        }

        // Try fallback key
        if !hasCoverage {
            if _, ok := installGuide["fallback"]; ok {
                hasCoverage = true
            }
        }

        if !hasCoverage {
            errors = append(errors, &StepValidationError{
                StepIndex: stepIndex,
                Message: fmt.Sprintf(
                    "install_guide missing entry for supported platform '%s' "+
                    "(no tuple key '%s', no OS fallback '%s', no generic 'fallback')",
                    platform, platformTuple, os,
                ),
            })
        }
    }

    // Step 2: Validate each install_guide key is valid
    for key := range installGuide {
        // Skip special fallback key
        if key == "fallback" {
            continue
        }

        // Check if key is platform tuple
        if strings.Contains(key, "/") {
            parts := strings.Split(key, "/")

            // Validate tuple format
            if len(parts) != 2 {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide key '%s' is invalid (tuple must be 'os/arch' format)",
                        key,
                    ),
                })
                continue
            }

            os, arch := parts[0], parts[1]

            // Validate OS is known
            if !isKnownOS(os) {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide tuple key '%s' contains unknown OS '%s' "+
                        "(must be one of: %s)",
                        key, os, strings.Join(TsukuSupportedOS(), ", "),
                    ),
                })
            }

            // Validate arch is known
            if !isKnownArch(arch) {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide tuple key '%s' contains unknown architecture '%s' "+
                        "(must be one of: %s)",
                        key, arch, strings.Join(TsukuSupportedArch(), ", "),
                    ),
                })
            }

            // Validate platform is in recipe's supported platforms
            if !containsPlatform(supportedPlatforms, key) {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide contains '%s' which is not in the recipe's "+
                        "supported platforms (%s)",
                        key, strings.Join(supportedPlatforms, ", "),
                    ),
                })
            }
        } else {
            // OS-only key

            // Validate OS is known
            if !isKnownOS(key) {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide key '%s' is not a recognized OS, platform tuple, "+
                        "or 'fallback'",
                        key,
                    ),
                })
            }

            // Validate OS is in recipe's supported OS set
            if !supportedOS[key] {
                errors = append(errors, &StepValidationError{
                    StepIndex: stepIndex,
                    Message: fmt.Sprintf(
                        "install_guide contains OS '%s' which is not in the recipe's "+
                        "supported platforms",
                        key,
                    ),
                })
            }
        }
    }

    return errors
}

func isKnownOS(os string) bool {
    known := TsukuSupportedOS()
    for _, k := range known {
        if k == os {
            return true
        }
    }
    return false
}

func isKnownArch(arch string) bool {
    known := TsukuSupportedArch()
    for _, k := range known {
        if k == arch {
            return true
        }
    }
    return false
}

func containsPlatform(platforms []string, target string) bool {
    for _, p := range platforms {
        if p == target {
            return true
        }
    }
    return false
}
```

**Recommendation:** Add this pseudocode to design doc implementation section.

---

## 5. Alternative Architecture Analysis

### 5.1 Were Simpler Alternatives Considered?

**Answer: YES** - The design doc correctly rejected simpler but inferior alternatives.

**Alternative 1: OS-only keys (Status Quo)**

**Simplicity:** HIGHEST (no code changes)

**Rejected because:**
- Doesn't solve the problem (can't express architecture-specific guidance)
- Creates architectural inconsistency (recipe-level precision vs step-level OS-only)
- Forces recipe proliferation (separate recipes per architecture)

**Verdict:** Correctly rejected.

**Alternative 2: Separate arch_install_guide field**

**Simplicity:** MEDIUM (new field, separate validation)

**Rejected because:**
- Adds API complexity (new field)
- Introduces precedence rules (arch_install_guide overrides install_guide)
- Potential for conflicting guidance across fields
- Doesn't align with "extend existing patterns" principle

**Analysis:**

The design doc criticism is fair but overstated. This alternative is viable:

**Pros (understated in design):**
- Clear separation of concerns (OS-level vs platform-level)
- Consistent with existing os_mapping/arch_mapping pattern
- Simpler validation (each field one granularity level)
- Explicit opt-in (only use arch_install_guide when needed)

**Cons (accurate):**
- Additional schema field
- Precedence rules to learn
- Not as intuitive as single namespace with hierarchical fallback

**Why chosen option is better:**

Option 1 (extend install_guide) provides same capability with:
- Zero new fields
- Backwards compatibility (existing keys work unchanged)
- Intuitive hierarchical fallback (no precedence rules needed)
- Incremental adoption (add tuple keys only where needed)

**Verdict:** Correctly rejected, though closer decision than design doc implies.

**Alternative 3: Template-based variable substitution**

**Simplicity:** LOW (template engine complexity)

**Example:**
```toml
[steps.install_guide]
darwin = "{BREW_PREFIX}/bin/brew install gcc"

[steps.install_guide.variables]
BREW_PREFIX = { "darwin/arm64" = "/opt/homebrew", "darwin/amd64" = "/usr/local" }
```

**Rejected because:**
- Over-engineered for rare edge case
- Requires template parsing and variable resolution
- Only useful when guidance differs slightly (not general solution)
- Validation complexity (ensure all variables have coverage)

**Verdict:** Correctly rejected. Solves specific case (path differences) but not general problem (different package names, different instructions).

### 5.2 Were More Complex Alternatives Considered?

**Alternative: Callable installation scripts**

**Complexity:** HIGHEST

**Example:**
```toml
[steps.install_guide]
darwin = { script = "scripts/install-darwin.sh" }
```

**Rejected because:**
- Security concerns (arbitrary script execution)
- Portability issues (script dependencies)
- Violates static recipe principle
- Out of scope for declarative TOML format

**Verdict:** Correctly rejected. Would fundamentally change tsuku's security model.

### 5.3 Architecture Simplicity Score

**Chosen architecture complexity: MEDIUM**

**Justification:**
- Reuses existing infrastructure (GetSupportedPlatforms, validation patterns)
- Minimal API surface (no new fields or types)
- Localized changes (3 functions in 2 files)
- Hierarchical fallback is intuitive (mirrors common config patterns like DNS resolution)

**Simpler alternatives exist but don't solve the problem.**

**More complex alternatives exist but add unnecessary sophistication.**

**Verdict:** Chosen architecture is at the right complexity level - simple enough to implement and maintain, complex enough to solve the problem fully.

---

## 6. Missing Architecture Considerations

### 6.1 TOML Parsing Compatibility

**Question:** Do TOML keys with slashes parse correctly?

**Design assumption:** Yes, but not verified in architecture doc.

**TOML Specification:**

TOML supports any valid UTF-8 string as a key. Slashes are valid:

**Bare keys:**
```toml
darwin = "..."  # Valid bare key
```

**Quoted keys:**
```toml
"darwin/arm64" = "..."  # Valid quoted key
```

**Table syntax:**
```toml
[install_guide]
"darwin/arm64" = "..."
darwin = "..."
```

**BurntSushi/toml library** (used by tsuku): Supports quoted keys with any character including slashes.

**Test needed:**
```go
func TestTOMLParsePlatformTuples(t *testing.T) {
    tomlData := `
    [metadata]
    name = "test"

    [[steps]]
    action = "require_system"
    command = "test"

    [steps.install_guide]
    "darwin/arm64" = "ARM guide"
    "darwin/amd64" = "Intel guide"
    darwin = "Generic darwin"
    linux = "Generic linux"
    fallback = "Generic fallback"
    `

    var recipe Recipe
    err := toml.Unmarshal([]byte(tomlData), &recipe)
    if err != nil {
        t.Fatalf("Failed to parse TOML with platform tuples: %v", err)
    }

    guide := recipe.Steps[0].Params["install_guide"].(map[string]interface{})
    if guide["darwin/arm64"] != "ARM guide" {
        t.Errorf("Tuple key not parsed correctly")
    }
}
```

**Recommendation:** Add TOML parsing test to Phase 1 prerequisite verification.

### 6.2 Type Safety Considerations

**Current type handling:**

Validation uses: `map[string]interface{}`
Runtime uses: `map[string]string`

**Missing from architecture:** Type conversion safety.

**Question:** What if install_guide value is not a string?

**Example:**
```toml
[steps.install_guide]
darwin = 123  # Integer instead of string
```

**Current behavior:** Type assertion fails, install_guide ignored

**Proposed behavior:** Validation error during recipe load

**Implementation:**

Add to validation:
```go
for key, value := range installGuide {
    if _, ok := value.(string); !ok {
        errors = append(errors, &StepValidationError{
            StepIndex: stepIndex,
            Message: fmt.Sprintf(
                "install_guide['%s'] must be a string, got %T",
                key, value,
            ),
        })
    }
}
```

**Recommendation:** Add type validation to Phase 2 implementation.

### 6.3 Backwards Compatibility Verification

**Design claim:** "Backwards compatible - existing OS-only keys continue to work"

**Verification needed:**

**Test 1: Existing recipes unchanged**
```toml
# docker.toml - existing recipe
[steps.install_guide]
darwin = "brew install --cask docker"
linux = "See https://docs.docker.com/engine/install/"
fallback = "Visit https://docs.docker.com/get-docker/"
```

**Expected:** Works identically before and after change

**Test 2: Fallback-only guidance**
```toml
[steps.install_guide]
fallback = "Generic install instructions"
```

**Expected:** Works identically before and after change

**Test 3: OS-only keys with architecture-constrained recipe**
```toml
[metadata]
supported_arch = ["arm64"]  # Only ARM

[steps.install_guide]
darwin = "brew install gcc"
linux = "apt install gcc"
```

**Expected:** darwin key covers darwin/arm64, linux key covers linux/arm64

**Recommendation:** Add backwards compatibility regression tests to Phase 4.

### 6.4 Performance Considerations

**Runtime lookup performance:**

Current: 1-2 map lookups (platform key, fallback key)
Proposed: 1-3 map lookups (tuple key, OS key, fallback key)

**Impact:** Negligible (map lookup is O(1), happens rarely - only on missing dependency error)

**Validation performance:**

Current: O(OS_count) - iterate over unique OS values
Proposed: O(platform_count) - iterate over all platform tuples

**Example:**
- 2 OS × 3 arch = 6 platforms
- Current: 2 iterations
- Proposed: 6 iterations

**Impact:** Minimal (validation happens once at recipe load)

**Verdict:** Performance impact is not a concern.

### 6.5 Testing Infrastructure

**Design mentions:** "Ensure existing test helpers can construct recipes with tuple keys"

**Current test helpers (from require_system_test.go:145-194):**

```go
tests := []struct {
    name         string
    installGuide map[string]string
    platform     string
    want         string
}{...}
```

**After change:**

```go
tests := []struct {
    name         string
    installGuide map[string]string
    os           string
    arch         string
    want         string
}{...}
```

**Test helper impact:** Minimal - test struct field change

**Recipe construction (for validation tests):**

Current pattern:
```go
recipe := &Recipe{
    Steps: []Step{
        {
            Action: "require_system",
            Params: map[string]interface{}{
                "install_guide": map[string]interface{}{
                    "darwin": "brew install",
                    "linux": "apt install",
                },
            },
        },
    },
}
```

Adding tuple keys:
```go
"install_guide": map[string]interface{}{
    "darwin/arm64": "ARM-specific",
    "darwin/amd64": "Intel-specific",
    "linux": "Generic Linux",
},
```

**Verdict:** Test infrastructure already supports tuple keys (maps accept any string key).

---

## 7. Architecture Clarifications Needed

### 7.1 Critical Clarifications Before Implementation

**1. Type Conversion Logic**

**Current gap:** Design shows `map[string]interface{}` in validation but `map[string]string` in runtime without explaining conversion.

**Recommendation:**

Add to design doc Section "Implementation Approach":

```
Type Handling:

The install_guide parameter is stored as map[string]interface{} in Step.Params
(TOML unmarshaling generic type). Runtime execution requires map[string]string.

Conversion pattern (add to require_system.go:Execute):

```go
// Extract install_guide with type conversion
installGuide := make(map[string]string)
if guideRaw, ok := params["install_guide"].(map[string]interface{}); ok {
    for k, v := range guideRaw {
        if str, ok := v.(string); ok {
            installGuide[k] = str
        }
    }
}

guide := getPlatformGuide(installGuide, runtime.GOOS, runtime.GOARCH)
```

Validation should verify all values are strings during recipe load (Phase 2).
```

**2. Validation Algorithm Detail**

**Current gap:** High-level algorithm provided but implementation pseudocode missing.

**Recommendation:**

Add Section 4.2 pseudocode (from this review) to design doc Implementation Approach section.

**3. Error Message Specifications**

**Current gap:** Examples provided but not comprehensive.

**Recommendation:**

Add to design doc Section "Validation Error Cases":

```
Complete Error Message Reference:

1. Missing platform coverage:
   "install_guide missing entry for supported platform '{platform}' (no tuple key '{tuple}', no OS fallback '{os}', no generic 'fallback')"

2. Tuple key not in supported platforms:
   "install_guide contains '{tuple}' which is not in the recipe's supported platforms ({platforms})"

3. Invalid tuple format:
   "install_guide key '{key}' is invalid (tuple must be 'os/arch' format)"

4. Unknown OS in tuple:
   "install_guide tuple key '{tuple}' contains unknown OS '{os}' (must be one of: {known_os})"

5. Unknown architecture in tuple:
   "install_guide tuple key '{tuple}' contains unknown architecture '{arch}' (must be one of: {known_arch})"

6. OS key not in supported platforms:
   "install_guide contains OS '{os}' which is not in the recipe's supported platforms"

7. Invalid key format:
   "install_guide key '{key}' is not a recognized OS, platform tuple, or 'fallback'"

8. Non-string value:
   "install_guide['{key}'] must be a string, got {type}"
```

**4. TOML Parsing Prerequisite**

**Current gap:** Assumes TOML handles slash-containing keys but doesn't verify.

**Recommendation:**

Add to Phase 1 Prerequisites:

```
TOML Parsing Verification:

Before implementing, verify that BurntSushi/toml correctly parses quoted keys
with slashes:

```go
func TestTOMLQuotedKeys(t *testing.T) {
    input := `
    [test]
    "key/with/slash" = "value"
    `

    var result map[string]interface{}
    err := toml.Unmarshal([]byte(input), &result)
    require.NoError(t, err)
    require.Equal(t, "value", result["test"].(map[string]interface{})["key/with/slash"])
}
```

Expected: Pass (TOML spec supports this, library should too)
```

### 7.2 Optional Clarifications (Nice to Have)

**1. Migration Examples**

Show recipe authors how to add tuple support:

```toml
# Before
[steps.install_guide]
darwin = "brew install gcc"

# After (incremental enhancement)
[steps.install_guide]
"darwin/arm64" = "/opt/homebrew/bin/brew install gcc"
"darwin/amd64" = "/usr/local/bin/brew install gcc"
darwin = "brew install gcc"  # Fallback for future darwin architectures
```

**2. Decision Rationale**

Why hierarchical fallback vs separate fields?

Add to Decision Outcome:
```
Hierarchical fallback within a single map (Option 1) was chosen over separate
arch_install_guide field (Option 2) because:

1. Single namespace is more intuitive - authors specify guidance and let the
   system determine which applies
2. No precedence rules to learn - fallback is implicit in key specificity
3. Consistent with DNS, CSS, and other hierarchical config patterns
4. Backwards compatible without migration
```

**3. Testing Strategy**

Specify test coverage requirements:

```
Required Test Coverage:

Phase 1 (getPlatformGuide):
- Exact tuple match
- OS fallback when tuple missing
- Generic fallback when OS missing
- nil installGuide
- Empty os/arch
- All three levels present (tuple takes precedence)

Phase 2 (Validation):
- Complete tuple coverage (all platforms have tuple keys)
- Complete OS coverage (all platforms have OS fallback)
- Mixed coverage (some tuple, some OS)
- Partial tuple without fallback (error)
- Tuple key not in supported platforms (error)
- Invalid tuple format (error)
- Unknown OS in tuple (error)
- Unknown arch in tuple (error)
- OS key not in supported platforms (error)
- Non-string values (error)
- Fallback-only coverage (valid)

Phase 4 (Integration):
- Existing docker.toml works unchanged
- Existing cuda.toml works unchanged
- New recipe with tuple keys validates and executes
- Mixed granularity recipe works correctly
```

---

## 8. Implementation Risk Assessment

### 8.1 Risk Analysis

**Low Risk Items:**

1. **getPlatformGuide() signature change**
   - Risk: LOW
   - Impact: 2 call sites, both in same file
   - Mitigation: Compiler will catch any missed call sites

2. **Backwards compatibility**
   - Risk: LOW
   - Impact: Existing recipes continue working
   - Mitigation: Regression tests in Phase 4

**Medium Risk Items:**

1. **Validation algorithm correctness**
   - Risk: MEDIUM
   - Impact: False positives (rejecting valid recipes) or false negatives (accepting invalid recipes)
   - Mitigation: Comprehensive test matrix, manual recipe testing

2. **Edge case handling**
   - Risk: MEDIUM
   - Impact: Unexpected behavior with unusual key combinations
   - Mitigation: Explicit test cases for all edge cases in Section 4

**Negligible Risk Items:**

1. **TOML parsing**
   - Risk: NEGLIGIBLE
   - Impact: TOML library already handles quoted keys
   - Mitigation: Prerequisite verification test

2. **Performance**
   - Risk: NEGLIGIBLE
   - Impact: 1-2 additional map lookups
   - Mitigation: N/A (not a concern)

**Overall Risk: LOW-MEDIUM**

The architecture is sound and low-risk. Primary risk is validation edge cases, mitigated by thorough testing.

### 8.2 Testing Coverage Requirements

**Minimum test coverage for approval:**

- Unit tests: 90%+ coverage of new code paths
- Integration tests: All edge cases from Section 4 verified
- Regression tests: Existing recipes (docker, cuda) work unchanged
- Platform-specific tests: Mock different os/arch combinations

**Test matrix size estimate:**

Phase 1 (getPlatformGuide): ~8 test cases
Phase 2 (Validation): ~15 test cases
Phase 4 (Integration): ~5 test scenarios

Total: ~30 test cases

**Verdict:** Achievable test coverage for low-risk implementation.

---

## 9. Answered Questions

### 9.1 Is the architecture clear enough to implement?

**Answer: YES** (Score: 8.5/10)

**Sufficient detail provided:**
- Component changes identified
- Function signatures specified
- Algorithm pseudocode outlined
- Data flow documented

**Minor clarifications needed:**
- Type conversion logic
- Detailed validation pseudocode
- Complete error message reference
- TOML parsing verification

**Recommendation:** Address clarifications in Section 7.1, then proceed.

### 9.2 Are there missing components or interfaces?

**Answer: NO**

**All necessary components identified:**
- getPlatformGuide() modification
- ValidateStepsAgainstPlatforms() enhancement
- Execute() call site updates

**Reuses existing infrastructure:**
- Recipe.GetSupportedPlatforms()
- TsukuSupportedOS() / TsukuSupportedArch()
- StepValidationError type

**Optional helper functions could be added but are not necessary.**

**Verdict:** Component list is complete and minimal.

### 9.3 Are the implementation phases correctly sequenced?

**Answer: YES**

**Dependencies correctly identified:**
- Phase 2 depends on Phase 1 (validation needs fallback logic)
- Phase 4 depends on Phases 1+2 (integration needs implementation)
- Phase 3 independent (documentation can run parallel)

**Sequencing is conservative (safe) with optimization opportunity:**
- Could split Phase 1 into 1a (implementation) and 1b (tests)
- Phase 2 could start after 1a
- Current sequencing is safer for first-time implementation

**Verdict:** Sequencing is correct. Optimization possible but not required.

### 9.4 Are there simpler alternatives we overlooked?

**Answer: NO**

**Simpler alternatives considered and correctly rejected:**
- Status quo (doesn't solve problem)
- Separate arch_install_guide field (more API surface for same capability)
- Template-based substitution (over-engineered)

**Chosen architecture is at appropriate complexity level:**
- Minimal API surface (no new fields)
- Reuses existing patterns (hierarchical fallback)
- Backwards compatible (existing keys work)
- Solves problem fully (supports architecture-specific guidance)

**Could be simpler but wouldn't solve the problem.**
**Could be more complex but unnecessary sophistication.**

**Verdict:** Chosen architecture is optimally simple for the requirements.

---

## 10. Final Recommendations

### 10.1 Architecture Approval

**Status: APPROVED with minor clarifications**

**Proceed with implementation after:**

1. **Add type conversion documentation** (Section 7.1.1)
2. **Add detailed validation pseudocode** (Section 7.1.2)
3. **Add complete error message reference** (Section 7.1.3)
4. **Add TOML parsing verification** (Section 7.1.4)

**These additions are documentation updates, not architecture changes.**

### 10.2 Implementation Priorities

**Phase 1: High Priority**
- Implement getPlatformGuide() with hierarchical fallback
- Update call sites with runtime.GOARCH
- Add comprehensive unit tests

**Phase 2: High Priority**
- Implement tuple-aware validation
- Add type checking (ensure values are strings)
- Add edge case tests from Section 4

**Phase 3: Medium Priority**
- Update recipe format documentation
- Add migration guide for recipe authors
- Document new error messages

**Phase 4: Medium Priority**
- Integration testing with test recipes
- Regression testing with docker.toml, cuda.toml
- Manual verification on actual platforms (optional)

### 10.3 Success Criteria

**Must have before merge:**

1. All Phase 1 and 2 tests passing
2. Backwards compatibility verified (existing recipes work)
3. Edge cases from Section 4 validated
4. Type safety enforced (non-string values rejected)
5. Error messages clear and actionable

**Should have before merge:**

1. Documentation updated (Phase 3)
2. Integration tests passing (Phase 4)
3. TOML parsing verified

**Nice to have:**

1. Manual testing on darwin/arm64 and darwin/amd64
2. Migration examples for recipe authors
3. Performance benchmarks (expected: negligible impact)

### 10.4 Post-Implementation Validation

**After merge, verify:**

1. Existing recipes still validate and execute correctly
2. New recipes with tuple keys work on all platforms
3. Validation errors are clear and helpful
4. Documentation is accurate and complete

**Monitor for:**

1. Recipe author confusion (validate error message clarity)
2. Edge cases not covered in tests
3. Performance impact (expected: none, but verify)

---

## 11. Conclusion

**The solution architecture is fundamentally sound and ready for implementation.**

**Key Strengths:**
- Minimal API surface (extends existing field, no new types)
- Backwards compatible (existing recipes unchanged)
- Clear implementation path (4 phases with dependencies identified)
- Reuses existing infrastructure (GetSupportedPlatforms, validation patterns)
- Appropriate complexity level (simple enough to maintain, complex enough to solve problem)

**Minor Gaps:**
- Type conversion needs documentation
- Validation algorithm needs detailed pseudocode
- Error messages need complete reference
- TOML parsing needs prerequisite verification

**None of these gaps are architectural issues - they are implementation details that should be documented before coding begins.**

**Recommendation: PROCEED with implementation after addressing clarifications in Section 7.1.**

**Estimated implementation effort:**
- Phase 1: 2-3 hours
- Phase 2: 4-6 hours
- Phase 3: 2-3 hours
- Phase 4: 2-4 hours

**Total: 10-16 hours** (approximately 2 days of focused development)

**Risk level: LOW-MEDIUM** (primary risk is validation edge cases, well-mitigated by test coverage)

---

## Appendix: Code Examples

### Example 1: Complete getPlatformGuide() Implementation

```go
// getPlatformGuide returns the installation guide for the given OS and architecture.
// It uses hierarchical fallback:
//   1. Try exact platform tuple (os/arch)
//   2. Fall back to OS-only key
//   3. Fall back to "fallback" key
//   4. Return empty string if no match
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

### Example 2: Call Site Update

```go
// Before
guide := getPlatformGuide(installGuide, runtime.GOOS)

// After
guide := getPlatformGuide(installGuide, runtime.GOOS, runtime.GOARCH)
```

### Example 3: Type Conversion in Execute()

```go
// Extract install_guide parameter with type conversion
installGuide := make(map[string]string)
if guideRaw, ok := params["install_guide"].(map[string]interface{}); ok {
    for k, v := range guideRaw {
        if str, ok := v.(string); ok {
            installGuide[k] = str
        }
    }
}

// Get platform-specific guide
guide := getPlatformGuide(installGuide, runtime.GOOS, runtime.GOARCH)
```

### Example 4: Validation Test Case

```go
func TestValidateInstallGuide_MixedGranularity(t *testing.T) {
    recipe := &Recipe{
        Metadata: MetadataSection{
            SupportedOS: []string{"darwin", "linux"},
            SupportedArch: []string{"arm64", "amd64"},
        },
        Steps: []Step{
            {
                Action: "require_system",
                Params: map[string]interface{}{
                    "command": "gcc",
                    "install_guide": map[string]interface{}{
                        "darwin/arm64": "/opt/homebrew/bin/brew install gcc",
                        "darwin/amd64": "/usr/local/bin/brew install gcc",
                        "linux": "apt install gcc",  // OS fallback for all linux architectures
                    },
                },
            },
        },
    }

    errors := recipe.ValidateStepsAgainstPlatforms()
    if len(errors) != 0 {
        t.Errorf("Expected no validation errors, got: %v", errors)
    }
}
```

### Example 5: Recipe with Platform Tuples

```toml
[metadata]
name = "gcc"
supported_os = ["darwin", "linux"]
supported_arch = ["arm64", "amd64"]

[[steps]]
action = "require_system"
command = "gcc"

[steps.install_guide]
"darwin/arm64" = "/opt/homebrew/bin/brew install gcc"
"darwin/amd64" = "/usr/local/bin/brew install gcc"
linux = "apt install gcc"
fallback = "Install GCC using your system package manager"
```
