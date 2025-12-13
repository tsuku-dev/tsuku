# Issue 421 Implementation Plan

## Summary
Wire up `--quiet`, `--verbose`, `--debug` CLI flags and corresponding environment variables (`TSUKU_QUIET`, `TSUKU_VERBOSE`, `TSUKU_DEBUG`) to initialize the global logger from `internal/log` at startup.

## Approach
Add persistent flags to the root command, check environment variables as fallbacks, determine the appropriate log level, create a CLIHandler with that level, and call `log.SetDefault()`. Display a warning banner when debug mode is enabled.

### Alternatives Considered
- **Numeric verbosity like kubectl's `-v=N`**: Rejected because the design doc explicitly chose named flags for simplicity
- **Environment-only configuration**: Rejected because flags should take precedence for immediate CLI control

## Files to Modify
- `cmd/tsuku/main.go` - Add verbose/debug flags, environment variable checks, logger initialization, debug banner

## Files to Create
- None (all changes in existing file)

## Implementation Steps
- [ ] Add `verboseFlag` and `debugFlag` variables alongside existing `quietFlag`
- [ ] Update `init()` to register `--verbose` and `--debug` as persistent flags
- [ ] Add `PersistentPreRun` to root command for logger initialization before any command runs
- [ ] Implement environment variable checks (`TSUKU_QUIET`, `TSUKU_VERBOSE`, `TSUKU_DEBUG`) with flag precedence
- [ ] Create and set the global logger using `log.NewCLIHandler` and `log.SetDefault`
- [ ] Display debug warning banner when debug mode is enabled
- [ ] Write unit tests for flag/env var precedence logic

## Testing Strategy
- Unit tests: Test helper function for log level determination (flag/env precedence)
- Manual verification: Run `tsuku --debug install` and verify banner appears, debug output shows

## Risks and Mitigations
- **Flag conflicts**: The existing `quietFlag` is already defined; we need to handle mutual exclusivity (debug/verbose override quiet)
- **Env var parsing**: Use `os.Getenv` and check for truthy values ("1", "true")

## Success Criteria
- [ ] `tsuku --quiet` shows only errors
- [ ] `tsuku --verbose` shows INFO level logs
- [ ] `tsuku --debug` shows DEBUG level logs with timestamps and source
- [ ] `TSUKU_DEBUG=1 tsuku` works same as `tsuku --debug`
- [ ] Flags override environment variables
- [ ] Debug banner displays when debug mode is active
- [ ] Help text documents all flags

## Open Questions
None - requirements are clear from issue and design doc.
