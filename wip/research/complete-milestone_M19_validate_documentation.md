# Documentation Gap Analysis: Milestone M19 (Dependency Provisioning: Build Environment)

## Milestone Summary

**Milestone 19: Dependency Provisioning: Build Environment**

This milestone implemented comprehensive build environment provisioning capabilities, enabling tsuku to provide build tools and libraries needed for source compilation. The implementation includes:

### Closed Issues (10 total)

1. **#547** - feat(actions): declare implicit dependencies for build actions
2. **#548** - feat(recipes): add pkg-config recipe using homebrew
3. **#549** - feat(recipes): add cmake recipe using homebrew
4. **#550** - feat(actions): enhance buildAutotoolsEnv with dependency paths
5. **#551** - feat(actions): implement setup_build_env action
6. **#552** - feat(recipes): add openssl recipe using homebrew
7. **#553** - feat(recipes): add ncurses recipe to validate pkg-config
8. **#554** - feat(recipes): add curl recipe to validate openssl
9. **#555** - feat(actions): implement cmake_build action
10. **#556** - feat(recipes): add ninja recipe to validate cmake_build

### Key Features Delivered

1. **New Actions**:
   - `setup_build_env` - Configures build environment from dependencies
   - `cmake_build` - CMake-based source builds

2. **Implicit Action Dependencies**:
   - `configure_make` implicitly requires: make, zig, pkg-config
   - `cmake_build` implicitly requires: cmake, make, zig, pkg-config

3. **Build Environment Configuration**:
   - PKG_CONFIG_PATH set from all install-time dependencies
   - CPPFLAGS with -I flags for dependency include directories
   - LDFLAGS with -L flags for dependency lib directories
   - CMAKE_PREFIX_PATH for CMake builds
   - CC/CXX compiler selection (zig cc fallback)

4. **New Recipes** (added to internal/recipe/recipes/):
   - pkg-config (build tool)
   - cmake (build system)
   - openssl (TLS/crypto library)
   - ncurses (terminal UI library)
   - curl (HTTP client with TLS support)
   - ninja (fast build tool)

## Documentation Coverage Assessment

### 1. Design Documentation

**File**: `docs/DESIGN-dependency-provisioning.md`

**Status**: EXCELLENT - Comprehensive and up-to-date

**Coverage**:
- ✅ Complete architecture explanation
- ✅ Implementation approach with all phases documented
- ✅ Milestone 19 implementation issues table fully updated (all marked "Done")
- ✅ Mermaid dependency diagram present and accurate
- ✅ Technical details for `setup_build_env` action
- ✅ Technical details for `cmake_build` action
- ✅ Build environment configuration explained
- ✅ Security considerations covered
- ✅ Platform test matrix documented

**Strengths**:
- Contains actual Go code examples for build environment setup
- Documents implicit dependency registration
- Explains buildAutotoolsEnv() enhancements
- Provides validation criteria and scripts
- Shows complete dependency chains across all milestones

**Gaps**: None identified

---

### 2. User-Facing Guide Documentation

**File**: `docs/GUIDE-actions-and-primitives.md`

**Status**: GOOD - Accurate but could be enhanced

**Coverage**:
- ✅ Lists `cmake_build` as an ecosystem primitive
- ✅ Documents `configure_make` action
- ✅ Section on "Build Environment Configuration" (lines 391-416)
- ✅ Explains automatic build environment setup
- ✅ Lists environment variables set (CC/CXX, PKG_CONFIG_PATH, CPPFLAGS, LDFLAGS, CMAKE_PREFIX_PATH)
- ✅ Notes that build essentials are installed as implicit dependencies

**Gaps**:
1. **Missing `setup_build_env` in action tables** - The action is mentioned in context but not listed in the action type tables
2. **No explicit examples** - While build environment configuration is explained, there's no concrete recipe example showing `setup_build_env` usage
3. **cmake_build parameters not documented** - No explanation of cmake_args, build_flags, etc.

**Recommendation**: Add `setup_build_env` to the "Build System Primitives" section with brief description

---

### 3. Build Essentials Documentation

**File**: `docs/BUILD-ESSENTIALS.md`

**Status**: INCOMPLETE - Missing new tools from M19

**Coverage**:
- ✅ Documents zig (compiler)
- ✅ Documents make (build tool)
- ✅ Documents zlib (library)
- ✅ Platform support information
- ✅ Validation information
- ✅ References to Actions guide

**Gaps**:
1. **Missing pkg-config** - Issue #548 added pkg-config, not documented
2. **Missing cmake** - Issue #549 added cmake, not documented
3. **Missing openssl** - Issue #552 added openssl, not documented
4. **Missing ncurses** - Issue #553 added ncurses, not documented
5. **Missing curl** - Issue #554 added curl, not documented (though curl may be a consumer, not essential)
6. **Missing ninja** - Issue #556 added ninja, not documented

**Recommendation**: Add sections for:
- pkg-config (Build Tools section)
- cmake (Build Tools section)
- openssl (Libraries section)
- ncurses (Libraries section)

---

### 4. Main README

**File**: `README.md`

**Status**: GOOD - General coverage adequate

**Coverage**:
- ✅ Section "Build Dependency Provisioning" (lines 226-242)
- ✅ Lists compilers: zig (with explanation)
- ✅ Lists build tools: make, pkg-config, cmake, autoconf, automake
- ✅ Lists libraries: zlib, openssl, ncurses, readline
- ✅ Example showing gdbm source build
- ✅ Notes automatic installation of build dependencies

**Gaps**:
1. **No mention of cmake_build action** - Only configure_make is implied
2. **No mention of setup_build_env action** - Users might want to know about explicit build env configuration
3. **Missing ninja** - Listed in implementation but not in README

**Recommendation**:
- Add brief mention that cmake-based builds are supported alongside autotools
- Consider adding ninja to build tools list

---

### 5. Website Documentation

**File**: `website/index.html`

**Status**: ADEQUATE - Landing page doesn't need deep technical details

**Coverage**:
- ✅ General "self-contained" messaging
- ✅ "Downloads pre-built tools or builds them in isolation" (line 45)

**Gaps**:
- No specific mention of build essentials, but this is appropriate for a landing page

**Recommendation**: No changes needed - landing page maintains appropriate level of detail

---

### 6. CLI Help Text

**File**: `cmd/tsuku/main.go`

**Status**: ADEQUATE - General help text doesn't detail actions

**Coverage**:
- ✅ General description of action-based recipes (lines 34-38)

**Gaps**:
- No command-level help mentions new actions, but this is expected (help is generated from recipe/action metadata)

**Recommendation**: No changes needed - CLI help is appropriately high-level

---

## Gap Analysis Summary

### Critical Gaps (Must Fix)

1. **BUILD-ESSENTIALS.md missing new tools**:
   - pkg-config (build tool)
   - cmake (build system)
   - openssl (library)
   - ncurses (library)
   - Consider: curl and ninja (consumer tools vs essentials)

### Important Gaps (Should Fix)

2. **GUIDE-actions-and-primitives.md incomplete**:
   - `setup_build_env` action not listed in action tables
   - `cmake_build` action parameters not documented
   - Missing concrete recipe examples using new actions

### Minor Gaps (Nice to Have)

3. **README.md enhancement opportunities**:
   - cmake_build action not mentioned
   - setup_build_env action not mentioned
   - ninja not listed in build tools

### No Action Needed

4. **Well-documented areas**:
   - ✅ DESIGN-dependency-provisioning.md - Excellent technical documentation
   - ✅ README.md general build provisioning section - Good overview
   - ✅ Website - Appropriate level of detail for landing page

---

## Verification Evidence

### Recipes Confirmed in Repository

Using `Glob` and `Read`, confirmed these recipes exist:

1. `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/n/ncurses.toml` - Uses `setup_build_env` and `configure_make`
2. `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/c/curl.toml` - Declares dependencies on openssl and zlib, uses `setup_build_env`
3. `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/n/ninja.toml` - Uses `cmake_build` action

### Implementation Code Confirmed

Using `Grep`, confirmed these files exist:

1. `internal/actions/setup_build_env.go` and `setup_build_env_test.go`
2. `internal/actions/cmake_build.go` and `cmake_build_test.go`
3. Multiple references to `buildAutotoolsEnv` enhancement

---

## Recommendations

### Priority 1: Update BUILD-ESSENTIALS.md

Add documentation for new build essentials:

```markdown
### Build Tools (add to existing section)

**pkg-config**
- Library discovery tool
- Reads .pc files from dependency lib/pkgconfig directories
- Required by configure_make and cmake_build actions
- Installed from Homebrew bottles
- Recipe: `internal/recipe/recipes/p/pkg-config.toml`

**cmake**
- Modern build system generator
- Required by cmake_build action
- Cross-platform support
- Installed from Homebrew bottles
- Recipe: `internal/recipe/recipes/c/cmake.toml`

### Libraries (add to existing section)

**openssl**
- TLS/crypto library (libssl, libcrypto)
- Dependency for tools requiring HTTPS/TLS support
- Complex RPATH configuration for relocation
- Installed from Homebrew bottles
- Recipe: `internal/recipe/recipes/o/openssl.toml`

**ncurses**
- Terminal UI library (libncurses)
- Dependency for terminal-based applications
- Built from source with pkg-config support
- Recipe: `internal/recipe/recipes/n/ncurses.toml`

**curl** (Consumer Tool)
- HTTP client with OpenSSL/TLS support
- Example consumer of openssl + zlib dependencies
- Validates complete build environment provisioning
- Recipe: `internal/recipe/recipes/c/curl.toml`

**ninja** (Consumer Tool)
- Fast build tool
- Example consumer of cmake_build action
- Validates cmake-based source builds
- Recipe: `internal/recipe/recipes/n/ninja.toml`
```

### Priority 2: Enhance GUIDE-actions-and-primitives.md

Add to "Build System Primitives" section (around line 56-64):

```markdown
#### Build Environment Setup

**setup_build_env** - Configures build environment from dependency graph

This action is typically used before build actions to set up environment variables:

```toml
[[steps]]
action = "setup_build_env"

[[steps]]
action = "configure_make"
source_dir = "curl-{version}"
configure_args = ["--with-openssl", "--with-zlib"]
```

The setup_build_env action:
- Sets PKG_CONFIG_PATH from all dependency lib/pkgconfig paths
- Sets CPPFLAGS with -I flags for dependency include directories
- Sets LDFLAGS with -L flags for dependency lib directories
- Sets CMAKE_PREFIX_PATH for CMake-based builds
- Configures CC/CXX compiler (uses zig cc if no system compiler)
```

Add cmake_build parameters documentation:

```markdown
**cmake_build parameters**:
- `source_dir` - Directory containing CMakeLists.txt
- `cmake_args` - Arguments passed to cmake configure step
- `build_flags` - Arguments passed to cmake build step (optional)
- `install_prefix` - Installation prefix (defaults to .install)
- `executables` - List of binaries to build
```

### Priority 3: Minor README.md enhancements

Update lines 228-232 to mention cmake:

```markdown
- **Compilers**: zig (C/C++ via zig cc fallback when system compiler unavailable)
- **Build tools**: make, pkg-config, cmake, ninja, autoconf, automake
- **Common libraries**: zlib, openssl, ncurses, readline
```

Add brief note about cmake support:

```markdown
When you install a tool that requires compilation (using autotools or cmake),
tsuku automatically installs the necessary build dependencies. No manual setup required.
```

---

## Documentation Quality Assessment

### Strengths

1. **Excellent technical documentation** - DESIGN-dependency-provisioning.md is comprehensive and well-maintained
2. **Good user-facing guides** - Actions guide and plan-based installation guide are clear
3. **Updated design docs** - Milestone implementation tables are current
4. **Consistent style** - Documentation follows established patterns

### Weaknesses

1. **Incomplete build essentials inventory** - BUILD-ESSENTIALS.md missing 6 tools from M19
2. **Action reference gaps** - setup_build_env action mentioned in prose but not in reference tables
3. **Missing concrete examples** - Few recipes showing setup_build_env + cmake_build usage

### Overall Grade: B+

The core technical documentation is excellent, but the user-facing reference documentation (BUILD-ESSENTIALS.md) has fallen behind the implementation. The gaps are straightforward to address with the recommendations above.

---

## Conclusion

Milestone M19 delivered significant functionality for build environment provisioning, and the design documentation accurately reflects this work. However, user-facing documentation has gaps:

- **Critical**: BUILD-ESSENTIALS.md missing 6 new tools (pkg-config, cmake, openssl, ncurses, curl, ninja)
- **Important**: GUIDE-actions-and-primitives.md missing setup_build_env in action tables and cmake_build parameters
- **Minor**: README.md could mention cmake support alongside autotools

These gaps are easily addressable and do not impact the technical quality of the implementation - they simply mean users may not discover all available build essentials without reading the design doc or browsing recipe files directly.
