# Issue 421 Summary

## What Was Implemented
CLI verbosity control via `--quiet`, `--verbose`, `--debug` flags and corresponding `TSUKU_QUIET`, `TSUKU_VERBOSE`, `TSUKU_DEBUG` environment variables. The global logger from `internal/log` is initialized at startup based on these settings.

## Changes Made
- `cmd/tsuku/main.go`: Added verboseFlag and debugFlag variables, registered as persistent flags, added PersistentPreRun hook to initialize logger, implemented determineLogLevel() for flag/env precedence, added isTruthy() helper
- `cmd/tsuku/main_test.go`: New file with unit tests for isTruthy() and determineLogLevel()

## Key Decisions
- **Flag precedence**: Flags override environment variables, following CLI conventions from gh, kubectl, docker
- **Level priority order**: debug > verbose > quiet > default, so enabling debug mode takes precedence even if quiet is also set
- **Truthy values**: Accept "1", "true", "yes", "on" (case-insensitive) for environment variables

## Trade-offs Accepted
- **No mutual exclusivity enforcement**: Multiple flags can be set; priority order determines outcome rather than error

## Test Coverage
- New tests added: 2 test functions covering 27 test cases
- Unit tests verify all flag/env combinations and precedence

## Known Limitations
- Logger is only initialized for commands, not for --help or --version (those don't run PersistentPreRun)

## Future Improvements
- Add structured logging calls throughout the codebase to emit debug/info/warn messages
