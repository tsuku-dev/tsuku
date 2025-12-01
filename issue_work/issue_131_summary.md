# Issue 131 Summary

## What Was Implemented

CPAN builder that generates recipes for CPAN distributions, enabling `tsuku create App-Ack --from cpan` to generate working recipes automatically.

## Changes Made
- `internal/builders/cpan.go`: New CPAN builder implementing Builder interface
- `internal/builders/cpan_test.go`: Comprehensive tests with mock HTTP server

## Key Decisions
- **Module name normalization**: Accept both App::Ack (module) and App-Ack (distribution) formats, normalize to distribution format internally
- **Executable inference**: Transform distribution name to executable (App-Ack -> ack), warn when uncertain (non-App distributions)
- **Version source format**: Use `metacpan:<distribution>` format consistent with existing version provider

## Trade-offs Accepted
- **Heuristic executable naming**: Cannot query MetaCPAN for actual executables; use naming conventions instead with warnings when uncertain
- **No Makefile.PL parsing**: More accurate but requires downloading tarballs; rejected for simplicity and speed

## Test Coverage
- New tests added: 16 tests across 4 test functions
- Coverage: 89.0% for builders package, >80% for all cpan.go functions

## Known Limitations
- Executable name inference may be wrong for distributions that don't follow standard naming conventions
- Non-App distributions (like Perl-Critic) will have warning about uncertain executable name

## Future Improvements
- Could fetch and parse META.json for more accurate executable discovery
- Could query MetaCPAN file search API for scripts in bin/ or script/ directories
