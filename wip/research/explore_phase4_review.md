# Platform Tuple Support Design Review

## Executive Summary

This review evaluates the problem statement and options analysis for adding platform tuple support to `install_guide` and `when` clauses. The analysis is **generally sound** with a **clear preference for Option 1** (extend install_guide with platform tuple keys). However, several areas need refinement before implementation.

**Key Findings:**
- Problem statement is specific and well-motivated with concrete example
- Option 1 is technically superior despite validation complexity
- Option 2 appears to be a strawman (designed to fail)
- Option 3 correctly identified as non-solution
- Missing analysis: validation edge cases and migration path
- Unstated assumption: "when" clause scope should match install_guide

**Recommendation:** Proceed with Option 1 with enhanced validation requirements (see Section 4).

---

## 1. Problem Statement Evaluation

### 1.1 Is It Specific Enough?

**Answer: YES**, with minor gaps.

**Strengths:**
- Concrete motivating example (Homebrew paths on darwin/arm64 vs darwin/amd64)
- Clear architectural inconsistency identified (recipe-level precision vs step-level OS-only)
- Explicit scope boundaries (in-scope vs out-of-scope)
- References existing implementation (PR #685)

**Gaps:**
1. **No quantification of impact**: How many existing recipes would benefit? The doc mentions docker.toml and cuda.toml use install_guide, but doesn't analyze whether they need tuple support.

2. **Missing real-world validation**: The Homebrew example is well-chosen (different paths: `/opt/homebrew` vs `/usr/local`), but the doc doesn't show this is a *current* problem in tsuku recipes. Are there recipes today that would immediately benefit?

3. **When clause ambiguity**: The problem statement mentions `when` clauses but provides minimal context. The doc states it's "undocumented and unused" - should it be in scope at all?

**Evidence from codebase:**
- Only 2 recipes use `install_guide` currently (docker.toml, cuda.toml)
- Both use OS-level keys only (darwin, linux, fallback)
- No evidence of current pain points requiring tuple-level precision

**Verification:**
```toml
# docker.toml
[steps.install_guide]
darwin = "brew install --cask docker"
linux = "See https://docs.docker.com/engine/install/ for platform-specific installation"
fallback = "Visit https://docs.docker.com/get-docker/ for installation instructions"
```

This doesn't require architecture-specific instructions today. The problem is **forward-looking** rather than addressing current pain.

### 1.2 Can Solutions Be Evaluated Against It?

**Answer: MOSTLY YES.**

The decision drivers provide evaluation criteria:
- Consistency ✓
- Backwards compatibility ✓
- Simplicity ✓
- Validation ✓
- Implementation scope ✓

However, missing criteria:
- **Usability**: How intuitive is it for recipe authors?
- **Error messages**: How clear are validation failures?
- **Documentation burden**: How much explanation is needed?

---

## 2. Missing Alternatives Analysis

### 2.1 Are There Missing Options?

**Answer: YES**, several alternatives not considered.

**Option 4: Platform-Specific install_guide Sections**

Instead of mixing granularities in one map, use platform tuples exclusively with OS-level defaults:

```toml
[[steps]]
action = "require_system"
command = "brew"

# Explicit platform tuples only
[steps.install_guide]
"darwin/arm64" = "/opt/homebrew/bin/brew install gcc"
"darwin/amd64" = "/usr/local/bin/brew install gcc"
"linux/amd64" = "apt install gcc"
"linux/arm64" = "apt install gcc"
fallback = "Install gcc using your system package manager"
```

**Pros:**
- Single granularity level (tuples only)
- No mixed-key ambiguity
- Explicit coverage requirements
- Reuses existing map structure

**Cons:**
- Verbose for simple cases (must repeat OS-only guidance across arches)
- No incremental migration path from OS-only keys
- Breaking change (existing OS-only keys would be invalid)

**Option 5: Template-Based install_guide**

Use placeholders that expand at runtime:

```toml
[steps.install_guide]
darwin = "{BREW_PREFIX}/bin/brew install gcc"
linux = "apt install gcc"

[steps.install_guide.variables]
BREW_PREFIX = { "darwin/arm64" = "/opt/homebrew", "darwin/amd64" = "/usr/local" }
```

**Pros:**
- Reduces duplication when only path differs
- Keeps OS-level structure
- Explicit variable declaration

**Cons:**
- New template system complexity
- Limited applicability (only useful when strings differ slightly)
- Unclear how to validate variable coverage

**Option 6: Callable install_guide Functions**

Defer to external scripts based on platform:

```toml
[steps.install_guide]
darwin = { script = "scripts/install-darwin.sh" }
linux = "apt install gcc"
```

**Pros:**
- Ultimate flexibility
- Handles complex platform-specific logic
- Reusable across recipes

**Cons:**
- Security concerns (executing arbitrary scripts)
- Portability issues (script dependencies)
- Out of scope for static TOML format

**Analysis:** Options 4-6 were likely rejected during design phase, but doc should acknowledge them with brief rationale.

### 2.2 Why Weren't These Considered?

Likely reasons (should be documented):
- **Option 4**: Breaking change, verbose
- **Option 5**: Over-engineered for rare edge case
- **Option 6**: Violates static recipe principle

---

## 3. Pros/Cons Fairness Evaluation

### 3.1 Option 1: Extend install_guide with Platform Tuple Keys

**Stated Pros (Fair Assessment):**
- Maximum flexibility ✓
- Backwards compatible ✓
- Intuitive fallback ✓
- Minimal API surface ✓

**Stated Cons (Needs Enhancement):**

**Con 1: "Validation complexity"**
- **Fair**: YES, but understated
- **Reality**: This is significant. Must validate:
  - Either OS-level coverage OR tuple-level coverage
  - No partial coverage gaps (e.g., darwin/arm64 present but darwin/amd64 missing without darwin fallback)
  - Tuple keys match recipe's supported platforms
  - OS keys match recipe's supported OS values

**Example validation scenario not addressed:**
```toml
supported_arch = ["arm64", "amd64"]
[steps.install_guide]
"darwin/arm64" = "..."  # arm64 specific
darwin = "..."          # Covers darwin/amd64? Or error?
```

Is this valid? The doc doesn't specify. Should be: YES (darwin key covers darwin/amd64).

**Con 2: "Mixed granularity"**
- **Fair**: YES, this is a legitimate UX concern
- **Underexplored**: Doc doesn't discuss how to prevent recipe author confusion
- **Mitigation needed**: Clear documentation examples showing when to use each granularity

**Con 3: "Partial coverage ambiguity"**
- **Fair**: YES
- **Example given is good**: "What if only darwin/arm64 is specified but not darwin/amd64?"
- **Missing**: The answer. Proposal should specify: validation MUST catch this OR fall back gracefully (document which).

**Missing Pros:**
- **Incremental adoption**: Existing recipes don't need changes
- **Composition**: Can provide specific overrides while keeping general fallback

**Missing Cons:**
- **Precedence complexity**: Three-level fallback (tuple → OS → fallback) may be hard to debug when wrong guide is shown
- **TOML map ordering**: Platform tuples with slashes in keys may be visually confusing in TOML

### 3.2 Option 2: Separate arch_install_guide Field

**Assessment: This appears to be a STRAWMAN.**

**Stated Cons (Overstated to Make Option 1 Look Better):**

**Con 1: "More API surface"**
- **Criticism**: Weak argument. One additional field is not significant API bloat.
- **Counter**: Separate concerns is often better design (see: os_mapping vs arch_mapping - they're separate!)

**Con 2: "Duplication potential"**
- **Criticism**: True, but so what? Recipe authors can choose to use only one field if no duplication needed.
- **Counter**: This is actually a PRO - explicit separation prevents accidental overwrites

**Con 3: "Less intuitive"**
- **Criticism**: Subjective. Could argue separate fields are MORE intuitive (clear separation of concerns)
- **Counter**: Similar pattern exists with os_mapping/arch_mapping

**Con 4: "Migration friction"**
- **Criticism**: Backwards compatibility is preserved (old install_guide still works)
- **Counter**: Adding tuple support is additive, not breaking

**Missing Pros for Option 2:**
- **Clearer validation**: Each field has single granularity level
- **Explicit opt-in**: Recipe authors clearly signal "I need arch-specific guides"
- **Consistent with existing pattern**: Similar to os_mapping/arch_mapping separation
- **Better error messages**: Can say "missing arch_install_guide entry" vs "missing install_guide entry at unknown granularity"

**Actual Weakness:**
The real problem with Option 2 is the lookup order: `arch_install_guide[os/arch] → install_guide[os] → install_guide[fallback]`. This creates an "override" semantic that's less intuitive than Option 1's hierarchical fallback within one namespace.

**Verdict:** Option 2 is viable but slightly inferior to Option 1. The doc's cons are exaggerated.

### 3.3 Option 3: Status Quo with Documentation

**Stated Pros/Cons (Fair Assessment):**
- All accurate
- Correctly identified as workaround, not solution

**Additional Con:**
- **Recipe naming collision**: Creating `homebrew-arm64.toml` and `homebrew-amd64.toml` means tools can't be referenced by simple name - users must know their architecture to choose the right recipe
- **Breaks tsuku install homebrew**: Which recipe would this install?

**Verdict:** Correctly dismissed.

---

## 4. Unstated Assumptions

### 4.1 Critical Assumptions Not Documented

**Assumption 1: install_guide and when clause should have same granularity**

The doc proposes extending BOTH `install_guide` and `when` clauses with platform tuples, assuming they should match in capability.

**Challenge:** `when` clauses are currently undocumented and unused (line 128-131). Why add tuple support to a feature that doesn't work yet?

**Recommendation:**
- Decouple the features
- Implement install_guide tuple support first
- Address when clause in separate issue/design (issue #686 should be split)

**Assumption 2: Validation should be strict**

The doc assumes validation should require complete platform coverage (line 318-331 shows current validation enforces OS-level coverage).

**Alternative approach:** Allow partial coverage with fallback to generic guide. This would be more flexible but potentially confusing.

**Question:** Should validation be strict (all platforms covered) or permissive (allow gaps with fallback)?

**Recommendation:** Start strict, relax if feedback indicates it's too rigid.

**Assumption 3: Platform computation from PR #685 is stable**

The doc assumes `Recipe.GetSupportedPlatforms()` API won't change. This is reasonable given PR #685 just merged, but should be stated explicitly.

**Assumption 4: TOML map keys can contain slashes**

The doc uses `"darwin/arm64"` as TOML key. While valid TOML, should verify this parses correctly with current TOML library.

**Verification needed:**
```toml
[steps.install_guide]
"darwin/arm64" = "test"  # Does this parse correctly?
```

**Assumption 5: require_system is the only action needing tuple support**

The scope focuses on `require_system` action. Are there other actions with platform-specific parameters?

**Evidence from codebase:**
- `os_mapping` and `arch_mapping` are step-level params (not action-specific)
- These already support per-platform values
- `install_guide` is specific to `require_system`

**Conclusion:** Assumption is valid - `install_guide` is action-specific, so only `require_system` affected.

### 4.2 Hidden Scope Boundaries

**What about other install_guide keys?**

Current examples show OS keys (darwin, linux) and fallback. Are there other special keys?

**From codebase (line 195-197):**
```go
// Try fallback
if guide, ok := installGuide["fallback"]; ok {
    return guide
}
```

Only `fallback` is special. Should doc explicitly state: "Any key not matching a platform tuple or OS name is ignored (except 'fallback')"?

**What about when clause syntax?**

Current code shows `when = { os = ["darwin", "linux"] }` (line 29).

Proposed syntax: `when = { platform = ["darwin/arm64", "darwin/amd64", "linux/amd64"] }`

**Issue:** This changes the when clause schema from `{os: [], arch: []}` to `{platform: []}`. Is this breaking?

**Should be:** Additive - support both syntaxes. Not addressed in doc.

---

## 5. Validation Edge Cases (Missing Analysis)

### 5.1 Complex Validation Scenarios

The doc acknowledges "validation complexity" but doesn't enumerate specific cases. Here are scenarios that need explicit validation behavior:

**Scenario 1: Mixed tuple and OS keys**
```toml
supported_os = ["darwin", "linux"]
supported_arch = ["arm64", "amd64"]

[steps.install_guide]
"darwin/arm64" = "..."
darwin = "..."
linux = "..."
```

**Question:** Is darwin/amd64 covered by the `darwin` key?
**Expected:** YES (fallback to darwin for darwin/amd64)
**Validation:** No error

**Scenario 2: Tuple without OS fallback**
```toml
[steps.install_guide]
"darwin/arm64" = "..."
linux = "..."
```

**Question:** What about darwin/amd64?
**Expected:** Validation ERROR - darwin/amd64 not covered
**Alternative:** Fall back to `fallback` key if present

**Scenario 3: Tuple key not in supported platforms**
```toml
supported_os = ["linux"]
supported_arch = ["amd64"]

[steps.install_guide]
"darwin/arm64" = "..."  # Not in supported platforms!
"linux/amd64" = "..."
```

**Expected:** Validation ERROR - darwin/arm64 not a supported platform

**Scenario 4: All platforms covered by tuples**
```toml
supported_os = ["darwin"]
supported_arch = ["arm64", "amd64"]

[steps.install_guide]
"darwin/arm64" = "..."
"darwin/amd64" = "..."
# No darwin key, no fallback key
```

**Expected:** NO ERROR - all platforms explicitly covered

**Scenario 5: Fallback key only**
```toml
[steps.install_guide]
fallback = "..."
```

**Expected:** NO ERROR - fallback covers everything (current behavior)

### 5.2 Proposed Validation Algorithm

The doc doesn't provide a validation algorithm. Here's what's needed:

```
For each supported platform tuple (os/arch):
  1. Check if install_guide[os/arch] exists → COVERED
  2. Else check if install_guide[os] exists → COVERED
  3. Else check if install_guide[fallback] exists → COVERED
  4. Else → VALIDATION ERROR

For each install_guide key (except "fallback"):
  1. Check if key matches "os/arch" format:
     - YES: Check if platform tuple is in GetSupportedPlatforms()
     - NO: Check if key is a supported OS value
  2. If neither → VALIDATION WARNING (unused key)
```

This should be documented in the design.

---

## 6. Implementation Concerns

### 6.1 getPlatformGuide() Enhancement

Current implementation (line 184-200):
```go
func getPlatformGuide(installGuide map[string]string, platform string) string {
    // platform is runtime.GOOS only
}
```

Proposed enhancement needs:
```go
func getPlatformGuide(installGuide map[string]string, os, arch string) string {
    platformTuple := fmt.Sprintf("%s/%s", os, arch)

    // Try tuple first
    if guide, ok := installGuide[platformTuple]; ok {
        return guide
    }

    // Fall back to OS
    if guide, ok := installGuide[os]; ok {
        return guide
    }

    // Fall back to fallback
    if guide, ok := installGuide["fallback"]; ok {
        return guide
    }

    return ""
}
```

**Issue:** Function signature change breaks callers. Need to audit all call sites.

### 6.2 TOML Unmarshaling

The doc states install_guide is `map[string]interface{}` in validation (line 321) but `map[string]string` in getPlatformGuide (line 184).

**Type inconsistency:**
- During validation: `step.Params["install_guide"].(map[string]interface{})`
- During execution: `map[string]string`

Need to ensure TOML unmarshaling consistently produces the right type.

### 6.3 Test Coverage Gaps

The doc mentions test locations but doesn't specify test scenarios needed:

**Required tests:**
1. Tuple key takes precedence over OS key
2. OS key used when tuple key missing
3. Fallback used when both missing
4. Validation rejects incomplete coverage
5. Validation accepts mixed granularity with full coverage
6. Validation rejects tuple keys not in supported platforms
7. TOML parsing of slash-containing keys

---

## 7. Migration Path (Not Addressed)

### 7.1 Existing Recipes

The doc states "backwards compatible" but doesn't address migration:

**Current state:** 2 recipes use install_guide (docker, cuda) with OS-only keys

**Migration needed:** NONE (backwards compatible)

**Enhancement opportunity:** Could these recipes benefit from tuple support today?

**Answer from web research:**

For docker.toml:
```toml
# Current
[steps.install_guide]
darwin = "brew install --cask docker"
```

Homebrew paths differ by architecture:
- darwin/arm64: `/opt/homebrew/bin/brew`
- darwin/amd64: `/usr/local/bin/brew`

However, `brew` command works from PATH regardless. No tuple support needed unless full path required.

**Conclusion:** Current recipes don't NEED tuple support, but might benefit for precision.

### 7.2 Documentation Updates

The doc doesn't mention documentation impact:

**Needs updating:**
- Recipe format reference (add tuple key syntax)
- require_system action docs (explain tuple fallback)
- Migration guide (how to add tuple support to existing recipes)
- Validation error reference (new error messages)

---

## 8. When Clause Analysis (Underspecified)

### 8.1 Current Status

Lines 128-131:
> The `Step.When` field exists in types.go but is not currently used:
> - Defined in recipe types but no runtime enforcement
> - Would need executor changes to support conditional step execution
> - Out of scope for install_guide support (can be addressed separately)

**Issue:** If it's "out of scope," why is it in the design doc title and considered options?

### 8.2 Proposed Syntax

Line 29 shows:
```toml
when = { os = ["darwin", "linux"] }
```

Proposed (not shown explicitly):
```toml
when = { platform = ["darwin/arm64", "linux/amd64"] }
```

**Questions not answered:**
1. Can `os`, `arch`, and `platform` coexist in same when clause?
2. Is platform = ["darwin/arm64"] equivalent to os = ["darwin"], arch = ["arm64"]?
3. How does validation work if when clause is not enforced?

### 8.3 Recommendation

**Decouple from this design.**

Create separate issue for when clause tuple support because:
- Different implementation (executor vs validation)
- Different user experience (conditional execution vs error messages)
- Different urgency (install_guide has real use case; when clause is unused)

---

## 9. Real-World Use Case Validation

### 9.1 Homebrew Paths

The motivating example is solid. Web research confirms:

**Homebrew Installation Paths by Architecture:**

| Architecture | Path | Platform |
|-------------|------|----------|
| Apple Silicon (arm64) | `/opt/homebrew` | M1/M2/M3 Macs |
| Intel (x86_64) | `/usr/local` | Intel Macs |
| Linux | `/home/linuxbrew/.linuxbrew` | Linux |

Source: [Homebrew Installation Documentation](https://docs.brew.sh/Installation)

**Installation commands differ:**
```bash
# darwin/arm64
/opt/homebrew/bin/brew install gcc

# darwin/amd64
/usr/local/bin/brew install gcc
```

**However:** If `brew` is in PATH, the command is the same. Full paths only needed if:
- User hasn't added Homebrew to PATH
- Recipe wants to be explicit about which Homebrew installation to use

**Question for design:** Is this a compelling enough use case?

**Answer:** YES, because:
1. Many users have both Intel and ARM Homebrew installed side-by-side
2. Explicit paths avoid ambiguity
3. Error messages should show correct path for user's architecture

### 9.2 Other Potential Use Cases

**Linux package managers:**
```toml
[steps.install_guide]
"linux/amd64" = "apt install gcc"
"linux/arm64" = "apt install gcc-aarch64-linux-gnu"  # Cross-compile package
```

Different package names based on architecture.

**Native vs Rosetta:**
```toml
[steps.install_guide]
"darwin/arm64" = "Native ARM binary available"
"darwin/amd64" = "Runs via Rosetta 2 on Apple Silicon"
```

Different user guidance based on architecture.

**Conclusion:** Use cases are real and valuable.

---

## 10. Answered Questions

### 10.1 Is the problem statement specific enough?

**YES**, with minor gaps:
- No current pain point quantified (forward-looking problem)
- Should clarify when clause scope
- Should provide more than one motivating example

**Recommendation:** Add section showing 2-3 concrete recipe examples that would benefit.

### 10.2 Are there missing alternatives?

**YES**:
- Template-based approach (variable substitution)
- Tuple-only keys (no OS-level keys)
- Callable scripts (out of scope but worth mentioning)

**Recommendation:** Add "Alternatives Considered and Rejected" section with brief rationale.

### 10.3 Are the pros/cons fair and complete?

**MOSTLY**, but:
- Option 1 cons understate validation complexity
- Option 2 cons are exaggerated (appears to be strawman)
- Missing pros: incremental adoption (Option 1), explicit separation (Option 2)
- Missing cons: debugging complexity (Option 1)

**Recommendation:** Rebalance pros/cons to be more objective, especially for Option 2.

### 10.4 Are there unstated assumptions?

**YES**, several critical ones:
- install_guide and when clause should evolve together (questionable)
- Validation should be strict (reasonable but not discussed)
- Slash-containing TOML keys work correctly (needs verification)
- require_system is only affected action (true but not stated)

**Recommendation:** Add "Assumptions" section to design doc.

### 10.5 Is any option a strawman?

**YES**, Option 2 appears designed to fail:
- Cons are overstated (API surface, duplication, friction)
- Missing significant pros (clear separation, explicit opt-in, consistent pattern)
- Real weakness (override semantics) not highlighted
- Conclusion feels predetermined

**Recommendation:** Either:
- Rebalance Option 2 analysis to be fair, OR
- Remove it and acknowledge Option 1 was chosen without serious alternatives

---

## 11. Recommendations for Design Doc Updates

### 11.1 Critical Updates Needed

**1. Add Validation Algorithm Section**

Document exact validation logic for mixed granularity scenarios. Specify behavior for each edge case in Section 5.1.

**2. Separate when Clause from install_guide**

Move when clause tuple support to separate design or defer to later phase. They have different implementation complexity and urgency.

**3. Add Migration Guide Section**

Even though backwards compatible, show recipe authors how to:
- Add tuple support to existing recipes
- Decide when tuple support is needed vs OS-only
- Test tuple keys during development

**4. Enhance Implementation Section**

Add:
- Function signature changes needed
- Call site audit checklist
- TOML parsing verification
- Complete test scenario matrix

**5. Add "Alternatives Considered and Rejected"**

Document why template-based, tuple-only, and other approaches were rejected.

**6. Rebalance Option 2 Analysis**

Either fairly analyze Option 2 or remove it and state Option 1 was chosen without strong alternatives.

### 11.2 Optional Enhancements

**1. Add Examples Section**

Show 3-5 complete recipe examples demonstrating:
- Simple OS-only (current behavior preserved)
- Mixed granularity (tuple for darwin, OS-only for linux)
- Full tuple coverage (all platforms explicit)
- Fallback-only (platform-agnostic tool)

**2. Add Error Message Examples**

Show what validation errors look like:
```
Error: install_guide missing coverage for platform 'darwin/amd64'
Available entries: darwin/arm64, linux
Suggestion: Add "darwin" = "..." to cover all darwin architectures
```

**3. Add Decision Record**

Document why Option 1 was chosen over alternatives with specific rationale.

---

## 12. Final Assessment

### 12.1 Overall Quality

**Score: 7/10**

**Strengths:**
- Clear problem statement with concrete example
- Good scope boundaries (in/out of scope)
- Solid understanding of existing implementation
- Correct identification of architectural inconsistency

**Weaknesses:**
- Validation complexity understated
- Option 2 appears to be strawman
- Missing validation algorithm specification
- when clause scope unclear
- Migration path not addressed

### 12.2 Readiness for Implementation

**Status: NEEDS REVISION**

**Blockers before implementation:**
1. Define exact validation algorithm for edge cases
2. Decouple when clause or clarify scope
3. Verify TOML parsing of slash-containing keys
4. Document function signature changes and call site impacts

**Non-blocking improvements:**
1. Add migration guide
2. Add error message examples
3. Add complete recipe examples
4. Rebalance Option 2 analysis

### 12.3 Recommended Decision

**Proceed with Option 1** with the following requirements:

**Required before implementation:**
- Document validation algorithm (Section 5.2)
- Separate when clause to phase 2
- Add test scenario matrix
- Verify TOML parsing compatibility

**Required before merge:**
- Update recipe format documentation
- Add migration guide
- Update validation error messages
- Add integration tests for edge cases

**Recommended improvements:**
- Add "Alternatives Considered" section
- Rebalance or remove Option 2
- Add complete recipe examples
- Show error message formats

---

## 13. Specific Actionable Feedback

### For Design Doc Author:

**Section: Context and Problem Statement**
- ✅ GOOD: Clear motivating example with Homebrew paths
- ⚠️ IMPROVE: Add 2-3 more concrete examples of recipes that would benefit
- ⚠️ IMPROVE: Clarify whether when clause is in scope or deferred

**Section: Decision Drivers**
- ✅ GOOD: Comprehensive list
- ➕ ADD: Usability, error message clarity, documentation burden

**Section: Considered Options - Option 1**
- ✅ GOOD: Correctly identified as preferred
- ⚠️ IMPROVE: Expand validation complexity con with specific scenarios
- ➕ ADD: Debugging complexity con (three-level fallback)
- ➕ ADD: Incremental adoption pro

**Section: Considered Options - Option 2**
- ⚠️ CONCERN: Appears to be strawman with exaggerated cons
- ⚠️ IMPROVE: Add missing pros (clear separation, consistent with os_mapping/arch_mapping)
- ⚠️ IMPROVE: Highlight real weakness (override semantics) vs superficial ones

**Section: Implementation Context**
- ✅ GOOD: Excellent detail on existing code
- ➕ ADD: Validation algorithm pseudocode
- ➕ ADD: Function signature change impact
- ➕ ADD: TOML parsing verification

**Missing Sections:**
- ➕ ADD: "Validation Edge Cases" with scenarios from Section 5.1
- ➕ ADD: "Migration Guide" for recipe authors
- ➕ ADD: "Alternatives Considered and Rejected"
- ➕ ADD: "Error Message Examples"
- ➕ ADD: "Test Scenarios Matrix"

---

## Appendix A: Research Sources

**Homebrew Architecture-Specific Paths:**
- [Homebrew Installation Documentation](https://docs.brew.sh/Installation)
- [Homebrew Common Issues](https://docs.brew.sh/Common-Issues)
- [Migrate from Intel to ARM brew on M1](https://github.com/orgs/Homebrew/discussions/417)
- [Cannot install in Homebrew on ARM processor error fix](https://database.guide/fix-cannot-install-in-homebrew-on-arm-processor-in-intel-default-prefix-usr-local/)

**Codebase Analysis:**
- internal/recipe/platform.go:194-220 (GetSupportedPlatforms)
- internal/recipe/platform.go:86-111 (SupportsPlatform)
- internal/recipe/platform.go:318-331 (ValidateStepsAgainstPlatforms)
- internal/actions/require_system.go:182-200 (getPlatformGuide)
- internal/recipe/types.go:178-184 (Step struct)

**Existing Recipes:**
- internal/recipe/recipes/d/docker.toml (OS-level install_guide)
- internal/recipe/recipes/c/cuda.toml (OS-level install_guide)
- 20+ recipes with os_mapping/arch_mapping pattern
