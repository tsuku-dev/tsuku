# Platform-Aware Recipes Design Document Review

## Executive Summary

The design document for platform-aware recipes is **well-structured and comprehensive** with a clear problem statement, thorough research, and fair option analysis. However, it is **incomplete** - the document ends abruptly after presenting options without a decision section, recommendations, or implementation plan.

**Key Findings:**
1. Problem statement is specific and well-motivated with concrete examples
2. Options analysis is balanced and fair across all four decision areas
3. Missing critical elements: decision/recommendation section, implementation plan, testing strategy
4. Some unstated assumptions about backwards compatibility need clarification
5. Decision drivers are comprehensive but could be prioritized

**Overall Assessment:** Strong foundation requiring completion before proceeding to implementation.

---

## 1. Problem Statement Specificity

### Strengths

The problem statement is **specific and well-articulated** with:

- **Concrete failure scenarios**: btop on macOS (404 error), hello-nix on macOS (nix-portable limitation)
- **Clear user impact**: Late failures with cryptic errors vs early, actionable feedback
- **Technical context**: References existing `Step.When` infrastructure and its limitations
- **Measurable outcomes**: Defines what "success" looks like (fail fast, clear UX, test efficiency, discovery support)

The scope section effectively delineates boundaries:
- In scope: Recipe-level constraints, CLI enforcement, test matrix, info command
- Out of scope: Cross-compilation, runtime detection, automatic fallbacks, website (deferred)

### Weaknesses

**Missing quantification:**
- No data on how often users encounter platform incompatibility errors
- No baseline for current error rate or support requests related to platform issues
- No metrics for measuring success (e.g., "reduce platform-related installation failures by X%")

**Assumption not explicitly stated:**
- The problem assumes users install recipes on incompatible platforms frequently enough to warrant this feature
- No discussion of whether recipe authors will actually populate platform metadata (adoption risk)

**Specificity gap:**
- "Cryptic errors" could be more specific - include actual error text users see
- Missing discussion of how this impacts CI costs (running tests on unsupported platforms)

### Recommendation for Improvement

Add a "User Impact" subsection with:
```markdown
### User Impact

**Current state:**
- Estimated XX% of installation failures are platform-related (based on telemetry/issues)
- Average user spends ~X minutes debugging platform incompatibility
- CI wastes ~X minutes per run testing unsupported platforms

**Target state:**
- Zero late-stage platform failures (caught at CLI entry)
- Platform incompatibility communicated in <1 second
- CI matrix automatically excludes unsupported combinations
```

---

## 2. Missing Alternatives

### Decision 1: Schema Design

**Missing Option:** Structured platform table with explicit support levels

```toml
[metadata.platforms]
linux.amd64 = "supported"
linux.arm64 = "supported"
darwin.amd64 = "unsupported"  # Explicit vs unknown
darwin.arm64 = "untested"     # Distinguish tested vs unknown
```

**Why consider:** Addresses the "never tested vs known to fail" gap mentioned in Option 1A cons. Allows recipe authors to be honest about testing coverage.

### Decision 2: Granularity

**Missing Option:** Wildcard support for partial tuples

```toml
[metadata]
supported_platforms = ["linux/*", "darwin/amd64"]  # Any Linux arch, only Intel Mac
```

**Why consider:** Balances verbosity (Option 2B con) with precision (Option 2A con). Common pattern in glob matching.

### Decision 3: Enforcement

**Missing Option:** Runtime skip with warning

Allow installation to proceed on unsupported platforms with:
- Big warning message
- `--force-platform` flag to acknowledge risk
- Mark installation as "unsupported" in state

**Why consider:** Some tools might work despite upstream not officially supporting the platform. Power users may want to try anyway.

### Decision 4: Error Messaging

**Missing Option:** Interactive recovery flow

```
Error: hello-nix is not available for darwin/arm64

What would you like to do?
  1. Search for alternatives (tsuku search hello)
  2. View platform support across all recipes (tsuku recipes --platform darwin/arm64)
  3. Cancel installation
```

**Why consider:** Turns error into action, reduces user friction, leverages existing commands.

**All options present are legitimate** - none are strawmen designed to fail.

---

## 3. Pros/Cons Balance

### Overall Assessment

The pros/cons analysis is **fair and comprehensive** across all options. Each option presents genuine trade-offs without obvious bias. However, there are areas where balance could be improved.

### Decision 1: Schema Design

**Balanced:**
- Option 1A (allowlist): Pros correctly highlight simplicity; cons acknowledge verbosity
- Option 1B (denylist): Pros note compactness; cons point out confusing inversion
- Option 1C (combined): Fairly presents flexibility vs complexity

**Improvement needed:**
- **Option 1A missing con:** Doesn't mention that explicit allowlists become stale (if new platforms added to Go, recipes don't automatically support them)
- **Option 1B missing pro:** Natural default (missing field = universal) means most recipes won't need this field at all
- **Option 1C missing con:** Risk of community fragmentation (some recipes use allowlist, others denylist - inconsistent user experience)

### Decision 2: Granularity

**Balanced:**
- Option 2A (separate): Fairly contrasts consistency with cartesian product issue
- Option 2B (tuples): Acknowledges precision vs verbosity trade-off
- Option 2C (hybrid): Presents flexibility vs complexity honestly

**Improvement needed:**
- **Option 2A missing pro:** Easier to extend incrementally (add `supported_libc` later without changing schema fundamentally)
- **Option 2B missing pro:** Matches Docker image naming (`golang:1.21-alpine-amd64`), familiar pattern
- **Option 2C missing analysis:** No discussion of migration path - if recipe starts with os/arch, how does it evolve to tuples?

### Decision 3: Enforcement

**Balanced:**
- All three options clearly present timing trade-offs
- Pros/cons are technically accurate

**Improvement needed:**
- **Option 3A missing con:** If validation happens only in install command, other code paths (future commands like `tsuku validate`, `tsuku plan-install`) must remember to add the check
- **Option 3B missing pro:** Supports executor reuse patterns (e.g., dry-run mode, plan generation)
- **Option 3C missing analysis:** Doesn't quantify "negligible" overhead - is this check O(1), O(n) in recipe steps?

### Decision 4: Error Messaging

**Balanced:**
- Options present clear UX trade-offs
- Pros/cons reflect implementation complexity honestly

**Improvement needed:**
- **Option 4A missing pro:** Fastest to implement, unblocks other work
- **Option 4B missing con:** Requires defining "similar" (keyword match? category tags? author-defined?) - significant scope increase
- **Option 4C missing pro:** Helps users decide if they should wait (if upstream issue is active) or switch tools (if upstream won't support platform)

---

## 4. Unstated Assumptions

### Critical Assumptions Needing Explication

**1. Backwards Compatibility Model**

- **Assumption:** Missing platform metadata = "supports all platforms"
- **Risk:** Recipe that never worked on Windows but lacked metadata will suddenly claim Windows support
- **Needs explicit statement:** "We assume existing recipes work on all platforms they've been tested on. Recipe authors must add constraints when updating recipes, not retroactively audit all existing recipes."

**2. Platform Detection Accuracy**

- **Assumption:** `runtime.GOOS` and `runtime.GOARCH` accurately represent the target platform
- **Risk:** Cross-compilation, Docker containers, WSL, emulation (Rosetta) can create mismatches
- **Needs discussion:** How do we handle darwin/arm64 running darwin/amd64 via Rosetta? Does recipe need to support darwin/amd64 explicitly?

**3. Recipe Author Adoption**

- **Assumption:** Recipe authors will populate platform metadata
- **Risk:** If optional and not enforced, metadata coverage may be sparse
- **Needs strategy:** Will there be CI checks for recipes with known platform-specific dependencies? How do we incentivize complete metadata?

**4. Testing Coverage**

- **Assumption:** CI can test all supported platform combinations
- **Risk:** If recipe claims `supported_platforms = ["linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"]`, can CI actually test all four?
- **Needs discussion:** What's the relationship between declared support and actual CI coverage? Can recipes claim support for untested platforms?

**5. Upstream Stability**

- **Assumption:** Upstream platform support doesn't change frequently
- **Risk:** If upstream adds Windows support, recipe metadata becomes stale
- **Needs guidance:** How often should recipe authors review platform metadata? Should there be automated checks against upstream releases?

**6. Monotonic Platform Support**

- **Assumption:** If a tool supports darwin/arm64, it supports darwin/amd64 (newer platforms superset older)
- **Risk:** Not always true (e.g., ARM-only tools, legacy platform deprecation)
- **Needs clarification:** No assumption should be made about platform relationships

### Secondary Assumptions

**7. Error Message Display Context**

- **Assumption:** Users run tsuku from interactive terminals where multi-line errors are readable
- **Consideration:** CI logs, scripts, non-interactive contexts may need different formatting

**8. Single Binary Model**

- **Assumption:** Recipes install single binaries per platform, not multi-arch bundles
- **Reality check:** Some upstream projects ship universal binaries (macOS), fat binaries, or multi-arch containers

**9. Platform as Only Constraint**

- **Assumption:** OS and arch are sufficient discriminators
- **Future-proofing:** libc version (glibc vs musl), kernel version, ABI compatibility may matter later

---

## 5. Strawman Detection

### Analysis

**No strawmen detected.** All options represent legitimate design choices with genuine trade-offs.

### Evidence

**Decision 1:** All three schema options (allowlist, denylist, combined) have real-world precedents:
- Allowlist: Nix's `meta.platforms`
- Denylist: Common in dependency exclusion patterns
- Combined: Cargo's feature flags (positive and negative)

**Decision 2:** All granularity options used in practice:
- Separate OS/arch: Existing `Step.When` in tsuku
- Tuples: Docker image tags, Go's GOOS/GOARCH
- Hybrid: Maven dependency scopes (compile/test/provided)

**Decision 3:** All enforcement points are valid:
- Preflight: Action validation pattern in tsuku
- Executor: Constructor validation in domain models
- Dual: Defense-in-depth security pattern

**Decision 4:** All error message strategies appear in production systems:
- Simple: Nix
- Alternatives: Homebrew's "similar formulae"
- Upstream links: Common in wrapper tools

**Balanced presentation:** Each option has 3-4 pros and 2-4 cons - no option is obviously superior or inferior.

---

## 6. Decision Drivers Completeness

### Provided Decision Drivers

The document lists six decision drivers:

1. **User experience** - Users should know upfront whether a tool supports their platform
2. **Existing infrastructure** - Leverage `Step.When` mechanism rather than duplicating logic
3. **Recipe simplicity** - Authors shouldn't need `when` clauses on every step
4. **Ecosystem integration** - Platform metadata must be consumable by CLI, tests, website
5. **Backwards compatibility** - Existing recipes without constraints should continue working
6. **Clear semantics** - Distinction between recipe-level and step-level conditions must be well-defined

### Assessment

**Strengths:**
- Covers multiple stakeholders (users, recipe authors, developers)
- Balances technical constraints (existing infrastructure) with UX
- Acknowledges ecosystem needs (CLI, tests, website)
- Includes compatibility requirement

**Missing Decision Drivers:**

**7. Implementation Cost/Complexity**
- Some options (e.g., 4B with alternatives suggestion) are significantly more complex
- Should explicitly weight "can we ship this in one iteration vs multiple?"

**8. Maintainability**
- Schema complexity affects long-term maintenance burden
- Should consider which option minimizes recipe author errors

**9. Debuggability**
- When things go wrong, how easy is it to diagnose?
- Simpler schemas = easier debugging for both users and developers

**10. Extensibility**
- Future needs: libc version, kernel version, ABI compatibility
- Which schema design best accommodates future platform dimensions?

**11. Ecosystem Consistency**
- Should align with broader Go ecosystem conventions
- GOOS/GOARCH tuples are Go-idiomatic; other patterns may confuse Go developers

**12. Migration Path**
- How do recipes evolve as upstream adds platform support?
- Should favor schemas that don't require rewrites when platforms change

**13. Error Recovery**
- Some options enable workarounds (Option 3 with --force flag)
- Should consider whether "hard no" or "warn and proceed" is better UX

**14. Documentation Burden**
- Complex schemas require more documentation
- Recipe authors are volunteers - minimize learning curve

### Prioritization

The decision drivers lack **prioritization**. Not all drivers are equally important.

**Recommended priority framework:**

**P0 (Must-have):**
- User experience (fail fast with clear errors)
- Backwards compatibility (existing recipes keep working)
- Recipe simplicity (low barrier for authors)

**P1 (Should-have):**
- Ecosystem integration (CLI/tests consume metadata)
- Clear semantics (recipe vs step level)
- Implementation cost (ship in reasonable timeframe)

**P2 (Nice-to-have):**
- Existing infrastructure reuse (optimize but don't block on it)
- Extensibility (don't block future needs)
- Maintainability (minimize long-term burden)

This prioritization would help resolve conflicts when drivers pull in different directions.

---

## 7. Critical Missing Sections

### 7.1 Decision and Rationale

**Missing:** The document ends after presenting options without a decision section.

**Impact:** Reader can't evaluate if the exploration led to a conclusion or is still open.

**Required section:**
```markdown
## Decision Outcome

### Chosen Options

**Decision 1 (Schema):** Option 1A - Allowlist via `supported_os`/`supported_arch`

**Rationale:** [Explain why chosen over alternatives based on decision drivers]

**Decision 2 (Granularity):** Option 2A - OS and Architecture Separate

**Rationale:** [...]

[Continue for all four decisions]

### Consequences

**Positive:**
- [What benefits does this decision provide?]

**Negative:**
- [What trade-offs are we accepting?]

**Neutral:**
- [What implementation details are implied?]
```

### 7.2 Implementation Plan

**Missing:** No roadmap for implementing the chosen design.

**Required section:**
```markdown
## Implementation Plan

### Phase 1: Schema and Validation
- Add `supported_os`/`supported_arch` fields to recipe schema
- Implement validation in recipe parser
- Add preflight checks in install command

### Phase 2: CLI Integration
- Update `tsuku info` to display platform support
- Enhance error messages with platform information
- Add test matrix filtering

### Phase 3: Ecosystem Rollout
- Document platform metadata in recipe authoring guide
- Add platform badges to existing high-priority recipes
- CI validation for platform metadata completeness

### Success Metrics
- [How will we know this succeeded?]
```

### 7.3 Testing Strategy

**Missing:** No discussion of how to test this feature.

**Required section:**
```markdown
## Testing Strategy

### Unit Tests
- Recipe parser correctly reads platform metadata
- Validation logic correctly determines support
- Error messages format properly

### Integration Tests
- Install command rejects unsupported platforms
- info command displays platforms correctly
- Test matrix skips unsupported combinations

### Recipe Validation Tests
- CI validates platform metadata syntax
- CI ensures recipes with known platform deps have metadata
- CI catches contradictions (recipe says supported but action doesn't work)
```

### 7.4 Migration Guide

**Missing:** No guidance for recipe authors on adopting platform metadata.

**Required section:**
```markdown
## Migration Guide for Recipe Authors

### When to Add Platform Metadata

Add `supported_os`/`supported_arch` when:
- Upstream project explicitly doesn't support certain platforms
- Recipe uses actions with platform constraints (e.g., nix-portable)
- CI tests only cover subset of platforms

### When NOT to Add Platform Metadata

Omit metadata when:
- Recipe works on all Go-supported platforms
- No platform-specific actions or dependencies

### Examples

[Concrete before/after examples for common scenarios]
```

### 7.5 Alternatives Considered and Rejected

**Missing:** Options were presented but no explicit rejection reasoning.

**Recommendation:** After decision section, add:
```markdown
## Alternatives Considered and Rejected

### Why Not Option 1B (Denylist)?

While compact for "all except X" cases, rejected because:
- Inverted logic confusing for recipe authors
- Doesn't help with "Linux-only" tools (most common case)
- Harder to display "supported platforms" in UI

### Why Not Option 2B (Tuples)?

[...continue for all non-chosen options...]
```

---

## 8. Research Quality

### Strengths

**Existing Patterns Section (lines 56-125):**
- Excellent survey of tsuku's current platform handling mechanisms
- Identifies five relevant patterns with code references
- Provides line number citations for code locations

**Comparative Analysis (lines 87-100):**
- Compares Homebrew, asdf, Nix approaches
- Identifies key takeaways applicable to tsuku
- Balanced - doesn't blindly copy competitors

**Specifications Section (lines 102-113):**
- Correctly identifies TOML and Go runtime as constraints
- Provides authoritative references

### Weaknesses

**No User Research:**
- No data on how often users hit platform errors
- No quotes from actual user feedback or issues
- No analysis of existing GitHub issues for platform-related failures

**No Upstream Project Survey:**
- Claims "btop on macOS fails with 404" but doesn't link to btop's platform support docs
- Could strengthen argument by listing X recipes with known platform issues

**No Performance Analysis:**
- Enforcement options have different performance characteristics
- No discussion of whether platform checks add meaningful latency

### Recommendation

Add subsection:
```markdown
### Empirical Evidence

**Issue Analysis:**
- #XXX: btop fails on macOS with 404 error (opened YYYY-MM-DD, 5 comments)
- #YYY: hello-nix unclear platform support (opened YYYY-MM-DD)
- #ZZZ: [...]

**Upstream Platform Support:**
- btop: Linux-only per [upstream README](link)
- hello-nix: nix-portable is Linux-only per [repo](link)
- [Continue for 5-10 recipes with known platform constraints]

**Telemetry Data (if available):**
- XX% of failed installations are HTTP 404 (likely platform mismatches)
- YY% of users are on darwin/arm64 (M1/M2 Macs - growing segment)
```

---

## 9. Specific Recommendations

### Immediate Actions (Required Before Implementation)

**1. Complete Decision Section**
- Choose one option for each of the four decisions
- Provide rationale based on decision drivers
- Document accepted trade-offs

**2. Add Implementation Plan**
- Break down into phases
- Estimate effort for each phase
- Identify dependencies between phases

**3. Clarify Assumptions**
- Add "Assumptions" subsection to problem statement
- Explicitly state backwards compatibility model
- Document platform detection edge cases (Rosetta, WSL, containers)

**4. Prioritize Decision Drivers**
- Mark P0/P1/P2 priority levels
- Use prioritization to resolve conflicts in decision rationale

### Recommended Enhancements

**5. Add Missing Alternatives**
- Decision 1: Structured table with support levels (supported/unsupported/untested)
- Decision 2: Wildcard support for partial tuples (`linux/*`)
- Decision 3: Runtime skip with `--force-platform` flag
- Decision 4: Interactive recovery flow

**6. Balance Pros/Cons**
- Add missing pros/cons identified in section 3
- Ensure each option has balanced representation

**7. Empirical Evidence**
- Add issue analysis (existing GitHub issues about platform failures)
- Survey upstream projects for platform support documentation
- Include telemetry data if available

**8. Add Testing Strategy**
- Unit test coverage plan
- Integration test scenarios
- CI validation requirements

### Optional Improvements

**9. Add Migration Guide**
- When to add platform metadata
- Examples for common scenarios
- Migration checklist for recipe authors

**10. Add Rejected Alternatives Section**
- Explain why non-chosen options were rejected
- Helps future readers understand decision context

**11. Add Extensibility Analysis**
- How does chosen design accommodate future needs (libc, ABI, etc.)?
- Migration path if requirements change

---

## 10. Overall Quality Assessment

### Document Structure: Excellent

- Clear organization with decision groupings
- Logical flow from problem → context → options
- Good use of examples and code snippets

### Research Depth: Very Good

- Thorough review of existing tsuku infrastructure
- Comparative analysis of other package managers
- Proper citations and references

### Option Analysis: Good

- Fair presentation of trade-offs
- No strawmen
- Some minor gaps in pros/cons balance

### Completeness: Poor

- Missing decision section (critical gap)
- Missing implementation plan
- Missing testing strategy

### Actionability: Moderate

- Provides enough context for informed decision
- Lacks concrete next steps
- No assignment of implementation work

---

## 11. Recommendations Summary

### Critical (Must Address Before Implementation)

1. **Add decision section** with chosen options and rationale for each of the four decisions
2. **Add implementation plan** with phases, effort estimates, and dependencies
3. **Clarify backwards compatibility model** - explicit statement about missing metadata = universal support
4. **Prioritize decision drivers** (P0/P1/P2) to guide decision-making

### Important (Should Address Before Finalization)

5. **Add testing strategy** covering unit, integration, and CI validation
6. **Balance pros/cons** with missing items identified in section 3
7. **Add "Assumptions" subsection** documenting platform detection edge cases
8. **Add empirical evidence** (issue analysis, upstream surveys, telemetry data)

### Nice-to-Have (Consider for Completeness)

9. **Evaluate missing alternatives** (structured tables, wildcards, --force-platform, interactive recovery)
10. **Add migration guide** for recipe authors
11. **Add rejected alternatives section** explaining why non-chosen options were excluded
12. **Add extensibility analysis** for future platform dimensions (libc, ABI, etc.)

---

## Conclusion

This is a **high-quality exploration** with solid research and fair option analysis. The problem is well-defined, options are legitimate, and no strawmen were detected.

**Blocker for proceeding:** The document is **incomplete** - it ends abruptly without decisions, recommendations, or implementation plan. This must be completed before implementation can begin.

**Recommendation:** Author should:
1. Choose one option for each decision based on prioritized decision drivers
2. Add decision outcome section with rationale
3. Add implementation plan with phases
4. Address critical gaps (assumptions, testing, backwards compatibility)
5. Proceed to implementation planning phase

**Overall grade:** B+ for exploration quality, incomplete for decision-making readiness.
