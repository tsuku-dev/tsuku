# Documentation Gap Analysis: Milestone M20 (Dependency Provisioning: Full Integration)

**Milestone**: M20 - Dependency Provisioning: Full Integration
**Analysis Date**: 2025-12-23
**Status**: FINDINGS

## Executive Summary

Milestone M20 successfully delivered three recipes (readline, sqlite, git) that validate the complete dependency provisioning system. All three recipes exist and are tested in CI. However, there are **significant documentation gaps** that prevent users from understanding and using these new capabilities.

**Finding Count**: 7 major documentation gaps identified

**Key Issues**:
1. Main README does not mention readline, sqlite, or git recipes
2. No user-facing documentation for dependency auto-provisioning workflow
3. BUILD-ESSENTIALS.md is outdated and incomplete
4. Missing examples of multi-dependency chain usage
5. No mention of the new library dependency features in user guides
6. Technical documentation exists but lacks user-friendly examples
7. No migration guide for users who previously needed system packages

---

## Milestone Overview

### Closed Issues

M20 contained three issues, all closed:

1. **#557**: feat(recipes): add readline recipe using homebrew_bottle
   - Merged in PR #661
   - Commit: 45d8d4c

2. **#558**: feat(recipes): add sqlite recipe to validate readline integration
   - Merged in PR #661
   - Commit: 45d8d4c

3. **#559**: feat(recipes): add git recipe to validate complete toolchain
   - Merged in PR #662
   - Commit: dbd3eb3

### What Was Delivered

**Recipes Created**:
- `internal/recipe/recipes/r/readline.toml` - GNU Readline library with ncurses dependency
- `internal/recipe/recipes/s/sqlite.toml` - SQLite with readline dependency
- `internal/recipe/recipes/g/git.toml` - Git with curl dependency

**Dependency Chains Validated**:
- sqlite → readline → ncurses (library dependency chain)
- git → curl → openssl/zlib + expat (multi-library dependency chain)

**Testing Infrastructure**:
- `test/scripts/verify-tool.sh` - Added verify functions for readline, sqlite, git
- `test/scripts/test-readline-provisioning.sh` - Docker-based validation
- `.github/workflows/build-essentials.yml` - CI jobs for sqlite and git-source

**Design Documentation**:
- `docs/DESIGN-dependency-provisioning.md` - Updated mermaid diagrams marking issues as done

---

## Documentation Coverage Analysis

### 1. Main README.md

**Current State**: The README does not mention readline, sqlite, or git recipes.

**Gaps Identified**:

1. **No mention of library provisioning** - Users installing tools that need libraries (like sqlite) have no guidance that tsuku now auto-provisions dependencies like readline.

2. **Missing examples** - No example showing:
   ```bash
   # This now works - readline and ncurses are auto-provisioned
   tsuku install sqlite
   ```

3. **Build Dependency Provisioning section incomplete** - Section exists (lines 226-242) but only mentions compilers and build tools. Does NOT mention library dependencies like readline, ncurses, openssl, or curl.

   Current content:
   ```markdown
   - **Compilers**: zig (C/C++ via zig cc fallback when system compiler unavailable)
   - **Build tools**: make, pkg-config, cmake, autoconf, automake
   - **Common libraries**: zlib, openssl, ncurses, readline
   ```

   The "Common libraries" bullet exists but provides NO context about:
   - Which of these are actually available (all 4 are now recipes)
   - That they auto-install as dependencies
   - Example workflows using them
   - That this eliminates need for system packages

4. **No git recipe mentioned** - Git is a major tool that many developers use. The README should highlight that git is now available via tsuku with automatic dependency provisioning.

5. **Missing user value proposition** - Should explain the user benefit: "No need to install system packages like libreadline-dev or libncurses-dev - tsuku provides everything."

**Recommendation**: Add subsection to "Build Dependency Provisioning" with:
- List of available library recipes (zlib, openssl, ncurses, readline, curl)
- Example showing sqlite auto-provisioning readline and ncurses
- Example showing git auto-provisioning curl and its dependencies
- Clear statement that users no longer need apt-get/brew for these dependencies

---

### 2. BUILD-ESSENTIALS.md

**Current State**: Outdated and incomplete. Does not reflect M20 deliverables.

**Gaps Identified**:

1. **Missing Libraries Section** - Document lists zlib, gdbm, libpng, pngcrush but does NOT list:
   - readline (delivered in M20)
   - sqlite (delivered in M20)
   - git (delivered in M20)
   - ncurses (delivered in M19, required by readline)
   - curl (delivered in M19, required by git)
   - openssl (delivered in M19, required by curl)

2. **Incomplete Available Build Essentials section** - Only shows:
   - Compilers: zig
   - Build Tools: make
   - Libraries: zlib, gdbm, libpng, pngcrush

   Should include full inventory from M18-M20:
   - Libraries: zlib, openssl, ncurses, readline, curl, libpng, pngcrush, gdbm, expat
   - Build Tools: make, pkg-config, cmake, ninja
   - Compilers: zig

3. **No dependency chain documentation** - Should document the key dependency chains that M20 validated:
   - sqlite → readline → ncurses
   - git → curl → openssl + zlib + expat

4. **Missing recipe paths** - For new libraries, should document recipe locations:
   - readline: `internal/recipe/recipes/r/readline.toml`
   - sqlite: `internal/recipe/recipes/s/sqlite.toml`
   - git: `internal/recipe/recipes/g/git.toml`

5. **Platform Support section outdated** - States "arm64 Linux is not currently supported for Homebrew bottles" but doesn't clarify which recipes this affects (most? all?)

**Recommendation**: Complete rewrite of BUILD-ESSENTIALS.md to:
- Reflect all recipes from M18, M19, M20
- Document dependency chains
- Add examples of using sqlite, git
- Clarify platform support matrix for each essential

---

### 3. GUIDE-actions-and-primitives.md

**Current State**: Comprehensive technical documentation but lacks user-facing examples.

**Gaps Identified**:

1. **No mention of readline, sqlite, git recipes** - The guide documents the `homebrew` action (lines 83-86) but provides no concrete examples using the M20 recipes.

2. **Library dependency example missing** - Should show how a tool recipe declares library dependencies:
   ```toml
   # Example from sqlite.toml
   [metadata]
   name = "sqlite"
   dependencies = ["readline"]

   [[steps]]
   action = "homebrew"
   formula = "sqlite"
   ```

3. **Multi-dependency chain example missing** - Should show git recipe as example of complex multi-dep chain:
   ```toml
   # git depends on curl, which depends on openssl + zlib + expat
   dependencies = ["curl"]
   ```

4. **Build Environment Configuration section** - Exists (lines 391-416) but focuses only on source builds. Should mention that library dependencies (readline, openssl) are also auto-configured via PKG_CONFIG_PATH.

**Recommendation**: Add subsection "Library Dependencies in Practice" with examples from M20 recipes.

---

### 4. DESIGN-dependency-provisioning.md

**Current State**: Excellent technical design documentation. Up to date with mermaid diagrams marking M20 issues as done.

**Gaps Identified**:

1. **No user-facing summary** - The design doc is 1,256 lines of technical detail. Needs a user-facing summary section at the top explaining:
   - What dependency provisioning is
   - What libraries are now available
   - How to use them (simple examples)
   - Link to user guide for more

2. **Missing "What This Means for Users" section** - Should translate the technical achievement into user benefits:
   - "You can now install sqlite without installing libreadline-dev"
   - "Git installs with all dependencies auto-provisioned"
   - "No more apt-get install build-essential"

3. **Examples use hypothetical recipes** - Many examples use `my-docker-tool.toml` and `gpu-app.toml` (lines 623-654). Should add real examples from M20:
   - sqlite.toml (real recipe delivered in M20)
   - git.toml (real recipe delivered in M20)

**Recommendation**: Add "User Guide Summary" section at top. Update examples to reference real delivered recipes.

---

### 5. DESIGN-homebrew.md

**Current State**: Comprehensive documentation of the homebrew action. Up to date technically.

**Gaps Identified**:

1. **No examples using M20 recipes** - The example recipe (lines 42-61) uses jq. Should add examples using readline, sqlite, or git to demonstrate library dependency handling.

2. **Dependency Discovery section** - Shows neovim example (lines 186-211) but doesn't show simpler examples like:
   ```
   $ tsuku create sqlite --from homebrew:sqlite

   Discovering dependencies...

   Dependency tree for sqlite:
     sqlite (needs recipe)
     └── readline (has recipe ✓)
         └── ncurses (has recipe ✓)

   All dependencies satisfied. Ready to install.
   ```

3. **Platform Support section** - Mentions 4 platforms (lines 63-77) but doesn't clarify which M20 recipes work on which platforms (all should work on 3 platforms per CI config).

**Recommendation**: Add M20 recipe examples. Show dependency tree for sqlite as a simple example.

---

### 6. User Workflow Documentation

**Current State**: No dedicated guide for "Installing Tools with Library Dependencies"

**Gap Identified**: Users upgrading from older tsuku versions or coming from other package managers need guidance on:

1. **What changed** - "You no longer need to install system packages for common libraries"

2. **Migration examples**:
   ```bash
   # Old workflow (before M18-M20):
   sudo apt-get install libreadline-dev libncurses-dev
   tsuku install my-tool

   # New workflow (after M18-M20):
   tsuku install my-tool  # Everything auto-provisioned
   ```

3. **Troubleshooting** - What if a library dependency is missing? How does tsuku handle it?

4. **Available libraries** - Clear list of which libraries tsuku now provides:
   - zlib (compression)
   - openssl (TLS/crypto)
   - ncurses (terminal UI)
   - readline (line editing)
   - curl (HTTP client)
   - libpng (PNG images)
   - expat (XML parsing)
   - gdbm (key-value database)

**Recommendation**: Create new guide: `docs/GUIDE-library-dependencies.md`

---

### 7. CLI Help Text / Command Documentation

**Gap Identified**: No way to discover available build essentials via CLI.

**Missing Commands/Features**:

1. **List build essentials** - No command to show available libraries:
   ```bash
   tsuku list --build-essentials
   # Should show: zlib, openssl, ncurses, readline, curl, etc.
   ```

2. **Show dependencies** - No way to preview what will be auto-installed:
   ```bash
   tsuku install sqlite --dry-run
   # Should show: Will install: sqlite, readline, ncurses
   ```

3. **Info command lacks dependency info** - `tsuku info sqlite` should show:
   ```
   Name: sqlite
   Dependencies: readline
   Transitive: readline → ncurses
   ```

**Recommendation**: File follow-up issues for CLI enhancements. For now, document the limitation in user guides.

---

## Detailed Findings Summary

### Finding 1: README.md Missing M20 Features
**Severity**: High
**Impact**: Users unaware of major new capabilities
**Location**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/README.md`
**Line Numbers**: 226-242 (Build Dependency Provisioning section)

**Current Content**:
```markdown
### Build Dependency Provisioning

tsuku automatically provides build tools and libraries needed for source builds, eliminating the need for system dependencies:

- **Compilers**: zig (C/C++ via zig cc fallback when system compiler unavailable)
- **Build tools**: make, pkg-config, cmake, autoconf, automake
- **Common libraries**: zlib, openssl, ncurses, readline
```

**Missing**:
- No examples using readline, sqlite, git
- No explanation of auto-provisioning workflow
- No user value proposition (eliminate apt-get/brew for libraries)
- List of libraries is incomplete (missing curl, libpng, expat, gdbm)

**Suggested Addition**:
```markdown
### Build Dependency Provisioning

tsuku automatically provides build tools and libraries needed for source builds, eliminating the need for system dependencies:

- **Compilers**: zig (C/C++ via zig cc fallback when system compiler unavailable)
- **Build tools**: make, pkg-config, cmake, ninja
- **Common libraries**: zlib, openssl, ncurses, readline, curl, libpng, expat, gdbm

#### Available Library Recipes

tsuku now includes recipes for common development libraries. When you install a tool that depends on these libraries, tsuku automatically provisions them:

**Core Libraries**:
- `zlib` - Compression library
- `openssl` - TLS/crypto library
- `curl` - HTTP client library
- `expat` - XML parser

**Terminal Libraries**:
- `ncurses` - Terminal UI library
- `readline` - Line editing library

**Other**:
- `libpng` - PNG image library
- `gdbm` - GNU database manager

#### Example: Installing SQLite

SQLite requires readline (which requires ncurses). tsuku auto-provisions the entire dependency chain:

\`\`\`bash
# Old workflow - required system packages:
# sudo apt-get install libreadline-dev libncurses-dev

# New workflow - tsuku provides everything:
tsuku install sqlite
# Auto-installs: sqlite → readline → ncurses
\`\`\`

#### Example: Installing Git

Git requires curl (which requires openssl, zlib, and expat). All dependencies are auto-provisioned:

\`\`\`bash
tsuku install git
# Auto-installs: git → curl → openssl + zlib + expat
\`\`\`

No manual dependency installation required. Everything is isolated to `$TSUKU_HOME`.
```

---

### Finding 2: BUILD-ESSENTIALS.md Outdated
**Severity**: High
**Impact**: Incomplete reference documentation
**Location**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/docs/BUILD-ESSENTIALS.md`
**Line Numbers**: 5-47 (Available Build Essentials section)

**Current Content**:
Lists only zlib, make, zig, gdbm, libpng, pngcrush.

**Missing**:
- readline (M20)
- sqlite (M20)
- git (M20)
- ncurses (M19)
- curl (M19)
- openssl (M19)
- pkg-config (M19)
- cmake (M19)
- ninja (M19)
- expat (M18)

**Suggested Structure**:
```markdown
## Available Build Essentials

### Compilers
**zig** - C/C++ compiler via `zig cc`

### Build Tools
**make** - GNU Make
**pkg-config** - Library discovery
**cmake** - CMake build system
**ninja** - Fast build tool

### Core Libraries
**zlib** - Compression (libz)
**openssl** - TLS/crypto
**curl** - HTTP client
**expat** - XML parser

### Terminal Libraries
**ncurses** - Terminal UI
**readline** - Line editing

### Graphics Libraries
**libpng** - PNG image library

### Databases
**gdbm** - GNU database manager
**sqlite** - SQLite database (with readline support)

### Full Stack Tools
**git** - Distributed version control (with curl, openssl, zlib, expat)

## Dependency Chains

Key validated dependency chains:

1. **sqlite → readline → ncurses**
   - Installing sqlite auto-provisions readline and ncurses

2. **git → curl → openssl + zlib + expat**
   - Installing git auto-provisions complete HTTP/TLS stack

3. **pngcrush → libpng → zlib**
   - Image tools get compression library automatically
```

---

### Finding 3: No User Guide for Library Dependencies
**Severity**: Medium
**Impact**: Users lack workflow documentation
**Location**: N/A - document does not exist
**Suggested Path**: `docs/GUIDE-library-dependencies.md`

**Needed Content**:
1. Introduction to library dependency auto-provisioning
2. List of available library recipes
3. How dependency chains work
4. Examples (sqlite, git)
5. Troubleshooting common issues
6. Relationship to system packages (users can skip apt-get/brew)

---

### Finding 4: GUIDE-actions-and-primitives.md Missing Examples
**Severity**: Medium
**Impact**: Technical documentation lacks real-world context
**Location**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/docs/GUIDE-actions-and-primitives.md`
**Line Numbers**: 83-86 (homebrew action), 391-416 (build environment)

**Missing**:
- Examples using readline.toml, sqlite.toml, git.toml
- Explanation of how dependencies flow through actions
- Library dependency declaration patterns

---

### Finding 5: DESIGN-dependency-provisioning.md Lacks User Summary
**Severity**: Low
**Impact**: Technical doc is intimidating for users
**Location**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/docs/DESIGN-dependency-provisioning.md`
**Line Numbers**: 1-156 (before Context section)

**Missing**:
- User-facing summary at the top
- "What This Means for Users" section
- Examples using real delivered recipes instead of hypothetical ones

---

### Finding 6: DESIGN-homebrew.md Missing M20 Examples
**Severity**: Low
**Impact**: Examples don't showcase new capabilities
**Location**: `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/docs/DESIGN-homebrew.md`
**Line Numbers**: 42-61 (example recipe), 186-211 (dependency discovery)

**Missing**:
- Examples using sqlite (simple dependency chain)
- Examples using git (complex multi-dep chain)
- Clarification of platform support for M20 recipes

---

### Finding 7: No CLI Discovery for Build Essentials
**Severity**: Medium
**Impact**: Users must read docs to discover available libraries
**Location**: CLI commands

**Missing Features**:
- `tsuku list --build-essentials` to show available libraries
- `tsuku install <tool> --dry-run` to preview dependencies
- `tsuku info <tool>` should show dependency chain

**Note**: This is a feature gap, not just documentation. Should be filed as follow-up issues.

---

## Recommendations

### Immediate Actions (Should Block Milestone Completion)

1. **Update README.md** - Add comprehensive "Build Dependency Provisioning" section with M20 examples
   - Priority: Critical
   - Effort: 1-2 hours
   - Impact: High - main user entry point

2. **Update BUILD-ESSENTIALS.md** - Complete inventory of all build essentials from M18-M20
   - Priority: High
   - Effort: 2-3 hours
   - Impact: High - reference documentation

### Short-term Actions (Complete Within 1 Week)

3. **Create GUIDE-library-dependencies.md** - User-facing guide for library dependency workflow
   - Priority: High
   - Effort: 3-4 hours
   - Impact: High - fills critical gap

4. **Update GUIDE-actions-and-primitives.md** - Add M20 recipe examples
   - Priority: Medium
   - Effort: 1-2 hours
   - Impact: Medium - improves technical docs

### Long-term Actions (Follow-up Issues)

5. **Add user summary to DESIGN-dependency-provisioning.md** - Make technical doc more accessible
   - Priority: Low
   - Effort: 1-2 hours
   - Impact: Low - mainly for contributors

6. **Update DESIGN-homebrew.md examples** - Use M20 recipes instead of jq
   - Priority: Low
   - Effort: 1 hour
   - Impact: Low - nice to have

7. **File CLI enhancement issues** - For `--dry-run`, `--build-essentials`, etc.
   - Priority: Medium
   - Effort: Variable (future work)
   - Impact: Medium - improves discoverability

---

## Impact Assessment

**User Impact**: HIGH

Without updated documentation, users will:
- Not discover that tsuku now provides readline, sqlite, git
- Continue installing system packages unnecessarily
- Miss the value proposition of dependency auto-provisioning
- Struggle to understand which libraries are available

**Contributor Impact**: MEDIUM

Without updated documentation, contributors will:
- Duplicate effort (unclear what's already implemented)
- Struggle to understand how to use the new library recipes
- Have incomplete reference material for BUILD-ESSENTIALS

**Project Impact**: HIGH

The documentation gaps undermine the achievement of M20:
- Major feature (full dependency chain validation) is invisible to users
- User-facing value of M18-M20 work is not communicated
- Growth in recipe count (154 recipes!) is not highlighted

---

## Validation Checklist

After documentation updates, verify:

- [ ] README.md mentions readline, sqlite, git recipes
- [ ] README.md explains library dependency auto-provisioning with examples
- [ ] BUILD-ESSENTIALS.md lists all recipes from M18, M19, M20
- [ ] BUILD-ESSENTIALS.md documents dependency chains (sqlite→readline→ncurses, git→curl→openssl+zlib+expat)
- [ ] New GUIDE-library-dependencies.md exists with user workflow examples
- [ ] GUIDE-actions-and-primitives.md includes M20 recipe examples
- [ ] User can understand what libraries tsuku provides without reading 1000+ line design docs
- [ ] Examples use real recipes (sqlite.toml, git.toml) not hypothetical ones

---

## Files Analyzed

**Recipes** (all exist, properly tested):
- `internal/recipe/recipes/r/readline.toml` ✓
- `internal/recipe/recipes/s/sqlite.toml` ✓
- `internal/recipe/recipes/g/git.toml` ✓

**Documentation**:
- `README.md` - Reviewed lines 1-435 ⚠️ Gaps found
- `docs/BUILD-ESSENTIALS.md` - Reviewed lines 1-85 ⚠️ Outdated
- `docs/GUIDE-actions-and-primitives.md` - Reviewed lines 1-435 ⚠️ Missing examples
- `docs/DESIGN-dependency-provisioning.md` - Reviewed lines 1-1256 ⚠️ Lacks user summary
- `docs/DESIGN-homebrew.md` - Reviewed lines 1-268 ⚠️ Missing M20 examples
- `docs/DESIGN-relocatable-library-deps.md` - Reviewed lines 1-531 ℹ️ Technical doc, no updates needed

**Testing**:
- `test/scripts/verify-tool.sh` - Reviewed ✓ Contains verify_readline, verify_sqlite, verify_git
- `test/scripts/test-readline-provisioning.sh` - Reviewed ✓ Docker-based validation
- `.github/workflows/build-essentials.yml` - Reviewed ✓ CI jobs exist

**Git History**:
- PR #661 (readline + sqlite) - Reviewed ✓
- PR #662 (git) - Reviewed ✓
- Commits 45d8d4c, dbd3eb3 - Reviewed ✓

---

## Conclusion

Milestone M20 successfully delivered the technical implementation: three recipes (readline, sqlite, git) that validate the complete dependency provisioning system. All code works and is tested in CI.

However, **documentation does not reflect this achievement**. Users cannot discover or understand the new capabilities without diving into design documents. The main README is missing critical information, and reference documentation is outdated.

**Recommendation**: Update documentation before marking M20 complete. The technical work is done; the user-facing communication is not.

**Priority Actions**:
1. Update README.md with M20 examples (1-2 hours)
2. Update BUILD-ESSENTIALS.md with complete inventory (2-3 hours)
3. Create GUIDE-library-dependencies.md (3-4 hours)

**Total Effort**: 6-9 hours of documentation work to complete the milestone.
