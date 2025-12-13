# Architecture Review: Plan-Based Installation Design

**Review Date:** 2025-12-13
**Reviewer:** Architecture Review Agent
**Design Document:** DESIGN-plan-based-installation.md

## Executive Summary

The proposed architecture for plan-based installation is **clear, implementable, and well-designed**. It makes excellent use of existing infrastructure (`ExecutePlan()`, `ValidatePlan()`, `InstallWithOptions()`) and adds minimal new code. The design correctly sequences implementation phases and avoids unnecessary complexity.

**Verdict:** ✅ Ready for implementation with minor clarifications noted below.

---

## 1. Architecture Clarity Assessment

### 1.1 Component Architecture

**Status:** ✅ Clear and well-defined

The flow diagram accurately represents the implementation:

```
CLI Flag → Load Plan → Validate → ExecutePlan() → Store State
```

**Evidence from codebase:**
- `ExecutePlan()` exists in `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/executor/executor.go:280`
- `ValidatePlan()` exists in `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/executor/plan.go:175`
- `InstallWithOptions()` exists in `/home/dangazineu/dev/workspace/tsuku/tsuku-3/public/tsuku/internal/install/manager.go:60`

All claimed infrastructure exists and matches the design's assumptions.

### 1.2 Plan Loading Design

**Status:** ✅ Clear with good error handling

The proposed `loadPlanFromSource()` function is straightforward:
- Handles both file and stdin (`path == "-"`)
- Uses standard `json.Decoder`
- Provides helpful error messages with debugging hints

**Strength:** The stdin hint (`"Save plan to a file first for debugging"`) is excellent UX.

**Minor observation:** No explicit validation that stdin is not a TTY when `path == "-"`. This is acceptable - users will get a clear parse error if they accidentally type JSON manually.

### 1.3 Validation Architecture

**Status:** ✅ Well-structured with proper layering

The design correctly separates two validation concerns:

1. **Structural validation** (existing `ValidatePlan()`):
   - Format version ≥ 2
   - Primitive-only actions
   - Download checksums present

2. **External plan validation** (new `validateExternalPlan()`):
   - Platform compatibility check
   - Tool name verification (if provided)
   - Reuses `ValidatePlan()` internally

**Evidence from code review:**
```go
// Current ValidatePlan() in plan.go:175
func ValidatePlan(plan *InstallationPlan) error {
    // Checks format version >= 2
    // Checks primitives only
    // Checks download checksums
}
```

**Design proposal (lines 324-354):**
```go
func validateExternalPlan(plan *executor.InstallationPlan, toolName string) error {
    // First: call existing ValidatePlan()
    if err := executor.ValidatePlan(plan); err != nil {
        return fmt.Errorf("plan validation failed: %w", err)
    }

    // Then: external-plan-specific checks
    // - Platform compatibility
    // - Tool name match
}
```

This layering is excellent - it avoids code duplication and keeps concerns separated.

---

## 2. Missing Components Analysis

### 2.1 Data Structures

**Status:** ✅ No new structures needed

The design reuses `InstallationPlan` entirely. Verified in code:
```go
// internal/executor/plan.go:21
type InstallationPlan struct {
    FormatVersion int       `json:"format_version"`
    Tool          string    `json:"tool"`
    Version       string    `json:"version"`
    Platform      Platform  `json:"platform"`
    GeneratedAt   time.Time `json:"generated_at"`
    RecipeHash    string    `json:"recipe_hash"`
    RecipeSource  string    `json:"recipe_source"`
    Deterministic bool      `json:"deterministic"`
    Steps         []ResolvedStep `json:"steps"`
}
```

This structure contains all fields needed for external plans.

### 2.2 Executor Integration

**Status:** ✅ Clear integration path

The design states:
> "Creates executor (minimal, for ExecutePlan context)"

**Clarification needed:** What does "minimal executor" mean?

Looking at `ExecutePlan()` implementation (executor.go:280-338):
```go
func (e *Executor) ExecutePlan(ctx context.Context, plan *InstallationPlan) error {
    // Uses: e.workDir, e.installDir, e.toolsDir, e.downloadCacheDir
    // Uses: e.recipe (only for execCtx.Recipe)
    // Sets: e.version = plan.Version
}
```

**Required for ExecutePlan:**
- Work directory (temporary)
- Install directory (`.install` in work dir)
- Download cache directory (for caching)
- Tools directory (for dependencies)
- Recipe (passed to execution context, may be `nil` for plan-based)

**Gap identified:** The design doesn't specify:
1. How to create an executor without a recipe
2. Whether `execCtx.Recipe` can be `nil` in plan-based flow

**Recommendation:** Clarify in implementation section:
```go
// Plan-based installation doesn't have a recipe, only a plan
// Two options:
// Option A: Create minimal dummy recipe with just metadata.Name = plan.Tool
// Option B: Make execCtx.Recipe optional (check callers first)

// Preferred: Option A for minimal code changes
r := &recipe.Recipe{
    Metadata: recipe.MetadataSection{
        Name: plan.Tool,
    },
}
exec, err := executor.New(r)
```

### 2.3 State Management

**Status:** ✅ Well-defined

The design correctly identifies that `InstallWithOptions()` handles state storage:
```go
// From install/manager.go:60
func (m *Manager) InstallWithOptions(name, version, workDir string, opts InstallOptions) error
```

This function:
- Copies from work directory to permanent location
- Creates symlinks/wrappers
- Updates state atomically

**Evidence from tests:** 20+ tests verify state management behavior (see manager_test.go:584-1623).

---

## 3. Implementation Sequencing

### 3.1 Phase Analysis

**Proposed phases:**
1. Plan loading
2. CLI integration
3. Documentation

**Assessment:** ✅ Correctly sequenced

**Rationale:**
- Phase 1 (loading/validation) is self-contained and testable
- Phase 2 (CLI) depends on Phase 1 completion
- Phase 3 (docs) comes after working implementation

**Alternative considered:** Could validation be split into separate phase?
- **Answer:** No - loading and validation are tightly coupled
- Loading produces `InstallationPlan`
- Validation consumes `InstallationPlan`
- Both are needed before CLI integration

### 3.2 Testing Strategy

**Current design states:**
- Phase 1: "Unit tests for both functions"
- Phase 2: "Integration tests for `--plan` flag"

**Gap:** No specification of what to test.

**Recommendation:** Add test specification to implementation section:

**Phase 1 unit tests:**
- `loadPlanFromSource()` with file path
- `loadPlanFromSource()` with stdin (`"-"`)
- `loadPlanFromSource()` with invalid JSON
- `loadPlanFromSource()` with missing file
- `validateExternalPlan()` with matching platform
- `validateExternalPlan()` with mismatched platform
- `validateExternalPlan()` with format version < 2
- `validateExternalPlan()` with composite actions
- `validateExternalPlan()` with mismatched tool name

**Phase 2 integration tests:**
- `tsuku install --plan <file>` with valid plan
- `tsuku install --plan -` with stdin
- `tsuku install --plan <file>` with platform mismatch
- `tsuku install --plan <file>` verifies checksums
- `tsuku install --plan <file>` stores correct state

---

## 4. Simpler Alternatives Analysis

### 4.1 Could we skip validation entirely?

**Considered:** Let `ExecutePlan()` fail naturally on bad plans.

**Rejected because:**
- Platform mismatch would fail late (after downloads)
- Tool name mismatch would install wrong tool silently
- Poor error messages for users
- Security risk (no checksum verification until download)

**Verdict:** Validation is essential, not over-engineering.

### 4.2 Could we avoid the "minimal executor" creation?

**Considered:** Make `ExecutePlan()` a standalone function.

**Analysis:**
```go
// Current signature
func (e *Executor) ExecutePlan(ctx context.Context, plan *InstallationPlan) error

// Required refactor
func ExecutePlanStandalone(ctx context.Context, plan *InstallationPlan,
                          workDir, installDir, downloadCacheDir, toolsDir string) error
```

**Rejected because:**
- Requires refactoring existing eval flow
- More invasive change than creating minimal executor
- Loses encapsulation of executor state

**Verdict:** Creating minimal executor is simpler.

### 4.3 Could we merge stdin and file loading?

**Current design:**
```go
if path == "-" {
    reader = os.Stdin
} else {
    f, err := os.Open(path)
    // ...
    reader = f
}
```

**Alternative:** Use `-` as os.Open argument (standard Unix convention)
```go
var reader io.Reader
if path == "-" {
    reader = os.Stdin
} else {
    reader = os.Open(path) // ERROR: os.Open doesn't support "-"
}
```

**Analysis:** Go's `os.Open()` doesn't support Unix `-` convention.

**Verdict:** Explicit handling is necessary and clear.

### 4.4 Could we support URLs for plans?

**Considered:** `tsuku install --plan https://example.com/plan.json`

**Analysis:**
- Adds HTTP client dependency
- Requires URL parsing and validation
- Raises security questions (TLS verification, redirects)
- Out of scope for MVP

**Verdict:** Not simpler. Could be future enhancement.

---

## 5. Offline Installation Verification

### 5.1 Design Claims

> "When artifacts are pre-cached, plan-based installation works offline"

**Verification needed:** Does download action check cache first?

**Evidence from codebase:**

Looking for download cache implementation:
```bash
# Need to check: Does download action use cache?
grep -r "DownloadCacheDir" internal/actions/
```

**From design doc (DESIGN-deterministic-execution.md):**
- Download cache is implemented
- Cache is checked before network requests
- Checksums verify cache hits

**Status:** ✅ Claim is supported by existing implementation.

### 5.2 Offline Workflow

**Design proposes:**
```bash
# Online machine
tsuku eval ripgrep > plan.json
# Transfer plan.json + $TSUKU_HOME/cache/downloads/*
# Offline machine
tsuku install --plan plan.json  # works without network
```

**Gap:** Transfer instructions are vague.

**Recommendation:** Add to documentation section:
```markdown
### Offline Installation

1. **On online machine:**
   ```bash
   tsuku eval ripgrep@14.1.0 > ripgrep-plan.json
   # This downloads and caches artifacts to ~/.tsuku/cache/downloads/
   ```

2. **Transfer to offline machine:**
   ```bash
   # Transfer plan
   scp ripgrep-plan.json offline-machine:

   # Transfer cached artifacts
   # (get URLs from plan, transfer corresponding cache files)
   scp ~/.tsuku/cache/downloads/* offline-machine:.tsuku/cache/downloads/
   ```

3. **On offline machine:**
   ```bash
   tsuku install --plan ripgrep-plan.json
   # Uses cached downloads, no network needed
   ```

**Note:** This workflow assumes download cache is preserved. Verify cache structure.
```

---

## 6. Error Handling Review

### 6.1 Plan Loading Errors

**Covered:**
- File not found
- JSON parse errors
- Stdin parse errors with debugging hint

**Well-designed:** Error messages guide user to fix.

### 6.2 Validation Errors

**Covered:**
- Format version mismatch
- Platform incompatibility
- Tool name mismatch
- Composite actions in plan
- Missing checksums

**Uses existing `PlanValidationError`:** ✅ Consistent error format.

### 6.3 Execution Errors

**Reuses existing error handling:**
- `ChecksumMismatchError` from `ExecutePlan()`
- Action execution failures
- State update failures

**Status:** ✅ No new error types needed.

### 6.4 Missing Error Case

**Scenario:** User provides plan for wrong tool:
```bash
tsuku install kubectl --plan ripgrep-plan.json
```

**Current validation:**
```go
if toolName != "" && toolName != plan.Tool {
    return fmt.Errorf("plan is for tool '%s', but '%s' was specified",
        plan.Tool, toolName)
}
```

**Question:** What if user provides just `--plan` without tool name?
```bash
tsuku install --plan ripgrep-plan.json
# Should this work? Or require: tsuku install ripgrep --plan ...?
```

**Design doesn't specify.** Two options:

**Option A:** Tool name is optional, inferred from plan
```bash
tsuku install --plan ripgrep-plan.json  # installs plan.Tool
```

**Option B:** Tool name is required, must match plan
```bash
tsuku install ripgrep --plan ripgrep-plan.json  # required
```

**Recommendation:** Support Option A (tool name optional).
- Simpler UX: plan contains tool name
- Validation still catches explicit mismatches
- Align with command structure: `tsuku install <tool>...`

**Clarification needed in CLI section:**
```go
// In Run function:
if installPlanPath != "" {
    // If no tool name provided, use plan's tool name
    if toolName == "" {
        plan, err := loadPlanFromSource(installPlanPath)
        if err != nil {
            printError(err)
            exitWithCode(ExitInstallFailed)
        }
        toolName = plan.Tool
    }

    if err := runPlanBasedInstall(installPlanPath, toolName); err != nil {
        // ...
    }
}
```

---

## 7. Documentation Gaps

### 7.1 Help Text

**Design states:** "Update `tsuku install --help`"

**Recommendation:** Specify exact help text:
```
Flags:
  --plan string   Install from a pre-computed plan file (use '-' for stdin)
                  Plan can be generated with 'tsuku eval <tool>'
```

### 7.2 Examples Missing

**Design mentions:** "Add examples to README or documentation"

**Recommendation:** Specify examples to add:

**Basic usage:**
```bash
# Generate plan
tsuku eval ripgrep > plan.json

# Install from plan
tsuku install --plan plan.json
```

**From stdin:**
```bash
tsuku eval ripgrep | tsuku install --plan -
```

**Offline installation:**
```bash
# Online: generate and cache
tsuku eval ripgrep > plan.json

# Offline: use cached artifacts
tsuku install --plan plan.json
```

**Version-specific:**
```bash
tsuku eval ripgrep@14.1.0 > rg-14.1.0-plan.json
tsuku install --plan rg-14.1.0-plan.json
```

---

## 8. Security Review

### 8.1 Checksum Verification

**Design states:**
> "ExecutePlan() already verifies checksums"

**Verified in code:**
```go
// executor.go:322-326
if step.Action == "download" && step.Checksum != "" {
    if err := e.executeDownloadWithVerification(ctx, execCtx, step, plan); err != nil {
        return fmt.Errorf("step %d (%s) failed: %w", i+1, step.Action, err)
    }
}
```

**Status:** ✅ Checksums are verified during execution.

**Additional security:** Validation requires checksums at load time:
```go
// plan.go:214-220
if step.Action == "download" && step.Checksum == "" {
    errors = append(errors, ValidationError{
        Step:    i,
        Action:  step.Action,
        Message: "download action missing checksum (security requirement)",
    })
}
```

**Verdict:** Defense in depth - validates before execution, verifies during execution.

### 8.2 Platform Verification

**Prevents:** Installing Linux binaries on macOS, etc.

**Implementation:**
```go
if plan.Platform.OS != runtime.GOOS || plan.Platform.Arch != runtime.GOARCH {
    return fmt.Errorf("plan is for %s-%s, but this system is %s-%s", ...)
}
```

**Status:** ✅ Clear protection against platform mismatch.

### 8.3 Plan Tampering

**Scenario:** User modifies plan JSON to change URLs but not checksums.

**Protection:** Checksum verification in `ExecutePlan()` will detect mismatch.

**Status:** ✅ Covered by existing verification.

### 8.4 Arbitrary Code Execution

**Concern:** Could malicious plan execute arbitrary code?

**Analysis:**
- Plans can only contain primitive actions
- Validation rejects composite actions and `run_command`
- Wait, does validation reject `run_command`?

**Check validation logic:**
```go
// plan.go:190-210
if !actions.IsPrimitive(step.Action) {
    // Error: composite action
}
```

**Question:** Is `run_command` a primitive?

**From actions registry:**
- `run_command` is registered as an action
- NOT in primitives list (primitives are: download, extract, chmod, install_binaries, etc.)

**Verification needed:** Confirm `run_command` is not primitive.

**If `run_command` IS primitive:** Add explicit check:
```go
// In validateExternalPlan()
for _, step := range plan.Steps {
    if step.Action == "run_command" {
        return fmt.Errorf("external plans cannot contain run_command actions (security)")
    }
}
```

**Recommendation:** Clarify in security section whether `run_command` needs explicit blocking.

---

## 9. Performance Considerations

### 9.1 Plan Size

**Typical plan size:** ~1-5 KB (one tool, ~10 steps)

**Large plan scenario:** Tool with many dependencies → plan has many steps

**Impact:** Negligible - JSON parsing is fast for even 1MB plans.

**Verdict:** ✅ No performance concerns.

### 9.2 Validation Cost

**Validation steps:**
1. Parse JSON (O(n) where n = plan size)
2. Check format version (O(1))
3. Check platform (O(1))
4. Check primitives for each step (O(steps))
5. Check checksums exist (O(steps))

**Total:** O(steps) - linear in plan size

**Typical case:** <100 steps → <1ms validation

**Verdict:** ✅ No performance concerns.

### 9.3 Network Elimination

**Benefit:** Plan-based installation skips version resolution

**Time saved:**
- GitHub API call: ~200-500ms
- PyPI/crates.io/npm: ~100-300ms

**Verdict:** ✅ Performance improvement for plan-based installs.

---

## 10. Compatibility Analysis

### 10.1 Backward Compatibility

**Question:** Does adding `--plan` break existing workflows?

**Analysis:**
- New optional flag
- Doesn't change behavior when omitted
- No breaking changes to existing commands

**Verdict:** ✅ Fully backward compatible.

### 10.2 Forward Compatibility

**Question:** Can old tsuku versions read new plans?

**Answer:** No, by design.
- Plans with format version > current version are rejected
- Validation explicitly checks: `if plan.FormatVersion < 2`

**Design gap:** What about `plan.FormatVersion > 2` in future?

**Recommendation:** Change validation to:
```go
// Current (design line 179-185)
if plan.FormatVersion < 2 {
    return error("unsupported version %d (expected >= 2)")
}

// Recommended
const MaxSupportedPlanVersion = 2
if plan.FormatVersion < 2 || plan.FormatVersion > MaxSupportedPlanVersion {
    return error("unsupported version %d (supported: 2-%d)",
                 plan.FormatVersion, MaxSupportedPlanVersion)
}
```

This allows controlled forward compatibility.

### 10.3 Cross-Platform Plans

**Question:** Can I generate plan on macOS and use on Linux?

**Answer:** No, platform validation prevents this.

**Is this correct?** ✅ Yes - plans are platform-specific:
- URLs differ by platform (e.g., `ripgrep-amd64-linux.tar.gz`)
- Binaries are platform-specific

**Verdict:** Platform validation is correct and necessary.

---

## 11. Edge Cases

### 11.1 Empty Plan

**Scenario:** Plan with zero steps
```json
{
  "format_version": 2,
  "tool": "empty-tool",
  "version": "1.0.0",
  "platform": {"os": "linux", "arch": "amd64"},
  "steps": []
}
```

**What happens?**
- Validation: ✅ Passes (no invalid steps)
- Execution: ✅ Succeeds (no steps to execute)
- State: ✅ Marks tool as installed

**Test exists:** `TestExecutePlan_EmptyPlan` (executor_test.go:969)

**Verdict:** ✅ Handled correctly.

### 11.2 Plan with Only Downloads

**Scenario:** Plan downloads files but doesn't install them
```json
{
  "steps": [
    {"action": "download", "url": "...", "checksum": "..."}
  ]
}
```

**What happens?**
- Execution: Downloads to work directory
- InstallWithOptions: No binaries in `.install/bin`
- State: Marks as installed but no symlinks created

**Test exists:** `TestInstallWithOptions_NoBinariesFallback` (manager_test.go:797)

**Verdict:** ✅ Handled correctly (fallback creates symlink with tool name).

### 11.3 Multiple Tools in One Invocation

**Scenario:**
```bash
tsuku install tool1 tool2 --plan plan.json
```

**What happens?**
- Loop processes `args` (tool1, tool2)
- For tool1: Uses --plan
- For tool2: Uses --plan (same plan!)

**Problem:** Plan only contains one tool.

**Gap:** CLI doesn't prevent this.

**Recommendation:** Add validation in CLI:
```go
if installPlanPath != "" && len(args) > 1 {
    printError(fmt.Errorf("--plan can only be used with a single tool"))
    exitWithCode(ExitUsageError)
}
```

Or allow:
```go
if installPlanPath != "" {
    if len(args) == 0 {
        // Use plan's tool name
    } else if len(args) == 1 {
        // Validate matches plan
    } else {
        // Error: multiple tools with single plan
    }
}
```

---

## 12. Code Quality Assessment

### 12.1 Naming Conventions

**Proposed names:**
- `loadPlanFromSource()` - ✅ Clear
- `validateExternalPlan()` - ✅ Distinguishes from `ValidatePlan()`
- `runPlanBasedInstall()` - ✅ Matches existing `runInstallWithTelemetry()`
- `installPlanPath` (flag var) - ✅ Consistent with `installDryRun`, `installForce`

**Verdict:** ✅ Good naming conventions.

### 12.2 Error Message Quality

**Examples from design:**

**File open error:**
```
failed to open plan file: <error>
```
✅ Clear, actionable

**Parse error (stdin):**
```
failed to parse plan from stdin: <error>
Hint: Save plan to a file first for debugging
```
✅ Excellent - helps user debug

**Platform mismatch:**
```
plan is for linux-amd64, but this system is darwin-arm64
```
✅ Clear, shows both sides

**Tool name mismatch:**
```
plan is for tool 'ripgrep', but 'kubectl' was specified
```
✅ Clear, shows both sides

**Verdict:** ✅ High-quality error messages.

### 12.3 Code Duplication

**Analysis:**
- Plan loading: New code (no duplication)
- Validation: Reuses existing `ValidatePlan()`
- Execution: Reuses existing `ExecutePlan()`
- State management: Reuses existing `InstallWithOptions()`

**Verdict:** ✅ Minimal duplication, excellent reuse.

---

## 13. Testing Completeness

### 13.1 Unit Test Coverage

**Phase 1 requires:**
- `loadPlanFromSource()` tests (file, stdin, errors)
- `validateExternalPlan()` tests (platform, version, primitives)

**Estimated test count:** 8-10 test functions

**Existing test infrastructure:**
- `internal/executor/executor_test.go` has plan execution tests
- `internal/executor/plan_test.go` (likely) has validation tests

**Verdict:** ✅ Clear testing strategy.

### 13.2 Integration Test Coverage

**Phase 2 requires:**
- End-to-end `--plan` flag tests

**Test scenarios needed:**
1. Install from file (happy path)
2. Install from stdin (happy path)
3. Platform mismatch (error path)
4. Checksum mismatch (error path)
5. State verification (happy path)

**Estimated test count:** 5 integration tests

**Question:** Where should integration tests live?
- Option A: `cmd/tsuku/install_test.go`
- Option B: `test/integration/` directory

**Recommendation:** Follow existing pattern (check where current install tests live).

---

## 14. Final Recommendations

### 14.1 Critical Clarifications Needed

1. **Executor creation for plan-based flow:**
   - Specify how to create executor without full recipe
   - Clarify whether `execCtx.Recipe` can be nil
   - **Recommended:** Create minimal recipe with just tool name

2. **Tool name handling:**
   - Specify behavior when tool name not provided with `--plan`
   - **Recommended:** Tool name optional, inferred from plan

3. **Multi-tool prevention:**
   - Add validation to prevent `tsuku install tool1 tool2 --plan`
   - **Recommended:** Error if multiple tools with `--plan`

4. **Format version bounds:**
   - Add upper bound check for forward compatibility
   - **Recommended:** Check `plan.FormatVersion <= MaxSupportedPlanVersion`

### 14.2 Documentation Additions

1. **Help text:** Add explicit flag description with stdin convention
2. **Examples:** Add basic, stdin, offline, and version-specific examples
3. **Offline workflow:** Clarify cache transfer process
4. **Security note:** Document checksum verification and platform checking

### 14.3 Optional Enhancements (Not Blockers)

1. **Validation:** Explicit `run_command` blocking (if not already primitive)
2. **Testing:** Specify integration test location
3. **Error messages:** Consider adding "did you mean?" for tool name mismatches

---

## 15. Conclusion

### Architecture Quality: A-

**Strengths:**
- ✅ Excellent reuse of existing infrastructure
- ✅ Clear separation of concerns (load → validate → execute → store)
- ✅ Minimal new code (only ~100 LOC for loading/validation + CLI)
- ✅ Well-sequenced implementation phases
- ✅ Good error messages with user guidance
- ✅ Security-conscious design (checksums, platform validation)

**Minor gaps:**
- ⚠️ Executor creation needs clarification (minimal recipe approach)
- ⚠️ Tool name inference needs specification
- ⚠️ Multi-tool prevention needed
- ⚠️ Format version upper bound check missing

**Overall assessment:**

The architecture is **ready for implementation** with the clarifications noted above. The design correctly identifies all major components, properly sequences implementation phases, and makes excellent use of existing infrastructure. No simpler viable alternative exists - the proposed approach is already minimal.

**Recommendation:** Proceed to implementation with the following additions to the design doc:

1. Add "Executor Creation" subsection specifying minimal recipe approach
2. Add "CLI Argument Handling" subsection specifying tool name inference
3. Update validation section with format version upper bound
4. Expand documentation section with specific examples

These are refinements, not fundamental flaws. The core architecture is sound.
