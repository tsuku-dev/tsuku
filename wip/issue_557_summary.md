# Issue 557 & 558 Summary

## What Was Implemented

Added readline recipe with ncurses dependency and sqlite recipe that builds from source with readline support. This validates the complete dependency chain: sqlite → readline → ncurses, demonstrating that tsuku can auto-provision library dependencies for source builds using configure_make.

## Changes Made

- `internal/recipe/recipes/r/readline.toml`: Added ncurses dependency declaration and Linux shared library files (.so)
- `internal/recipe/recipes/s/sqlite.toml`: Created new recipe that builds sqlite from source with readline support
- `internal/recipe/recipes/n/ncurses.toml`: Fixed bug - removed hardcoded checksum (should be computed dynamically)
- `test/scripts/test-readline-provisioning.sh`: Created integration test validating readline provisioning in clean environment
- `test/scripts/verify-tool.sh`: Added verify_readline and verify_sqlite functions
- `.github/workflows/build-essentials.yml`: Added test-sqlite job to CI matrix (tests on 3 platforms)
- `docs/DESIGN-dependency-provisioning.md`: Updated mermaid diagrams to mark #557 and #558 as done

## Key Decisions

- **Use homebrew bottles for readline**: Faster than building from source, follows pattern from other library recipes
- **Build sqlite from source**: Required to validate configure_make action with library dependencies (issue #558 explicitly requires source build)
- **Hardcode sqlite version in URL**: SQLite uses non-standard version format (3.51.1 → 3510100) incompatible with {version} substitution. Filed issue #660 to track this limitation for future improvement
- **Include comprehensive test script**: Following pattern from PR #659 (cmake/ninja), created Docker-based test that validates provisioning in clean Ubuntu 22.04 without system readline/ncurses
- **Add dedicated CI job for sqlite**: Rather than adding to existing test-configure-make job, created separate test-sqlite job to highlight the dependency chain validation

## Trade-offs Accepted

- **Manual URL maintenance for sqlite**: Until issue #660 is resolved, sqlite recipe URL must be manually updated when version changes. This is acceptable because:
  - SQLite releases are infrequent
  - Version provider still works (pulls version from Homebrew formula)
  - Only the download URL needs manual updates
  - Clear TODO comment references the issue

- **No direct test of readline binary**: readline is a library with no standalone binary. Validated indirectly through sqlite usage, which is more realistic anyway

## Test Coverage

- **New tests added**:
  - test-readline-provisioning.sh (Docker-based integration test)
  - verify_readline function in verify-tool.sh
  - verify_sqlite function in verify-tool.sh
  - test-sqlite CI job (runs on 3 platforms)
- **Coverage change**: No unit test changes (recipe-only modifications)
- **Integration coverage**: Complete dependency chain validated end-to-end in clean environment

## Known Limitations

- SQLite recipe requires manual URL updates when version changes (tracked in #660)
- readline recipe uses homebrew bottles which may have platform-specific limitations (similar to other homebrew-based recipes)

## Future Improvements

- Implement version transformation functions (issue #660) to support SQLite's non-standard version format
- Consider adding more library recipes that build from source to further validate the dependency resolution system
