# Design Review: Plan-Based Installation

**Document**: DESIGN-plan-based-installation.md
**Reviewer**: Design Analysis Agent
**Date**: 2025-12-13

## Executive Summary

The design document for plan-based installation is **well-structured and ready for implementation** with minor clarifications needed. The problem statement is specific and measurable, the options analysis is fair and complete, and the chosen solutions form a coherent approach. However, there are a few areas where assumptions should be made explicit and validation logic needs clarification.

**Overall Assessment**: APPROVED with recommendations for minor clarifications.

---

## 1. Problem Statement Analysis

### Specificity Assessment: **STRONG**

The problem statement is specific enough to evaluate solutions against. It identifies three concrete limitations with the current architecture:

1. **Air-gapped environments** - Organizations cannot leverage pre-computed plans
2. **CI distribution** - Build pipelines cannot generate plans centrally
3. **Team standardization** - Teams cannot share exact installation specifications

Each limitation is paired with specific consequences, making it clear what success looks like.

### Measurability

The problem statement provides clear success criteria:
- ✅ Can install from externally-provided plan: `tsuku install --plan <file>`
- ✅ Supports stdin for composability: `tsuku eval tool | tsuku install --plan -`
- ✅ Works offline when artifacts are pre-cached
- ✅ Validates plans before execution

### Scope Clarity: **EXCELLENT**

The explicit "In scope / Out of scope" section is particularly strong:
- In scope: Core installation from plans, stdin support, checksum verification
- Out of scope: Plan signing, multi-tool plans, format migration, lock files

This prevents scope creep and sets clear boundaries.

### Recommendation

**No changes needed.** The problem statement is specific, measurable, and appropriately scoped.

---

## 2. Missing Alternatives Analysis

### Coverage Assessment: **COMPREHENSIVE**

The design considers three orthogonal decisions, each with 2-3 options. Let me verify completeness:

#### Decision 1: Plan Input Method (COMPLETE)

**Options considered:**
- 1A: File path only
- 1B: File path with stdin support

**Missing alternative?** Could consider:
- 1C: Environment variable (`TSUKU_PLAN=...`)
- 1D: URL support (`--plan https://...`)

**Analysis:** These are legitimately out of scope:
- Environment variables don't align with Unix CLI conventions for this use case
- URL support adds complexity (caching, authentication) without clear value over `curl | tsuku install --plan -`

**Verdict:** No critical alternatives missing.

#### Decision 2: Plan Validation Strategy (COMPLETE)

**Options considered:**
- 2A: Minimal validation (format only)
- 2B: Comprehensive pre-execution validation

**Missing alternative?** Could consider:
- 2C: Lazy validation (validate each step before execution)
- 2D: Optional validation with `--skip-validation` flag

**Analysis:**
- Lazy validation (2C) is essentially what 2A becomes in practice (late failure)
- Optional validation (2D) contradicts the security-first philosophy stated in decision drivers
- The binary choice between minimal and comprehensive is appropriate

**Verdict:** No critical alternatives missing.

#### Decision 3: Tool Name Handling (COMPLETE)

**Options considered:**
- 3A: Tool name from plan only
- 3B: Tool name required, must match plan
- 3C: Tool name optional, defaults from plan

**Missing alternative?** Could consider:
- 3D: Allow mismatch with warning (dangerous, violates safety driver)
- 3E: Infer from plan, ignore CLI argument entirely

**Analysis:**
- Allowing mismatch (3D) violates "Safety" decision driver
- Ignoring CLI argument (3E) is essentially 3A with confusing UX

The three options cover the design space: strict, flexible, or balanced.

**Verdict:** No critical alternatives missing.

### New Alternative to Consider: **NONE**

The option space is well-explored. The choices are orthogonal and collectively exhaustive.

---

## 3. Pros/Cons Fairness and Completeness

### Decision 1: Plan Input Method

#### Option 1A: File Path Only

**Pros listed:**
- Simple implementation (just read file) ✅
- Clear semantics ✅

**Cons listed:**
- Requires intermediate file for piping ✅
- Doesn't support streaming workflows ✅
- Extra disk I/O ✅

**Missing pros:** None significant.

**Missing cons:**
- Doesn't align with upstream design's explicit mention of piping support
- Inconsistent with common Unix tools (docker, kubectl use `-` for stdin)

**Fairness:** Slightly **understated** the cons. The upstream design (DESIGN-deterministic-resolution.md) explicitly mentions `tsuku eval tool | tsuku install --plan -` as the canonical workflow. This should be noted as a con for 1A.

#### Option 1B: File Path with Stdin Support

**Pros listed:**
- Supports both batch and streaming workflows ✅
- Follows Unix convention ✅
- No intermediate files needed ✅
- Enables `curl ... | tsuku install --plan -` patterns ✅

**Cons listed:**
- Slightly more complex (must detect stdin mode) ✅
- Stdin can only be read once (no retry on parse failure) ✅

**Missing cons:**
- Piping from stdin makes debugging harder (can't inspect the plan file after failure)
- Error messages should include the plan content or suggest saving to file first

**Fairness:** **Fair and complete.** The "stdin can only be read once" con is honest about the trade-off.

### Decision 2: Plan Validation Strategy

#### Option 2A: Minimal Validation (Format Only)

**Pros listed:**
- Simple implementation ✅
- Fast validation ✅
- Supports hand-edited plans ✅

**Cons listed:**
- Platform mismatches discovered during execution (late failure) ✅
- Checksum failures provide cryptic errors ✅
- May partially install before failing ✅

**Missing pros:** None significant.

**Missing cons:** None significant. The cons accurately capture the UX and safety issues.

**Fairness:** **Fair and complete.**

#### Option 2B: Comprehensive Pre-Execution Validation

**Pros listed:**
- Fails fast with clear error messages ✅
- No partial installation on validation failure ✅
- Catches stale/incompatible plans immediately ✅
- Better user experience ✅

**Cons listed:**
- More validation code ✅
- Slightly slower startup (negligible for file I/O) ✅

**Missing cons:**
- May reject valid plans that work on the current system (e.g., if platform detection evolves)
- Tighter coupling between plan format and current code version

**Analysis:** The codebase shows that `ValidatePlan()` already exists (in `internal/executor/plan.go`) and performs strict validation including:
- Format version checks
- Primitive-only action validation
- Download checksum validation

The proposed validation logic duplicates some of this but adds:
- Platform compatibility check (OS/arch matching)
- Tool name matching

**Concern:** The design's proposed `validateExternalPlan()` function (lines 324-354) doesn't mention using the existing `ValidatePlan()`. This could lead to duplicated validation logic.

**Fairness:** **Fair**, but implementation should clarify relationship with existing `ValidatePlan()`.

### Decision 3: Tool Name Handling

#### Option 3A: Tool Name from Plan Only

**Pros listed:**
- Plan is self-contained ✅
- No mismatch possible ✅
- Simpler command syntax ✅

**Cons listed:**
- Different syntax than normal install ✅
- User must inspect plan to know what will be installed ✅
- Doesn't match existing install command structure ✅

**Fairness:** **Fair and complete.**

#### Option 3B: Tool Name Required, Must Match Plan

**Pros listed:**
- Explicit about what's being installed ✅
- Catches accidental wrong-plan usage ✅
- Consistent with normal install command structure ✅

**Cons listed:**
- Redundant information (tool name in two places) ✅
- Error on mismatch requires user to fix command ✅

**Missing pros:** None significant.

**Missing cons:**
- Breaks composability: `cat plan.json | tsuku install --plan -` requires extracting tool name from JSON first
- Scripting becomes harder: `for plan in *.json; do tsuku install $(jq -r .tool $plan) --plan $plan; done`

**Fairness:** **Understated** cons. Option 3B has significant usability downsides for scripting that aren't fully captured.

#### Option 3C: Tool Name Optional, Defaults from Plan

**Pros listed:**
- Flexible for different use cases ✅
- Supports both explicit and implicit workflows ✅
- Good balance of safety and convenience ✅

**Cons listed:**
- Most complex option ✅
- Two valid syntaxes to document ✅

**Missing cons:**
- Validation logic is more complex (must handle both cases)
- Error messages must account for two modes ("tool name required in plan when not provided on CLI")

**Fairness:** **Fair.** The complexity con is accurately stated.

---

## 4. Unstated Assumptions

### Assumption 1: Plan Format Stability

**Stated:** "Plan format migration from older versions (format version 1 is stable)" is out of scope.

**Unstated assumption:** Plans are forward-compatible within the same major version. A plan generated by tsuku v1.2 will work with tsuku v1.5.

**Evidence from code:** The `PlanFormatVersion` constant is 2, and `ValidatePlan()` rejects versions < 2. This suggests plans may NOT be forward-compatible across format versions.

**Risk:** If users store plans for long-term use (months/years), format evolution could break them.

**Recommendation:** Make explicit:
- How long plans are supported (e.g., "plans are compatible within the same format version")
- Whether plan validation will reject future minor revisions or only major changes
- Document expected plan lifetime (hours/days for CI, not months/years)

### Assumption 2: Cached Artifacts Validity

**Stated:** "Offline installation when artifacts are pre-cached"

**Unstated assumption:** Cached artifacts are trusted. If the cache contains a file with the correct checksum, it's used without re-verification from an external source.

**Evidence:** The design mentions reusing `ExecutePlan()` which includes checksum verification, but doesn't clarify whether cached files bypass download or are re-verified.

**From code:** The existing download action checks cache first (mentioned in design context), so this assumption is sound.

**Recommendation:** Make explicit in "Offline Installation" section (lines 392-408):
- Cached files are verified against plan checksums
- Cache poisoning mitigated by checksum verification
- No external verification is possible in offline mode (this is by design)

### Assumption 3: Single-Tool Plans Only

**Stated:** "Multi-tool plans (plans are single-tool by design)" is out of scope.

**Unstated assumption:** A plan-based install will only ever install one tool at a time. Dependencies are handled separately.

**Evidence from code:** The `InstallationPlan` struct has singular `Tool` and `Version` fields, confirming single-tool design.

**Potential issue:** What happens if a tool has dependencies? Does `tsuku install --plan foo.json` also install dependencies, or must each dependency have its own plan?

**From DESIGN-deterministic-resolution.md:** Plans are for single tools. Dependencies would require separate plans or normal installation flow.

**Recommendation:** Add to scope section:
- "Dependencies of plan-based installs are handled through normal installation flow (not plan-based)"
- OR: "Plan-based installation assumes dependencies are already installed"

### Assumption 4: Plan Checksums Are Authoritative

**Stated in Security Considerations:** "Checksums in external plans are trusted as authoritative."

**Implication:** If a plan says "ripgrep-14.1.0.tar.gz has SHA256 abc123", tsuku will download and verify against abc123, but won't verify that abc123 is the "correct" checksum for ripgrep 14.1.0.

**This is actually correct behavior**, but worth making more explicit in the problem statement or decision drivers.

**Recommendation:** Add to Decision Drivers:
- "Trust model: Plans are code. Users must verify plan source before execution."

### Assumption 5: Platform Compatibility Is Strict

**From proposed validation (lines 333-336):**
```go
if plan.Platform.OS != runtime.GOOS || plan.Platform.Arch != runtime.GOARCH {
    return fmt.Errorf("plan is for %s-%s, but this system is %s-%s", ...)
}
```

**Unstated assumption:** Plans are strictly platform-specific. A Linux amd64 plan cannot be used on Linux arm64, even if the binaries are universal/compatible.

**Potential issue:** Some tools distribute universal binaries (e.g., Java jars, Python scripts). The strict platform check would reject a plan that could actually work.

**Recommendation:** Either:
1. Accept this limitation and document it clearly
2. Add a future enhancement for platform-independent plans
3. Allow platform checks to be advisory (warning) rather than errors in some cases

**Preferred:** Accept limitation. Universal binaries can use the normal install flow, not plan-based flow.

---

## 5. Strawman Detection

### Analysis Methodology

A strawman option is one that's intentionally weak to make another option look better. Indicators:
- Unrealistically bad cons
- Missing obvious pros
- Unrealistic implementation assumptions

### Decision 1: Plan Input Method

**Option 1A (File Path Only)** - NOT a strawman.
- It's a legitimate simple implementation
- Some tools do work this way (e.g., Docker Compose requires file path in most cases)
- The cons are real (no piping), not exaggerated

**Verdict:** Legitimate option.

### Decision 2: Plan Validation Strategy

**Option 2A (Minimal Validation)** - NOT a strawman.
- It's a valid "trust the input" philosophy
- The cons (late failure, partial install) are real trade-offs, not invented problems
- Some tools do work this way (e.g., shell scripts often fail mid-execution)

**Verdict:** Legitimate option. The design chose 2B for good reasons, not because 2A was artificially weak.

### Decision 3: Tool Name Handling

**Option 3A (Plan Only)** - NOT a strawman.
- It's the most "plan-centric" approach
- The con "doesn't match existing install command structure" is real but not disqualifying
- Some tools use this approach (e.g., `kubectl apply -f plan.yaml` doesn't require resource name on CLI)

**Option 3B (Required Match)** - NOT a strawman.
- It's the most "safety-first" approach
- The cons are real but could be acceptable for strict environments
- Git uses similar patterns (`git checkout <branch>` is explicit)

**Verdict:** All three options are legitimate. 3C is genuinely the best balance, not just compared to weak alternatives.

### Overall Strawman Assessment: **NONE DETECTED**

All options represent legitimate design choices with honest trade-off analysis.

---

## 6. Missing Validation Details

### Gap 1: Relationship with Existing `ValidatePlan()`

**Issue:** The design proposes `validateExternalPlan()` (lines 324-354) but doesn't mention the existing `ValidatePlan()` function in `internal/executor/plan.go`.

**Existing validation logic:**
```go
// From plan.go lines 175-227
func ValidatePlan(plan *InstallationPlan) error {
    // Format version check (must be >= 2)
    // Primitive-only action check
    // Download checksum check
}
```

**Proposed validation logic:**
```go
// From design lines 324-354
func validateExternalPlan(plan *executor.InstallationPlan, toolName string) error {
    // Format version check (must equal PlanFormatVersion)
    // Platform compatibility check (NEW)
    // Tool name check (NEW)
    // Primitive-only action check
}
```

**Overlap:** Format version and primitive-only checks are duplicated.

**Recommendation:** Clarify in implementation section:
- "Call existing `ValidatePlan()` first for format/structure checks"
- "Add platform and tool name validation specific to external plans"
- This avoids duplication and ensures consistent validation

### Gap 2: Format Version Strictness

**Inconsistency:**

Existing `ValidatePlan()`:
```go
if plan.FormatVersion < 2 {  // Rejects old versions, allows newer
```

Proposed `validateExternalPlan()`:
```go
if plan.FormatVersion != executor.PlanFormatVersion {  // Rejects anything != 2
```

**Question:** Should external plans accept newer format versions? What if tsuku v1.5 can read v3 plans backward-compatibly?

**Recommendation:** Align with existing logic:
```go
if plan.FormatVersion < executor.PlanFormatVersion {
    return fmt.Errorf("unsupported plan format version %d (minimum supported: %d)", ...)
}
if plan.FormatVersion > executor.PlanFormatVersion {
    // Log warning but proceed? Or reject? Document the policy.
}
```

### Gap 3: Primitive Action Validation

**From proposed code (lines 346-350):**
```go
for i, step := range plan.Steps {
    if !actions.IsPrimitive(step.Action) {
        return fmt.Errorf("step %d uses non-primitive action '%s'; plan may be corrupted", ...)
    }
}
```

**This duplicates existing `ValidatePlan()` logic.** See plan.go lines 188-211.

**Recommendation:** Remove duplication. Call `ValidatePlan()` first, then add external-plan-specific checks (platform, tool name).

---

## 7. Security Considerations Review

### Coverage: **COMPREHENSIVE**

The security section (lines 436-477) addresses:
- Download verification ✅
- Execution isolation ✅
- Supply chain risks ✅
- User data exposure ✅
- Plan file trust model ✅

Each threat has analysis and mitigation. The residual risks are honestly stated.

### Depth Assessment

**Download Verification** (lines 439-442)
- Correctly notes checksums are authoritative from plan
- Correctly notes ChecksumMismatchError fails installation
- **Missing:** What happens if download succeeds but checksum is wrong? Does it leave partial files? (Answer from code: No, work directory is cleaned up on error)

**Execution Isolation** (lines 444-452)
- Correctly identifies sandboxing to work directory
- Correctly notes primitive-only actions
- **Enhancement:** Could mention that work directories are mode 0700 (from executor.go line 37)

**Supply Chain Risks** (lines 454-464)
- Honestly states residual risk: "users who accept plans from untrusted sources may install malicious software"
- Correctly notes HTTPS enforcement
- **Missing:** Mention that plan generation itself (tsuku eval) inherits any compromise in the registry/recipe system

**User Data Exposure** (lines 466-472)
- Correctly notes no credentials in plans
- **Enhancement:** Plans do expose what tools you're installing (metadata exposure), but this is acceptable

### Recommendation

**Minor enhancements only.** The security analysis is honest and thorough. Add:
1. Explicit note that work directory cleanup happens on error
2. Mention that plan generation inherits upstream trust (registry, recipes)
3. Consider adding "Threat: Cache Poisoning" section

---

## 8. Implementation Gaps

### Gap 1: Error Handling for Stdin Parse Failures

**From Option 1B con:** "Stdin can only be read once (no retry on parse failure)"

**Question:** What error message does the user get if they pipe invalid JSON?

**Recommendation:** Add to implementation section:
```go
if err := decoder.Decode(&plan); err != nil {
    if path == "-" {
        return nil, fmt.Errorf("failed to parse plan from stdin: %w\nHint: Save to file first for debugging", err)
    }
    return nil, fmt.Errorf("failed to parse plan from %s: %w", path, err)
}
```

### Gap 2: State Storage Details

**From lines 380-385:** "Integration with Existing Flow" mentions state storage is unchanged.

**Question:** Does plan-based installation store the plan in state.json? Or just the tool version?

**From DESIGN-deterministic-resolution.md line 66:** "Plan storage in state.json for installed tools"

**Implication:** Plan-based install should store the plan in state for re-install.

**Recommendation:** Add explicit step in Phase 2 implementation:
- "Store plan in state.json alongside version info"
- "Enables re-install to detect if plan has changed"

### Gap 3: Dependency Handling

**Not addressed:** What happens if the tool in the plan has dependencies?

**Scenario:**
```bash
# Plan for 'foo' which depends on 'bar'
tsuku install --plan foo-plan.json
```

**Questions:**
- Does tsuku install 'bar' automatically (normal dependency resolution)?
- Or does it fail with "dependency 'bar' not installed"?
- Or does it assume 'bar' is already installed?

**Recommendation:** Add to "Scope" section under "Out of scope":
- "Dependency resolution for plan-based installs (plan-based install assumes dependencies are already installed via normal flow)"

---

## 9. Alignment with Upstream Design

### Reference: DESIGN-deterministic-resolution.md

**Upstream deliverables for Milestone 3 (lines 102-114):**
- `tsuku install --plan <file>` ✅ Addressed
- Support for piping: `tsuku eval tool | tsuku install --plan -` ✅ Addressed (chosen option 1B)
- Offline installation when artifacts are pre-downloaded ✅ Addressed (lines 392-408)

**Upstream philosophy (lines 9-17):**
- "A recipe is a program that produces a deterministic installation plan" ✅ Honored
- "Enable air-gapped deployments" ✅ Addressed
- "Separate recipe evaluation from plan execution" ✅ Maintained

**Upstream integration notes (lines 133-145):**
- LLM validation's air-gapped container execution is conceptually similar ✅ Acknowledged
- Should reuse PreDownloader for checksums ✅ Not addressed in this design (out of scope - plan generation is separate)

### Alignment Assessment: **EXCELLENT**

The design faithfully implements Milestone 3 as specified in the upstream strategic design.

---

## 10. Recommendations Summary

### Critical (Must Address)

1. **Clarify validation logic relationship** (Section 6, Gap 1)
   - Call existing `ValidatePlan()` first
   - Add external-plan-specific checks (platform, tool name)
   - Avoid duplicating primitive/checksum validation

2. **Specify dependency handling** (Section 8, Gap 3)
   - Add to scope: Plan-based install assumes dependencies already installed
   - OR: Dependencies installed via normal flow, not from plan

3. **Align format version checking** (Section 6, Gap 2)
   - Use `< PlanFormatVersion` not `!= PlanFormatVersion`
   - Document policy for newer format versions

### Important (Should Address)

4. **Make assumptions explicit** (Section 4)
   - Plan lifetime expectations (hours/days for CI, not archival)
   - Platform compatibility is strict (no cross-platform plans)
   - Cached artifacts trusted based on checksum alone

5. **Enhance Option 3B cons** (Section 3)
   - Note scripting complexity: must extract tool name from JSON
   - Breaks simple piping workflows

6. **Add stdin error handling details** (Section 8, Gap 1)
   - Error message should hint at saving to file for debugging

### Nice to Have (Consider)

7. **Security enhancements** (Section 7)
   - Mention work directory cleanup on error
   - Note plan generation inherits upstream trust
   - Consider cache poisoning threat

8. **State storage clarification** (Section 8, Gap 2)
   - Explicitly state plan is stored in state.json
   - Enables future re-install detection

---

## 11. Final Verdict

### Problem Statement: ✅ APPROVED
- Specific, measurable, appropriately scoped
- Clear success criteria

### Options Analysis: ✅ APPROVED with minor notes
- No critical alternatives missing
- Pros/cons are fair (minor omissions noted above)
- No strawman options detected

### Chosen Solution: ✅ APPROVED
- 1B + 2B + 3C is coherent and well-justified
- Trade-offs are honestly stated
- Aligns with upstream design requirements

### Implementation Approach: ⚠️ NEEDS CLARIFICATION
- Address validation logic duplication
- Clarify dependency handling
- Align format version checking

**Overall Recommendation:** APPROVE for implementation with the critical clarifications above. The design is sound and ready for tactical execution once the validation logic relationship and dependency handling are made explicit.

---

## Appendix: Code Evidence

### Evidence 1: Existing Validation Logic

From `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/executor/plan.go`:

```go
// Lines 16-17: Format version is 2
const PlanFormatVersion = 2

// Lines 175-227: ValidatePlan function
func ValidatePlan(plan *InstallationPlan) error {
    var errors []ValidationError

    // Check format version
    if plan.FormatVersion < 2 {
        errors = append(errors, ValidationError{
            Step:    -1,
            Action:  "",
            Message: fmt.Sprintf("unsupported plan format version %d (expected >= 2)", plan.FormatVersion),
        })
    }

    // Validate each step
    for i, step := range plan.Steps {
        // Check if action is a primitive
        if !actions.IsPrimitive(step.Action) {
            // ... detailed error handling ...
        }

        // Check checksum for download actions
        if step.Action == "download" && step.Checksum == "" {
            errors = append(errors, ValidationError{
                Step:    i,
                Action:  step.Action,
                Message: "download action missing checksum (security requirement)",
            })
        }
    }

    if len(errors) > 0 {
        return &PlanValidationError{Errors: errors}
    }
    return nil
}
```

### Evidence 2: ExecutePlan Signature

From `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/executor/executor.go`:

```go
// Line 280: ExecutePlan accepts InstallationPlan
func (e *Executor) ExecutePlan(ctx context.Context, plan *InstallationPlan) error {
    fmt.Printf("Executing plan: %s@%s\n", plan.Tool, plan.Version)
    fmt.Printf("   Work directory: %s\n", e.workDir)

    // Store version for later use
    e.version = plan.Version

    // Create execution context from plan
    execCtx := &actions.ExecutionContext{
        Context:          ctx,
        WorkDir:          e.workDir,
        InstallDir:       e.installDir,
        // ... other fields ...
    }
    // ... execution continues ...
}
```

This confirms that `ExecutePlan()` is the correct integration point and already exists.

### Evidence 3: InstallationPlan Structure

From `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/executor/plan.go`:

```go
// Lines 21-44: InstallationPlan struct
type InstallationPlan struct {
    FormatVersion int       `json:"format_version"` // Currently 2
    Tool          string    `json:"tool"`           // Single tool name
    Version       string    `json:"version"`        // Single version
    Platform      Platform  `json:"platform"`       // OS and Arch
    GeneratedAt   time.Time `json:"generated_at"`
    RecipeHash    string    `json:"recipe_hash"`
    RecipeSource  string    `json:"recipe_source"`
    Deterministic bool      `json:"deterministic"`
    Steps         []ResolvedStep `json:"steps"`
}
```

Confirms single-tool design (not multi-tool).
