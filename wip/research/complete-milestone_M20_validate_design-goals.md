# Design Goal Validation: M20 - Dependency Provisioning: Full Integration

**Milestone**: 20 - Dependency Provisioning: Full Integration
**Design Document**: docs/DESIGN-dependency-provisioning.md
**Validation Date**: 2025-12-23

## Executive Summary

**Status**: FINDINGS
**Finding Count**: 2
**Overall Assessment**: Milestone 20 successfully delivered all three planned recipes (readline, sqlite, git) with proper dependency chains and CI validation. However, two findings were identified: (1) the git recipe uses a simplified dependency list that doesn't match the design document's specification, and (2) the Build Essentials CI workflow is currently failing on main branch.

## Design Goals for Milestone 20

From the design document (DESIGN-dependency-provisioning.md, lines 97-123), Milestone 20 "Dependency Provisioning: Full Integration" had the following goals:

### Phase 7: Full Integration (Complex Tools)

**Goal**: Validate complete toolchain with real-world complex tools.

**Planned Deliverables**:
1. **Issue #557**: readline recipe depending on ncurses
2. **Issue #558**: sqlite recipe depending on readline
3. **Issue #559**: git recipe depending on curl, openssl, zlib, expat

**Success Criteria** (from design, lines 977-986):
- Create recipes/readline.toml (depends on ncurses)
- Create recipes/sqlite.toml (depends on readline)
- Create recipes/git.toml (depends on curl, openssl, zlib, expat)
- CI: Build git from source, verify `git --version`
- CI: Verify git can clone a repository (full functional test)
- CI: Build sqlite, verify `sqlite3 --version`
- **Gate**: git and sqlite build and function correctly on all 4 platforms

## Implementation Analysis

### Issue #557: readline recipe

**Closed**: 2025-12-23
**PR**: #661 (merged)

**Implementation**:
- Recipe location: `internal/recipe/recipes/r/readline.toml`
- Declares dependency on ncurses: ✓
- Uses homebrew_bottle action for production
- Includes both Linux (.so) and macOS (.dylib) library files
- Installs libraries: libreadline, libhistory (static and shared)

**Validation**:
- CI test added: test-sqlite-source job validates readline provisioning
- Docker test script validates provisioning in clean Ubuntu 22.04
- Verification function added to test/scripts/verify-tool.sh
- **Status**: ✓ COMPLETE

### Issue #558: sqlite recipe

**Closed**: 2025-12-23
**PR**: #661 (merged)

**Implementation**:
- Production recipe: `internal/recipe/recipes/s/sqlite.toml` (homebrew bottle)
- Test recipe: `testdata/recipes/sqlite-source.toml` (configure_make from source)
- Declares dependency on readline: ✓
- Test recipe uses configure_make with --enable-readline flag
- Validates complete dependency chain: sqlite → readline → ncurses

**Validation**:
- CI test added: test-sqlite-source job (3 platforms: Linux x86_64, macOS Intel, macOS Apple Silicon)
- Verification includes:
  - sqlite3 --version works
  - Basic SQL operations (CREATE TABLE, INSERT, SELECT)
  - Readline support validated
- **Status**: ✓ COMPLETE

**Known Limitation**: Issue #660 filed for non-standard version format requiring manual URL updates

### Issue #559: git recipe

**Closed**: 2025-12-23
**PR**: #662 (merged)

**Implementation**:
- Production recipe: `internal/recipe/recipes/g/git.toml` (homebrew bottle)
- Test recipe: `testdata/recipes/git-source.toml` (configure_make from source)
- **FINDING #1**: Production recipe declares only `curl` dependency, not the full set
  - Design specifies: curl, openssl, zlib, expat
  - Actual implementation: `dependencies = ["curl"]`
  - Test recipe correctly includes all: curl, openssl, zlib, expat

**Validation**:
- CI test added: test-git-source job (3 platforms)
- Verification includes:
  - git --version works from relocated path
  - git clone successfully clones repository (validates HTTPS/TLS via curl/openssl)
- Verification function added to test/scripts/verify-tool.sh
- **Status**: ✓ IMPLEMENTED (with findings)

## Design Intent vs Implementation Comparison

### Stated Capabilities (Design Document)

From lines 971-986, Phase 7 should deliver:

1. **readline recipe with ncurses dependency**: ✓ Delivered
2. **sqlite recipe with readline dependency**: ✓ Delivered
3. **git recipe with complex multi-dependency chain**: ⚠️ Partially delivered (see Finding #1)
4. **CI validation on all platforms**: ✓ Delivered (but see Finding #2)
5. **Functional testing (git clone, sqlite queries)**: ✓ Delivered

### Code Evidence

**readline.toml** (lines 1-39):
```toml
[metadata]
name = "readline"
dependencies = ["ncurses"]  # ✓ Correct dependency

[[steps]]
action = "homebrew"
formula = "readline"
```

**sqlite.toml** (production, lines 1-24):
```toml
[metadata]
name = "sqlite"
dependencies = ["readline"]  # ✓ Correct dependency
```

**sqlite-source.toml** (test recipe, lines 1-47):
```toml
[metadata]
name = "sqlite-source"
dependencies = ["readline"]  # ✓ Correct dependency

[[steps]]
action = "configure_make"
configure_args = ["--enable-readline"]  # ✓ Validates readline integration
```

**git.toml** (production, lines 1-24):
```toml
[metadata]
name = "git"
dependencies = ["curl"]  # ⚠️ Missing openssl, zlib, expat
```

**git-source.toml** (test recipe, lines 1-46):
```toml
[metadata]
name = "git-source"
dependencies = ["curl", "openssl", "zlib", "expat"]  # ✓ Correct full dependency list

[[steps]]
action = "configure_make"
configure_args = ["--with-curl", "--with-openssl", "--with-zlib", "--with-expat"]
```

## Findings

### Finding #1: git production recipe has incomplete dependency list

**Severity**: Medium
**Type**: Deviation from design

**Description**: The production git recipe (`internal/recipe/recipes/g/git.toml`) declares only `curl` as a dependency, while the design document (line 981) explicitly specifies it should depend on "curl, openssl, zlib, expat".

**Evidence**:
- Design doc (line 981): "Create recipes/git.toml (depends on curl, openssl, zlib, expat)"
- Implementation: `dependencies = ["curl"]` (line 5 of git.toml)
- Test recipe correctly includes all four: `dependencies = ["curl", "openssl", "zlib", "expat"]`

**Impact**:
- Production git recipe using homebrew bottle may not correctly declare transitive dependencies
- The test recipe (git-source) does validate the complete dependency chain
- Since production uses homebrew bottle (pre-built), the missing declarations may not cause runtime failures, but they deviate from design intent for dependency transparency

**Recommendation**: Update production git.toml to include full dependency list for consistency with design document and transparency about what git requires.

### Finding #2: Build Essentials CI workflow failing on main branch

**Severity**: High
**Type**: Implementation quality issue

**Description**: The most recent run of the Build Essentials workflow on main branch failed (as of 2025-12-23T20:45:29Z).

**Evidence**:
- GitHub Actions run: conclusion=failure, status=completed
- Multiple failed runs visible in recent history
- Workflow file exists at `.github/workflows/build-essentials.yml`

**Impact**:
- Cannot verify that M20 recipes pass CI validation on main branch
- CI gate mentioned in design (line 986) may not be satisfied
- Tests on feature branch succeeded, but main branch integration failed

**Recommendation**: Investigate and fix the Build Essentials CI failure to ensure the milestone gate criteria are met.

## Gap Analysis

### Missing Features

None identified. All three planned recipes were delivered.

### Partial Implementations

**Git production recipe dependency list**: As noted in Finding #1, the production git recipe has a simplified dependency list compared to design specification. However, this may be intentional for homebrew bottle recipes where dependencies are embedded in the bottle itself.

### Significant Deviations

**Platform coverage**: Design mentions "all 4 platforms" (line 986) including:
- Linux x86_64 ✓
- Linux arm64 (excluded - Homebrew doesn't publish bottles)
- macOS Intel ✓
- macOS Apple Silicon ✓

The implementation tests on 3 platforms instead of 4, which is noted in CI comments. This appears to be a practical constraint rather than a design deviation.

## Validation Summary

### Capabilities Delivered

1. ✓ **readline recipe**: Provides GNU Readline library with ncurses dependency
2. ✓ **sqlite recipe**: Validates library dependency chains (sqlite → readline → ncurses)
3. ✓ **git recipe**: Validates complex multi-dependency toolchain
4. ✓ **CI validation**: Automated testing on 3 platforms
5. ✓ **Functional testing**: git clone, sqlite SQL operations verified
6. ✓ **Build environment**: setup_build_env action correctly provisions dependencies
7. ✓ **Source builds**: configure_make action works with library dependencies

### Design Goals Met

- [x] Create readline recipe depending on ncurses
- [x] Create sqlite recipe depending on readline
- [⚠️] Create git recipe depending on curl, openssl, zlib, expat (partial - simplified in production)
- [x] CI builds git from source
- [x] CI verifies git --version
- [x] CI verifies git can clone repository
- [x] CI builds sqlite
- [x] CI verifies sqlite3 --version
- [⚠️] Gate: git and sqlite build correctly on all platforms (main branch CI failing)

### Testing Coverage

**CI Jobs Added** (from .github/workflows/build-essentials.yml):
- test-sqlite-source: Lines 188-223 (validates sqlite → readline → ncurses chain)
- test-git-source: Lines 225-260 (validates git → curl → openssl/zlib + expat chain)

**Verification Scripts** (from test/scripts/verify-tool.sh):
- verify_readline(): Tests library file existence and installation
- verify_sqlite(): Tests version, basic SQL operations
- verify_git(): Tests version and git clone functionality

**Platform Coverage**:
- Linux x86_64: ✓
- macOS Intel (macos-15-intel): ✓
- macOS Apple Silicon (macos-14): ✓
- Linux arm64: Excluded (Homebrew limitation)

## Files Modified

**Recipes**:
- internal/recipe/recipes/r/readline.toml (added)
- internal/recipe/recipes/s/sqlite.toml (added)
- internal/recipe/recipes/g/git.toml (added)
- testdata/recipes/sqlite-source.toml (added)
- testdata/recipes/git-source.toml (added)

**CI/Testing**:
- .github/workflows/build-essentials.yml (modified - 2 new jobs)
- test/scripts/verify-tool.sh (modified - 3 new verification functions)
- test/scripts/test-readline-provisioning.sh (added - Docker test)

**Documentation**:
- docs/DESIGN-dependency-provisioning.md (updated - mermaid diagrams marked #557, #558, #559 as done)

## Conclusion

Milestone 20 successfully delivers the core functionality specified in the design document: three complex recipes (readline, sqlite, git) with proper dependency chains, validated through comprehensive CI testing across three platforms. The implementation demonstrates that tsuku can provision complex multi-library dependency chains end-to-end.

Two findings require attention:
1. The production git recipe should declare its full dependency list for consistency with design
2. The Build Essentials CI workflow failure on main branch needs investigation

Despite these findings, the milestone substantially achieves its stated goal of "Full Integration" - validating the complete toolchain with real-world complex tools.

## Recommendations

1. **Immediate**: Investigate and resolve Build Essentials CI failure on main branch
2. **Short-term**: Update git.toml to include full dependency list (curl, openssl, zlib, expat) for design consistency
3. **Medium-term**: Address issue #660 (sqlite version format) for improved maintainability
4. **Long-term**: Consider Linux arm64 support once Homebrew bottle availability improves
