# Issue 130 Summary

## What Was Implemented

Created a `cpan_install` action that installs Perl distributions via cpanm with local::lib isolation. The action generates wrapper scripts that set PERL5LIB at runtime, following the same pattern as gem_install for Ruby gems.

## Changes Made

- `internal/actions/cpan_install.go`: New action implementing CPAN distribution installation
  - Distribution name validation (isValidDistribution)
  - CPAN version validation (isValidCpanVersion)
  - Executable name validation against shell metacharacters
  - Wrapper script generation with PERL5LIB isolation
  - PERL* environment variable clearing during cpanm execution

- `internal/actions/cpan_install_test.go`: Comprehensive test suite
  - 23 distribution name validation tests
  - 17 version validation tests
  - 8 Execute() validation tests
  - Action registration test

- `internal/actions/util.go`: Added helper functions
  - ResolvePerl() - finds tsuku's perl installation
  - ResolveCpanm() - finds tsuku's cpanm from perl installation

- `internal/actions/action.go`: Registered CpanInstallAction

## Key Decisions

- **Following gem_install pattern**: Wrapper scripts rename original cpanm scripts to `.cpanm` suffix and create new wrappers that set PERL5LIB before execution
- **Clearing all PERL* env vars**: Prevents contamination from system Perl configuration
- **Relocatable wrappers**: Use BASH_SOURCE and symlink resolution to work after installation directory moves
- **Empty version valid**: Allows installing latest version when no version specified

## Trade-offs Accepted

- **Requires /bin/bash**: Wrapper scripts use bash-specific syntax; not portable to non-bash systems (matches gem_install)
- **No XS compilation support**: Pure Perl distributions only; XS modules requiring native compilation are out of scope
- **cpanm --notest**: Skips distribution tests for faster installation; security trade-off documented in design doc

## Test Coverage

- New tests added: 49 test cases across 5 test functions
- All tests pass

## Known Limitations

- Requires perl to be installed as a dependency (`dependencies = ["perl"]` in recipe)
- perl recipe not yet available in registry (blocked by tsuku-registry#13)
- No support for XS modules requiring native compilation

## Future Improvements

- Add CPAN builder (issue #131) to generate recipes from CPAN distribution names
- Add integration tests (issue #144) once perl recipe is available
