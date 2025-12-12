# Issue 406 Implementation Plan

## Summary

Add `tsuku plan export <tool>` command to export stored installation plans as standalone JSON files.

## Approach

Add `planExportCmd` as a new subcommand of `planCmd` in `plan.go`. Reuse existing plan retrieval logic from `runPlanShow`. Export using the same JSON format as `tsuku eval` (via `printJSON`).

### Alternatives Considered

- **Separate file `export.go`**: Rejected - command is small and fits naturally with other plan subcommands
- **Use `plan show --json > file.json`**: Rejected - issue requires default filename generation and `-` for stdout

## Files to Modify

- `cmd/tsuku/plan.go` - Add `planExportCmd` subcommand with export logic

## Files to Create

None - adding to existing plan.go

## Implementation Steps

- [ ] Add `planExportCmd` with flags (`--output`/`-o`)
- [ ] Add helper function `getPlanForTool` to extract common plan retrieval logic
- [ ] Implement `runPlanExport` with:
  - [ ] Default filename: `<tool>-<version>-<os>-<arch>.plan.json`
  - [ ] `-` support for stdout
  - [ ] Custom output path via `--output`
- [ ] Add unit tests for filename generation
- [ ] Run `go vet`, `go test`, and `go build` to verify

## Testing Strategy

- Unit tests: Test default filename generation
- Manual verification: `tsuku install gh && tsuku plan export gh`

## Success Criteria

- [ ] `tsuku plan export <tool>` exports plan to default filename
- [ ] Default filename is `<tool>-<version>-<os>-<arch>.plan.json`
- [ ] `--output` / `-o` allows custom output path
- [ ] `-` outputs to stdout (enables piping)
- [ ] JSON format matches `tsuku eval` output
- [ ] Clear error if tool not installed
- [ ] Clear error if tool has no stored plan
- [ ] Help text documents usage
- [ ] All tests pass, no lint errors

## Open Questions

None.
