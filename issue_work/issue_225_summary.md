# Issue 225 Summary

## What Was Implemented
Added a check in the install command to prevent direct installation of library recipes, showing a helpful error message that explains libraries are installed automatically as dependencies.

## Changes Made
- `cmd/tsuku/install.go`: Added check after `IsLibrary()` detection to block direct installs when `isExplicit=true` and `parent=""`
- `cmd/tsuku/dependency_test.go`: Added `TestLibraryInstallPrevention` with 4 test cases

## Key Decisions
- Check placement: After recipe loading but before `installLibrary()` call - recipe must be loaded to determine if it's a library
- Error message: Includes specific example (`tsuku install ruby` for libyaml) to guide users

## Trade-offs Accepted
- Logic-based test instead of integration test: The actual `installWithDependencies` function is difficult to test in isolation due to global dependencies, so we test the blocking condition logic directly

## Test Coverage
- New tests added: 1 table-driven test with 4 cases
- Coverage change: Minimal impact, new code is simple conditional

## Known Limitations
- Error message uses hardcoded example (libyaml/ruby) - could be made dynamic in the future

## Future Improvements
- Could dynamically show which tools depend on the library in the error message
