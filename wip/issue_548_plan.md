# Issue 548 Implementation Plan

## Summary
Create a pkg-config recipe using the homebrew action to provision pkg-config for library discovery. Add build-from-source test to the build tools test matrix to validate it works correctly.

## Approach
Following the established pattern used by make.toml and other homebrew-based build tools. The recipe will:
1. Use homebrew version source to detect latest version
2. Use homebrew action to install the bottle
3. Install binaries in directory mode
4. Include verify section to test `pkg-config --version`
5. Add to test matrix with a build-from-source test

### Alternatives Considered
- **Build from source**: More complex, unnecessary since homebrew bottles are available for all platforms
- **System pkg-config**: Not self-contained, violates tsuku philosophy

## Files to Modify
- `test-matrix.json` - Add homebrew_pkg-config test to validate installation and a build-from-source test

## Files to Create
- `internal/recipe/recipes/p/pkg-config.toml` - Recipe definition

## Implementation Steps
- [x] Create pkg-config.toml recipe with homebrew action
- [x] Add basic installation test to build-essentials.yml
- [x] Add build-from-source test that uses pkg-config (gdbm-source already uses configure_make which depends on pkg-config)
- [x] Test locally that recipe works on current platform
- [x] Verify test matrix entries are correctly formatted

## Testing Strategy
- Unit tests: Existing recipe validation tests will cover the new recipe
- Integration: Test matrix will validate on all 4 platforms in CI
- Manual verification: Install pkg-config locally and test `pkg-config --version`
- Build test: A configure-based build that uses pkg-config to find libraries

## Risks and Mitigations
- **Risk**: Homebrew formula name might differ from "pkg-config"
  - **Mitigation**: Verify formula name exists in homebrew/core
- **Risk**: Binary path might be non-standard
  - **Mitigation**: Follow pattern from make.toml which handles both bin/make and bin/gmake

## Success Criteria
- [ ] pkg-config recipe installs successfully
- [ ] `pkg-config --version` works from relocated path
- [ ] CI test passes on all platforms
- [ ] Build-from-source test validates pkg-config functionality

## Open Questions
None - the pattern is well-established from existing recipes
