# Issue 241 Implementation Plan

## Summary

Enhance `tsuku remove` to warn (not error) when removing a tool that other tools depend on, with `--force` flag to proceed anyway.

## Approach

Modify the existing `RequiredBy` check in `cmd/tsuku/remove.go` to:
1. Show a warning instead of error when dependents exist
2. Add `--force` flag to proceed despite warnings
3. For hidden/auto-installed deps, allow removal with dependents (use --force implicitly)

### Alternatives Considered
- Scan all tools' InstallDependencies/RuntimeDependencies: Not needed since RequiredBy already tracks reverse dependencies
- Interactive prompt: Not chosen - flags are clearer for CLI automation

## Files to Modify
- `cmd/tsuku/remove.go` - Add --force flag and change error to warning

## Files to Create
None

## Implementation Steps
- [x] Add --force flag to remove command
- [x] Change RequiredBy check from error to warning with --force bypass
- [x] Test the warning and force flag behavior

## Testing Strategy
- Manual verification with mocked dependencies
- Unit tests via existing test infrastructure

## Risks and Mitigations
- **Breaking change for scripts**: Using warning instead of error could break scripts that expect error exit. Mitigated by keeping non-zero exit code without --force.

## Success Criteria
- [x] Uninstall checks if any installed tool depends on target
- [x] Warning lists dependent tools
- [x] User can proceed with `--force`
- [x] Hidden deps can be removed with --force

## Open Questions
None
