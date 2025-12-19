# Issue 547 Summary

## What Was Implemented
Added Dependencies() methods to configure_make and cmake_build actions to declare their implicit install-time dependencies. These dependencies are automatically installed by the existing dependency resolution system when these build actions are used in recipes.

## Changes Made
- `internal/actions/configure_make.go`: Added Dependencies() method returning InstallTime deps: make, zig, pkg-config
- `internal/actions/cmake_build.go`: Added Dependencies() method returning InstallTime deps: cmake, make, zig, pkg-config
- `internal/actions/dependencies_test.go`: Added TestActionDependencies_BuildActions test function with cases for both build actions
- `wip/issue_547_plan.md`: Marked all implementation steps as complete

## Key Decisions
- **Use Action interface method instead of registry map**: The design doc mentioned an ActionDependencies registry, but the codebase uses the Action.Dependencies() interface method. This is more type-safe and keeps dependency declarations with their actions.
- **Only implement for existing actions**: Acceptance criteria mentioned meson_build, but it doesn't exist in the codebase yet. Implemented only for configure_make and cmake_build.

## Trade-offs Accepted
- **meson_build not implemented**: Will need to be added when that action is created in the future.

## Test Coverage
- New tests added: 1 test function (TestActionDependencies_BuildActions) with 2 test cases
- All existing tests continue to pass
- Coverage maintained at baseline levels

## Known Limitations
- meson_build dependencies not declared (action doesn't exist yet)

## Future Improvements
When meson_build action is added, it should declare InstallTime dependencies: meson, ninja, zig, pkg-config
