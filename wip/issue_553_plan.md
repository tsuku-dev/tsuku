# Issue 553 Implementation Plan

## Summary
Create ncurses recipe that builds from source using configure_make with setup_build_env to validate pkg-config integration and dependency provisioning.

## Approach
Build ncurses from source tarball using autotools. The recipe will:
1. Download ncurses source tarball
2. Extract it
3. Run setup_build_env to explicitly configure the build environment
4. Run configure_make to build and install
5. Install binaries and libraries
6. Verify with pkg-config

This validates the complete build environment provisioning chain:
- setup_build_env configures PKG_CONFIG_PATH, CPPFLAGS, LDFLAGS
- configure_make uses these to build ncurses
- pkg-config can query the installed ncurses

### Alternatives Considered
- **Use homebrew action**: Simpler but doesn't validate configure_make or pkg-config integration. Issue specifically asks for configure_make to validate the build environment setup.
- **Build without setup_build_env**: configure_make already calls buildAutotoolsEnv() internally, but the acceptance criteria explicitly mentions setup_build_env as a separate step to demonstrate explicit environment configuration.

## Files to Create
- `internal/recipe/recipes/n/ncurses.toml` - ncurses recipe using configure_make

## Implementation Steps
- [x] Download ncurses source tarball (use download_file or github_archive)
- [ ] Extract tarball
- [ ] Run setup_build_env
- [ ] Run configure_make with appropriate parameters
- [ ] Install binaries (ncurses executables)
- [ ] Add verify section using pkg-config

## Testing Strategy
- Unit tests: Recipe validation (schema, required fields)
- Manual verification:
  - Install ncurses: `./tsuku install ncurses`
  - Check pkg-config: `pkg-config --libs ncurses` should return correct paths
  - Verify version: `pkg-config --modversion ncurses`
- CI validation: Test on all 4 platforms (Linux x86_64, Linux arm64, macOS x86_64, macOS arm64)

## Risks and Mitigations
- **Risk**: ncurses source tarball URL may change or become unavailable
  - **Mitigation**: Use a stable GNU mirror (ftpmirror.gnu.org) and specify exact version

- **Risk**: configure_make may fail on some platforms
  - **Mitigation**: Use minimal configure args, rely on autotools defaults

- **Risk**: ncurses has complex build outputs (multiple libraries, symlinks)
  - **Mitigation**: Focus on core executables and let install_binaries handle what's needed

## Success Criteria
- [ ] Recipe installs successfully on all 4 platforms
- [ ] `pkg-config --libs ncurses` returns correct tsuku paths
- [ ] `pkg-config --modversion ncurses` returns the installed version
- [ ] CI validates installation on all platforms
- [ ] No pre-existing test failures introduced

## Open Questions
None - path forward is clear from existing patterns and acceptance criteria.
