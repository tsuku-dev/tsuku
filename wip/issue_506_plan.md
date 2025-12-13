# Issue 506 Implementation Plan

## Summary

Add `loadPlanFromSource()` and `validateExternalPlan()` functions as utilities for loading and validating installation plans from files or stdin. These utilities are the foundation for the `--plan` flag on the install command.

## Approach

Follow the design document (docs/DESIGN-plan-based-installation.md) exactly. Create a new file `cmd/tsuku/plan_utils.go` to house the utilities, keeping them close to where they'll be used (the install command) while separating concerns cleanly.

### Alternatives Considered
- **Put utilities in `internal/executor/`**: Would require import from cmd package back to internal; utilities are CLI-specific (stdin handling) so better in cmd
- **Inline in install.go**: Would make install.go too large and harder to test independently

## Files to Create
- `cmd/tsuku/plan_utils.go` - `loadPlanFromSource()` and `validateExternalPlan()` functions
- `cmd/tsuku/plan_utils_test.go` - Unit tests for both functions

## Implementation Steps
- [x] Create `cmd/tsuku/plan_utils.go` with `loadPlanFromSource()` function
- [x] Add `validateExternalPlan()` function that wraps `executor.ValidatePlan()`
- [x] Create unit tests for file loading scenarios
- [x] Create unit tests for stdin loading scenarios
- [x] Create unit tests for validation scenarios (platform, tool name)
- [x] Run `go vet`, `go test`, and `go build` to verify

## Testing Strategy
- Unit tests for `loadPlanFromSource()`:
  - Read from valid file path
  - Read from stdin (using `-`)
  - Handle file not found error
  - Handle invalid JSON with helpful error message
  - Handle stdin parse error with debugging hint
- Unit tests for `validateExternalPlan()`:
  - Valid plan passes validation
  - Platform mismatch (OS) produces clear error
  - Platform mismatch (Arch) produces clear error
  - Tool name mismatch produces clear error
  - Tool name empty (optional) passes validation
  - Structural validation failures from `ValidatePlan()` are propagated

## Risks and Mitigations
- **Risk**: Stdin tests can be tricky to write
- **Mitigation**: Use `io.Reader` abstraction and inject test readers

## Success Criteria
- [x] `loadPlanFromSource()` handles file and stdin paths correctly
- [x] `validateExternalPlan()` calls `ValidatePlan()` and adds platform/tool checks
- [x] Unit tests cover all scenarios from acceptance criteria
- [x] `go vet ./...` passes
- [x] `go test ./...` passes (excluding pre-existing LLM failures)
- [x] `go build ./cmd/tsuku` succeeds

## Open Questions
None - design document provides clear specifications.
