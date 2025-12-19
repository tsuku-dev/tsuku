# Issue 547 Implementation Plan

## Summary
Add Dependencies() methods to configure_make, cmake_build, and meson_build actions to declare their implicit install-time dependencies (make, zig, pkg-config, cmake, ninja).

## Approach
Following the established pattern in the codebase where each action implements the Dependencies() method from the Action interface. The existing ResolveDependencies() function in actions/resolver.go already handles collecting these dependencies, so we only need to override the BaseAction default for the three build actions.

### Alternatives Considered
- **ActionDependencies registry map**: The design doc mentions this approach, but the codebase uses the Action interface method instead, which is more type-safe and discoverable.
- **Centralizing all dependencies in one file**: Current pattern distributes dependencies to each action file, keeping related code together. More maintainable.

## Files to Modify
- `internal/actions/configure_make.go` - Add Dependencies() method returning {InstallTime: ["make", "zig", "pkg-config"]}
- `internal/actions/cmake_build.go` - Add Dependencies() method returning {InstallTime: ["cmake", "make", "zig", "pkg-config"]}
- `internal/actions/dependencies_test.go` - Add test cases for the three build actions

## Files to Create
None - meson_build doesn't exist yet, so we'll only implement configure_make and cmake_build per the actual codebase state.

## Implementation Steps
- [ ] Add Dependencies() method to ConfigureMakeAction
- [ ] Add Dependencies() method to CMakeBuildAction
- [ ] Add test cases for configure_make and cmake_build dependencies
- [ ] Run tests to verify dependencies are resolved correctly

## Testing Strategy
- Unit tests: Add test cases to dependencies_test.go for the new build action dependencies
- Integration: The existing ResolveDependencies tests should cover the integration
- Manual verification: Build a recipe using configure_make or cmake_build and verify dependencies are installed

## Risks and Mitigations
- **Risk**: meson_build mentioned in acceptance criteria but doesn't exist yet
  - **Mitigation**: Only implement for existing actions (configure_make, cmake_build). Document that meson_build should be added when that action is created.

## Success Criteria
- [ ] configure_make reports InstallTime: ["make", "zig", "pkg-config"]
- [ ] cmake_build reports InstallTime: ["cmake", "make", "zig", "pkg-config"]
- [ ] Tests pass for both actions
- [ ] No existing tests break

## Open Questions
None - the pattern is clear from existing ecosystem actions (go_build, cargo_build, npm_exec, etc.)
