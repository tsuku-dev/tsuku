# Issue 440 Implementation Plan

## Summary

Update the plan generator to detect composite actions (those implementing `Decomposable`), call `DecomposeToPrimitives()`, and record only primitive actions in generated plans. Bump plan format version to 2 with backward compatibility for v1.

## Approach

The plan generator currently records steps directly from the recipe. The change integrates the existing `actions.DecomposeToPrimitives()` function into the plan generation flow. For each step in the recipe, if the action is decomposable (implements `Decomposable`), we call `DecomposeToPrimitives()` and add the resulting primitive steps to the plan.

### Alternatives Considered
- **Decompose in executor**: Run decomposition at execution time. Rejected because it defeats the purpose of deterministic plans - the plan would still contain composite actions.
- **Rewrite resolveStep entirely**: More invasive change. Rejected in favor of minimal modification that integrates with existing flow.

## Files to Modify

- `internal/executor/plan.go` - Bump `PlanFormatVersion` to 2, update `ActionEvaluability` to remove composites, add v1 compatibility
- `internal/executor/plan_generator.go` - Integrate `DecomposeToPrimitives()` call for composite actions
- `internal/executor/plan_test.go` - Update test expectations for v2 format
- `internal/actions/decomposable.go` - Add `IsDecomposable()` helper function

## Files to Create

None

## Implementation Steps

- [ ] Add `IsDecomposable()` helper function to actions package
- [ ] Update `PlanFormatVersion` constant to 2 in plan.go
- [ ] Remove composite actions from `ActionEvaluability` map (primitives only)
- [ ] Modify `GeneratePlan()` to create `EvalContext` for decomposition
- [ ] Modify `resolveStep()` to detect and decompose composite actions
- [ ] Add backward compatibility check for reading v1 plans
- [ ] Update existing tests for v2 format
- [ ] Add integration test: github_archive recipe produces primitive-only plan

## Testing Strategy

- **Unit tests**: Test `IsDecomposable()` returns correct values for primitives vs composites
- **Integration tests**:
  - Generate plan for recipe with `github_archive` action, verify only primitives in output
  - Verify format_version is 2 in generated plans
  - Test backward compatibility: can still unmarshal v1 plans

## Risks and Mitigations

- **Performance**: Decomposition involves network calls (download for checksum). Already mitigated by existing `PreDownloader` with caching.
- **Breaking changes**: Old plans with composite actions. Mitigated by version check and allowing v1 plans to be re-evaluated.

## Success Criteria

- [ ] Plans generated with composite recipes contain only primitive actions
- [ ] `PlanFormatVersion` is 2
- [ ] All primitives from `actions.Primitives()` are in `ActionEvaluability` as true
- [ ] Tests pass including existing and new integration tests
- [ ] Can still read v1 format plans (backward compatibility)
