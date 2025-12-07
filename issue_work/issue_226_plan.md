# Issue 226 Implementation Plan

## Summary
Add `--all` flag to `tsuku list` command that includes libraries (from $TSUKU_HOME/libs/) in the output, marked with a `[lib]` indicator.

## Approach
Libraries are already hidden by default since they're in a separate `libs/` directory. We need to add:
1. A method to list installed libraries
2. A flag to include libraries in the list command
3. Visual indicator for libraries in the output

### Alternatives Considered
- `--include-libraries` flag: Rejected in favor of `--all` which is shorter and more intuitive
- Showing libraries inline with tools: Rejected - will show libraries in a separate section for clarity

## Files to Modify
- `internal/install/library.go` - Add ListLibraries method
- `cmd/tsuku/list.go` - Add `--all` flag and library display logic

## Files to Create
None

## Implementation Steps
- [ ] Add ListLibraries method to Manager
- [ ] Add `--all` flag to list command
- [ ] Display libraries with `[lib]` marker when `--all` is used
- [ ] Add unit test for ListLibraries

## Testing Strategy
- Unit tests: Test ListLibraries returns libraries from libs/ directory
- Manual verification: `tsuku list` vs `tsuku list --all`

## Risks and Mitigations
- Risk: Empty libs directory could cause issues
- Mitigation: Return empty slice if libs/ doesn't exist

## Success Criteria
- [ ] `tsuku list` hides libraries by default
- [ ] `tsuku list --all` shows libraries marked with [lib]
- [ ] Unit test for ListLibraries
- [ ] All tests pass

## Open Questions
None
