# Issue Decomposition Review: Plan-Based Installation

## Executive Summary

**Overall Assessment**: The issue decomposition is well-structured and appropriate for the design scope. The three issues correctly capture the implementation phases and have proper dependencies. However, there are several gaps in coverage that would leave parts of the design unimplemented.

**Critical Findings**:
1. **State storage integration is underspecified** - Issue 2 mentions "State storage integration" but doesn't detail what needs to be implemented for plan-based installations
2. **Error handling coverage is incomplete** - Several error scenarios from the design are not explicitly mentioned
3. **Minimal executor creation is not addressed** - The design specifies creating a minimal recipe/executor for ExecutePlan context, but this is not mentioned in any issue
4. **Missing test scenarios** - Integration tests in Issue 2 don't specify coverage for key workflows like stdin input and offline installation

**Recommendation**: Split Issue 2 into two issues and add explicit coverage for missing implementation details.

---

## 1. Design Scope Coverage Analysis

### In Scope Items

| Design Requirement | Covered By | Assessment |
|-------------------|------------|------------|
| `tsuku install --plan <file>` | Issue 2 | ✅ Fully covered |
| Support for stdin (`--plan -`) | Issue 2 | ✅ Covered in implementation |
| Checksum verification | Issue 2 | ⚠️ Implicit (via ExecutePlan) but not mentioned |
| Clear error handling | Issues 1 & 2 | ⚠️ Partially covered, lacks detail |
| Plan loading from file/stdin | Issue 1 | ✅ Fully covered |
| Plan validation (platform, format, tool name) | Issue 1 | ✅ Fully covered |

### Design Components vs Issues

| Component | Issue Coverage | Gap Analysis |
|-----------|----------------|--------------|
| Plan Loading (`loadPlanFromSource`) | Issue 1 | ✅ Complete |
| Plan Validation (`validateExternalPlan`) | Issue 1 | ✅ Complete |
| CLI Flag (`--plan`) | Issue 2 | ✅ Complete |
| Orchestration (`runPlanBasedInstall`) | Issue 2 | ⚠️ Missing executor creation details |
| State Storage (InstallWithOptions) | Issue 2 | ❌ Underspecified - see below |
| Documentation | Issue 3 | ✅ Complete |

**State Storage Gap**: The design shows that `runPlanBasedInstall` must:
1. Create a minimal recipe for executor context
2. Call `ExecutePlan()`
3. Store result using `InstallWithOptions()`

Issue 2 mentions "State storage integration" but doesn't specify:
- How to create the minimal recipe (just metadata.name field)
- What `InstallOptions` to use for plan-based installation
- How to populate runtime dependencies (if any)
- How to handle the RequestedVersion field

From `install_deps.go` (lines 450-468), normal installation does:
```go
installOpts.RequestedVersion = versionConstraint
installOpts.Plan = executor.ToStoragePlan(plan)
// ... resolve dependencies ...
mgr.InstallWithOptions(toolName, version, exec.WorkDir(), installOpts)
```

Plan-based installation should store the plan similarly, but with different dependency handling since the plan doesn't resolve dependencies (per design's "Out of scope" section).

---

## 2. Missing Implementation Issues

### Gap 1: Executor Creation for Plan-Based Install

**Where in design**: "Executor Creation for Plan-Based Install" section (lines 395-432)

**What's missing**: The design shows detailed code for creating a minimal recipe:
```go
minimalRecipe := &recipe.Recipe{
    Metadata: recipe.MetadataSection{
        Name: plan.Tool,
    },
}
exec, err := executor.NewWithVersion(minimalRecipe, plan.Version)
```

This is a key implementation detail that must be tested. Issue 2 doesn't mention it.

**Impact**: Without this, implementers might try to reuse normal recipe loading, which would fail for plan-based installation since we don't want to load/parse the full recipe.

**Recommendation**: Add explicit mention in Issue 2 or create a separate issue for "Executor creation without full recipe loading".

### Gap 2: State Storage Specifics

**Where in design**: "Executor Creation for Plan-Based Install" section, lines 427-431

**What's missing**:
- What `InstallOptions` fields to populate for plan-based installation
- Whether to set `IsExplicit = true` (likely yes, since user directly invoked install)
- How to handle `RequestedVersion` (use plan.Version? or empty?)
- Whether to store the plan via `InstallOptions.Plan` field

**Impact**: State inconsistency between normal and plan-based installations could break other commands (list, update, remove).

**Recommendation**: Either expand Issue 2's "State storage integration" bullet or create Issue 2.5 for state integration testing.

### Gap 3: Error Handling Scenarios

**Where in design**: Throughout, but especially in validation sections

**Error scenarios from design that should be tested**:
1. Plan for wrong platform (linux vs darwin)
2. Plan for wrong architecture (amd64 vs arm64)
3. Tool name mismatch (plan says "ripgrep", user says "rg")
4. Invalid JSON in plan file
5. Stdin parsing failure (with helpful error message)
6. Unsupported format version
7. Composite actions in plan (should fail validation)
8. Download checksum mismatch during execution

Issue 1 covers basic validation unit tests, but Issue 2's integration tests don't explicitly list these scenarios.

**Recommendation**: Expand Issue 2's integration test description to include specific error scenarios.

### Gap 4: Offline Installation Workflow

**Where in design**: "Offline Installation" section (lines 447-464)

**What's missing**: The design explicitly calls out the offline workflow as a key use case:
```bash
# Online machine
tsuku eval ripgrep > plan.json
# Transfer plan.json and $TSUKU_HOME/cache/downloads/* to offline machine
# Offline machine
tsuku install --plan plan.json  # works without network
```

No issue mentions testing this workflow.

**Impact**: Offline installation is a primary driver for this feature (air-gapped deployments). Without testing, we can't verify it works.

**Recommendation**: Add to Issue 2 or create a separate integration test issue for offline scenarios.

---

## 3. Dependency Analysis

### Declared Dependencies

```
Issue 1 (no deps)
    └── Issue 2 (blocked by Issue 1)
        └── Issue 3 (blocked by Issue 2)
```

**Assessment**: ✅ Dependencies are correct and linear.

### Dependency Correctness

| From | To | Relationship | Correct? |
|------|-----|--------------|----------|
| Issue 2 | Issue 1 | Uses `loadPlanFromSource` and `validateExternalPlan` | ✅ Yes |
| Issue 3 | Issue 2 | Documents implemented feature | ✅ Yes |

**Circular Dependencies**: ❌ None found

**Missing Dependencies**: ❌ None found

**Hidden Dependencies**:
- Issue 2 depends on existing `ExecutePlan()` (internal/executor) - ✅ Already implemented
- Issue 2 depends on existing `InstallWithOptions()` (internal/install) - ✅ Already implemented
- Issue 1 depends on existing `ValidatePlan()` (internal/executor) - ✅ Already implemented

All hidden dependencies are satisfied by existing code.

---

## 4. Issue Atomicity Analysis

### Issue 1: Plan Loading Utilities

**Scope**:
- `loadPlanFromSource()` for file and stdin
- `validateExternalPlan()` wrapping `ValidatePlan()`
- Unit tests

**Size**: Small - Two straightforward functions with unit tests

**Atomicity**: ✅ Appropriate
- Single responsibility: plan I/O and validation
- Can be reviewed/tested independently
- No complex orchestration

**Could be split?** No - the two functions are tightly coupled (load then validate)

**Should be combined?** No - already minimal

### Issue 2: CLI Integration

**Scope**:
- `--plan` flag
- Optional tool name handling
- `runPlanBasedInstall()` orchestration
- State storage integration
- Integration tests

**Size**: Medium-Large - Multiple components plus integration tests

**Atomicity**: ⚠️ **Should be split**

This issue combines:
1. CLI flag parsing and argument validation
2. Orchestration function (`runPlanBasedInstall`)
3. Executor creation without full recipe
4. State storage integration
5. Integration testing

**Recommendation**: Split into:
- **Issue 2a**: CLI flag and orchestration (functional implementation)
- **Issue 2b**: State storage integration and comprehensive integration tests

Rationale:
- Issue 2a can be tested manually with temporary state (proof of concept)
- Issue 2b ensures production-ready state handling
- Allows parallel work if needed (one dev on CLI, another on state edge cases)
- Clearer review scope

### Issue 3: Documentation

**Scope**: Help text, README examples, workflow documentation

**Size**: Small

**Atomicity**: ✅ Appropriate
- Single responsibility: user-facing documentation
- Can be done independently after implementation
- No code changes

---

## 5. Sequencing and Value Delivery

### Current Sequence

```
Phase 1: Issue 1 (Plan Loading)
    ↓
Phase 2: Issue 2 (CLI Integration)
    ↓
Phase 3: Issue 3 (Documentation)
```

**Assessment**: ✅ Good sequencing, but value delivery is back-loaded

### Value Delivery Timeline

| After Completion Of | User Value | Developer Value |
|---------------------|------------|-----------------|
| Issue 1 | None (internal utilities) | Can unit test plan loading |
| Issue 2 | ✅ Full feature usable | Can integration test e2e |
| Issue 3 | ✅ Discoverable via docs | External contributors can understand |

**First Usable Milestone**: After Issue 2 (all value delivered at once)

### Alternative Sequencing for Earlier Value?

**Not feasible** - This feature requires all components (load → validate → execute → store) to be useful. The current sequence is optimal.

### Sequencing Risks

1. **Issue 1 could block Issue 2 longer than expected** if edge cases emerge (stdin handling, large files, etc.)
   - Mitigation: Issue 1 is small and straightforward, unlikely to take long

2. **Issue 2 complexity could delay overall delivery** since all value is in Issue 2
   - Mitigation: Split Issue 2 as recommended above

---

## 6. Integration Points Analysis

### CLI Integration Points

| Component | Integration Point | Issue Coverage |
|-----------|------------------|----------------|
| Cobra flag parsing | `installCmd.Flags()` | Issue 2 ✅ |
| Argument validation | Multiple tools vs `--plan` conflict | Issue 2 ⚠️ (not explicit) |
| Executor creation | `executor.NewWithVersion()` | Issue 2 ⚠️ (not explicit) |
| Plan execution | `exec.ExecutePlan()` | Issue 2 ✅ |
| State storage | `mgr.InstallWithOptions()` | Issue 2 ⚠️ (underspecified) |
| Error handling | `ChecksumMismatchError` | Issue 2 ❌ (not mentioned) |

### Data Flow Integration

```
loadPlanFromSource (Issue 1)
    ↓ InstallationPlan
validateExternalPlan (Issue 1)
    ↓ validated InstallationPlan
runPlanBasedInstall (Issue 2)
    ↓ create minimal recipe
    ↓ executor.NewWithVersion (Issue 2 - not explicit)
    ↓ exec.ExecutePlan (Issue 2)
    ↓ mgr.InstallWithOptions (Issue 2 - underspecified)
    ↓ state updated
```

**Missing explicit coverage**: Executor creation and state storage details

---

## 7. Test Coverage Analysis

### Unit Tests (Issue 1)

**Specified**: "Unit tests"

**Should cover** (not explicit in issue):
- `loadPlanFromSource` with file path
- `loadPlanFromSource` with stdin ("-")
- `loadPlanFromSource` with non-existent file
- `loadPlanFromSource` with invalid JSON
- `validateExternalPlan` with platform mismatch
- `validateExternalPlan` with tool name mismatch
- `validateExternalPlan` with matching tool name
- `validateExternalPlan` with empty tool name (optional)
- `validateExternalPlan` with unsupported format version
- `validateExternalPlan` with composite actions

**Recommendation**: Add test scenario list to Issue 1 description.

### Integration Tests (Issue 2)

**Specified**: "Integration tests"

**Should cover** (not explicit in issue):
1. `tsuku install --plan plan.json` (happy path)
2. `cat plan.json | tsuku install --plan -` (stdin)
3. `tsuku install ripgrep --plan plan.json` (explicit tool name match)
4. `tsuku install rg --plan plan-ripgrep.json` (tool name mismatch - error)
5. `tsuku install --plan plan-linux.json` on Darwin (platform error)
6. `tsuku install --plan invalid.json` (parse error)
7. `tsuku install --plan -` with empty stdin (error)
8. `tsuku install foo bar --plan plan.json` (multiple tools error)
9. Verify state after plan-based installation
10. Verify binaries in ~/.tsuku/bin after installation

**Offline scenario** (not mentioned):
11. Install with pre-cached artifacts (no network)

**Recommendation**: Expand Issue 2 with explicit integration test scenarios.

---

## 8. Comparison with Design Implementation Phases

### Design Phases vs Issues

| Design Phase | Issue Mapping | Alignment |
|--------------|---------------|-----------|
| Phase 1: Plan Loading | Issue 1 | ✅ Exact match |
| Phase 2: CLI Integration | Issue 2 | ⚠️ Partial - missing details |
| Phase 3: Documentation | Issue 3 | ✅ Exact match |

**Assessment**: Good alignment, but Issue 2 needs more detail to match Phase 2 scope.

### Design Phase 2 Includes (from design):

> - Add `--plan` flag to install command
> - Add `runPlanBasedInstall()` function that:
>   1. Loads plan from source
>   2. Validates plan
>   3. Creates executor (minimal, for ExecutePlan context)
>   4. Calls ExecutePlan()
>   5. Stores result in state
> - Integration tests for `--plan` flag

**Issue 2 description**:
> - `--plan` flag accepting file or "-"
> - Optional tool name handling
> - `runPlanBasedInstall()` orchestration
> - State storage integration
> - Integration tests

**Missing from Issue 2**:
- "Creates executor (minimal, for ExecutePlan context)" - This is a key detail
- Explicit call to ExecutePlan() - Implied but not stated
- "Stores result in state" - Mentioned as "State storage integration" but no detail

---

## 9. Recommended Issue Structure

### Recommended Split

**Issue 1**: Plan Loading Utilities (no change)
- `loadPlanFromSource()` for file and stdin
- `validateExternalPlan()` wrapping `ValidatePlan()`
- Unit tests (expand with test scenario list)

**Issue 2a**: CLI Flag and Orchestration
- Add `--plan` flag to install command
- Implement `runPlanBasedInstall()` orchestration:
  1. Load plan via `loadPlanFromSource()`
  2. Validate plan via `validateExternalPlan()`
  3. Create minimal recipe for executor context
  4. Create executor via `NewWithVersion()`
  5. Execute plan via `ExecutePlan()`
  6. Store in state via `InstallWithOptions()` (basic implementation)
- Handle optional tool name argument
- Handle multiple tools error case
- Basic integration test (happy path only)

**Issue 2b**: State Integration and Comprehensive Testing
- Determine correct `InstallOptions` values for plan-based installation
- Test state consistency with normal installation
- Comprehensive integration tests:
  - File and stdin input
  - Explicit tool name matching/mismatching
  - Platform and architecture validation errors
  - Invalid plan formats
  - Offline installation with cached artifacts
  - Multiple tools error case
- Verify state correctness after installation

**Issue 3**: Documentation (no change)
- Update `--help` text
- Add README examples
- Document air-gapped and CI workflows

### Dependency Graph (Recommended)

```
Issue 1 (no deps)
    └── Issue 2a (blocked by Issue 1)
        ├── Issue 2b (blocked by Issue 2a)
        └── Issue 3 (blocked by Issue 2a, can parallel with 2b)
```

Issue 3 only needs basic functionality (2a), not comprehensive testing (2b), so they can proceed in parallel.

---

## 10. Specific Recommendations

### For Issue 1
1. **Add explicit test scenario list** covering:
   - File path loading (success and failure)
   - Stdin loading (success and failure)
   - JSON parsing errors
   - Platform validation (match and mismatch)
   - Tool name validation (match, mismatch, optional)
   - Format version validation
   - Composite action detection

2. **Add acceptance criteria**:
   - [ ] `loadPlanFromSource()` handles file paths and `-`
   - [ ] `loadPlanFromSource()` returns helpful errors for stdin failures
   - [ ] `validateExternalPlan()` calls `ValidatePlan()` for structural checks
   - [ ] `validateExternalPlan()` validates platform compatibility
   - [ ] `validateExternalPlan()` validates tool name if provided
   - [ ] All error messages are user-friendly
   - [ ] Unit test coverage > 90%

### For Issue 2 (or 2a if split)
1. **Add executor creation detail**:
   - Create minimal recipe with only `Metadata.Name` field
   - Use `executor.NewWithVersion()` with plan's version

2. **Add state storage detail**:
   - Use `DefaultInstallOptions()` as base
   - Set `RequestedVersion` to plan.Version (or empty?)
   - Set `Plan` to `executor.ToStoragePlan(plan)`
   - Set `IsExplicit` to true (user directly requested)

3. **Add argument validation detail**:
   - Error if multiple tools provided with `--plan`
   - Allow zero or one tool name
   - Validate tool name matches plan if provided

4. **Add acceptance criteria**:
   - [ ] `--plan` flag accepts file path or `-`
   - [ ] Optional tool name validated against plan
   - [ ] Multiple tools with `--plan` returns clear error
   - [ ] Minimal recipe created with correct tool name
   - [ ] Executor executes plan successfully
   - [ ] State updated with correct version and plan
   - [ ] Binaries linked to `~/.tsuku/bin`

### For Issue 2b (if split)
1. **Define comprehensive test matrix** (see section 7 above)

2. **Add offline testing**:
   - Pre-cache artifacts
   - Verify installation succeeds without network

3. **Add state verification tests**:
   - Compare state from normal vs plan-based installation
   - Verify `tsuku list` shows correct version
   - Verify `tsuku remove` works correctly

### For Issue 3
1. **Add specific documentation sections**:
   - `--help` text examples
   - README section for plan-based installation
   - Air-gapped workflow example
   - CI/CD pipeline example

2. **Add acceptance criteria**:
   - [ ] `tsuku install --help` shows `--plan` flag
   - [ ] README has plan-based installation section
   - [ ] Air-gapped workflow documented with complete example
   - [ ] CI workflow example provided

---

## 11. Security and Safety Considerations

### Covered by Issues

| Security Concern | Issue Coverage | Assessment |
|------------------|----------------|------------|
| Platform validation | Issue 1 | ✅ Covered |
| Checksum verification | Issue 2 (via ExecutePlan) | ⚠️ Implicit, not tested |
| Format version check | Issue 1 | ✅ Covered |
| Unknown action rejection | Issue 1 | ✅ Covered (ValidatePlan) |
| Composite action rejection | Issue 1 | ✅ Covered (ValidatePlan) |

### Missing from Issues

**Checksum mismatch testing**: The design explicitly discusses checksum verification as a security requirement, but no issue mentions testing `ChecksumMismatchError` handling.

**Recommendation**: Add to Issue 2b integration tests:
- Modify cached artifact to trigger checksum mismatch
- Verify clear error message
- Verify installation fails without partial state

---

## 12. Final Assessment

### Coverage Score

| Criterion | Score | Notes |
|-----------|-------|-------|
| Design scope coverage | 85% | Missing executor creation and state details |
| Implementation completeness | 80% | Core functionality covered, edge cases underspecified |
| Test coverage | 70% | Integration tests underspecified |
| Dependency correctness | 100% | Linear, no circular deps |
| Atomicity appropriateness | 75% | Issue 2 should be split |
| Sequencing effectiveness | 90% | Good sequence, could improve value delivery with split |

**Overall Score**: 83% (Good, with improvements needed)

### Critical Gaps Summary

1. **Executor creation without full recipe** - Not explicitly mentioned in Issue 2
2. **State storage specifics** - Mentioned but not detailed
3. **Integration test scenarios** - Not enumerated
4. **Offline installation testing** - Not mentioned
5. **Error handling coverage** - Partial

### Actionable Recommendations

**High Priority**:
1. Split Issue 2 into 2a (CLI) and 2b (Testing)
2. Add executor creation details to Issue 2a
3. Add state storage specifics to Issue 2a or 2b
4. Enumerate integration test scenarios in Issue 2b

**Medium Priority**:
5. Add test scenario lists to Issue 1
6. Add offline testing to Issue 2b
7. Add checksum mismatch testing to Issue 2b

**Low Priority**:
8. Add acceptance criteria to all issues
9. Expand Issue 3 with specific documentation sections

### Conclusion

The issue decomposition is fundamentally sound with correct dependencies and reasonable atomicity. The three-issue structure aligns with the design's three phases. However, Issue 2 is underspecified and should be split into functional implementation (2a) and comprehensive testing (2b).

The main risk is that implementers may miss key details like minimal executor creation and state storage specifics, leading to rework or incomplete implementation. Adding these details will ensure smooth execution.

With the recommended improvements, this issue set will fully deliver the plan-based installation feature as designed.
