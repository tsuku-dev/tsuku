# Issue 442 Implementation Plan

## Summary

Add `Deterministic` boolean field to `ResolvedStep` (in executor) and `PlanStep` (in install) structs, plus plan-level `Deterministic` field on `InstallationPlan` and `Plan`. Tier 1 primitives are deterministic, ecosystem primitives are not.

## Approach

The design doc specifies that ecosystem primitives have residual non-determinism (compiler versions, native extensions) and should be explicitly marked. This requires:

1. Adding `Deterministic` field to step structs
2. Adding `Deterministic` field to plan structs (false if any step is non-deterministic)
3. Classifying actions as deterministic or not during plan generation
4. Displaying determinism status in `tsuku plan show`

### Determinism Classification

**Deterministic (Tier 1 primitives):**
- download, extract, chmod, install_binaries
- set_env, set_rpath, link_dependencies, install_libraries

**Non-deterministic (Ecosystem primitives):**
- go_build, cargo_build, npm_exec, pip_install, gem_exec

### Alternatives Considered

- **Per-step only, no plan-level flag**: Simpler but harder for users to quickly see overall determinism status. Design doc specifically asks for both.

- **Rename `Evaluable` to `Deterministic`**: Would change meaning - Evaluable means "can be captured in plan", Deterministic means "will reproduce exactly". A non-evaluable step can't be in a plan at all, while a non-deterministic step can be in a plan but may vary.

## Files to Modify

- `internal/executor/plan.go` - Add `Deterministic` field to `ResolvedStep` and `InstallationPlan`
- `internal/executor/plan_generator.go` - Set `Deterministic` during step resolution
- `internal/install/state.go` - Add `Deterministic` field to `PlanStep` and `Plan`
- `internal/actions/decomposable.go` - Add `IsDeterministic()` helper function
- `internal/actions/decomposable_test.go` - Fix pre-existing test failure + add determinism tests
- `cmd/tsuku/plan.go` - Display determinism status in `plan show` output
- `internal/executor/plan_test.go` - Add tests for deterministic flag

## Implementation Steps

- [ ] Add `IsDeterministic(action string) bool` to actions package
- [ ] Fix pre-existing `TestPrimitives` test failure (expected 12, now 13 primitives)
- [ ] Add `Deterministic` field to `ResolvedStep` in executor/plan.go
- [ ] Add `Deterministic` field to `InstallationPlan` in executor/plan.go
- [ ] Update plan generator to set `Deterministic` on steps and compute plan-level flag
- [ ] Add `Deterministic` field to `PlanStep` in install/state.go
- [ ] Add `Deterministic` field to `Plan` in install/state.go
- [ ] Update `printPlanHuman()` in cmd/tsuku/plan.go to show determinism status
- [ ] Add unit tests for determinism flag propagation

## Testing Strategy

- **Unit tests**:
  - Test `IsDeterministic()` returns correct values for all primitive types
  - Test plan generation sets correct deterministic flags on steps
  - Test plan-level deterministic flag computation (all deterministic = true, any non-deterministic = false)
  - Test JSON roundtrip preserves deterministic field

- **Manual verification**:
  - `tsuku eval` on a recipe with only download/extract shows `deterministic: true`
  - `tsuku eval` on a recipe with go_build shows `deterministic: false`
  - `tsuku plan show` displays determinism status

## Success Criteria

- [ ] `Deterministic` field added to step and plan structs
- [ ] Tier 1 primitives marked as deterministic
- [ ] Ecosystem primitives marked as non-deterministic
- [ ] Plan-level `Deterministic` is false if any step is non-deterministic
- [ ] `tsuku plan show` displays determinism status
- [ ] All tests pass (including fix for pre-existing failure)
- [ ] JSON schema documented with new field

## Open Questions

None - requirements are clear from the design doc.
