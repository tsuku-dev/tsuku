# Issue 405 Implementation Plan

## Summary

Add `tsuku plan show <tool>` command to display stored installation plans in human-readable format.

## Approach

Create a new `plan.go` file with parent `planCmd` and child `planShowCmd`. Follow the `cache.go` subcommand pattern. Read plan from state.json via StateManager and format for display.

### Alternatives Considered

- **Single command `tsuku plan <tool>`**: Rejected - leaves room for future plan subcommands (export, replay)
- **JSON-only output**: Rejected - issue requires human-readable format (with optional --json flag)

## Files to Create

- `cmd/tsuku/plan.go` - Main plan command with show subcommand
- `cmd/tsuku/plan_test.go` - Unit tests

## Files to Modify

- `cmd/tsuku/main.go` - Register planCmd with rootCmd

## Implementation Steps

- [ ] Create `cmd/tsuku/plan.go` with:
  - [ ] Parent `planCmd` (container for subcommands)
  - [ ] Child `planShowCmd` with tool argument
  - [ ] Load state via StateManager
  - [ ] Format plan output (tool, version, platform, steps)
  - [ ] Highlight non-evaluable steps
  - [ ] Error handling for: tool not installed, no plan stored
  - [ ] --json flag for JSON output
- [ ] Register `planCmd` in `cmd/tsuku/main.go`
- [ ] Add unit tests
- [ ] Run `go vet`, `go test`, and `go build` to verify

## Output Format Design

```
Plan for gh@2.40.0

Platform: linux/amd64
Generated: 2024-12-12 21:30:00 UTC
Recipe:   registry (hash: abc123...)

Steps:
  1. [download_archive] https://github.com/...
     Checksum: sha256:deadbeef...
     Size: 12.5 MB
  2. [extract] format=tar.gz
  3. [install_binaries] binaries=gh
  4. [run_command] (non-evaluable)
     command: ./configure && make
```

## Testing Strategy

- Unit tests: Verify output formatting
- Unit tests: Verify error messages for missing tool/plan
- Manual verification: `tsuku install gh && tsuku plan show gh`

## Success Criteria

- [ ] `tsuku plan show <tool>` displays formatted plan
- [ ] Non-evaluable steps clearly marked
- [ ] Clear error if tool not installed
- [ ] Clear error if tool has no plan
- [ ] `--json` flag outputs raw JSON
- [ ] Help text documents usage
- [ ] All tests pass, no lint errors

## Open Questions

None.
