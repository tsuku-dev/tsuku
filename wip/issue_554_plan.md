# Implementation Plan: Issue #554 - Add curl Recipe to Validate OpenSSL Integration

## Summary

This issue validates the complete TLS stack by creating a curl recipe that builds from source using tsuku-provided openssl and zlib dependencies. The recipe will use the `configure_make` action with `setup_build_env` to configure the build environment, demonstrating that the dependency provisioning system correctly provides libraries to autotools-based builds.

**Key Validation:** This recipe proves that tsuku can provide a complete TLS-capable HTTP client without requiring any system dependencies beyond the C runtime.

## Approach

**Build Strategy:**
- Download curl 8.11.1 source tarball from curl.se
- Use `configure_make` action with `--with-openssl` and `--with-zlib` flags
- Dependencies (openssl, zlib) will be automatically provided via `setup_build_env`
- Verify TLS functionality by checking `curl --version` shows OpenSSL backend
- Add functional test to verify HTTPS requests work

**Recipe Location:** `internal/recipe/recipes/c/curl.toml`

**Dependencies:**
- openssl (already available via homebrew bottles)
- zlib (already available via homebrew bottles)
- make, zig, pkg-config (implicit dependencies from configure_make action)

## Alternatives Considered

### 1. Use Homebrew Bottles vs Build from Source

**Option A: Homebrew Bottles (NOT CHOSEN)**
- Pros: Faster installation, pre-tested
- Cons: Doesn't validate the build environment setup, which is the goal of this issue

**Option B: Build from Source (CHOSEN)**
- Pros: Validates complete build toolchain, tests dependency provisioning
- Cons: Slower installation, more complex
- **Decision:** Build from source to validate the dependency provisioning system

### 2. Which curl Version

**Option A: Latest stable (8.17.0)**
- Pros: Most recent features and security fixes
- Cons: Checksum changes frequently, may have unknown issues

**Option B: Known stable version (8.11.1)**
- Pros: Well-tested, stable checksum
- Cons: Slightly older
- **Decision:** Use 8.11.1 for stability and known compatibility

### 3. Configure Flags

**Option A: Minimal flags (--with-openssl --with-zlib)**
- Pros: Simple, focused on validation goals
- Cons: May miss some features
- **Decision:** Use minimal flags to focus on OpenSSL/zlib validation

**Option B: Full feature set**
- Pros: More complete curl installation
- Cons: More dependencies, complexity
- **Rejected:** Too complex for validation goal

## Files to Create

### 1. Recipe File
**Path:** `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/c/curl.toml`

**Content Structure:**
```toml
[metadata]
name = "curl"
description = "Command line tool for transferring data with URLs"
homepage = "https://curl.se/"
dependencies = ["openssl", "zlib"]
binaries = ["bin/curl"]

[version]
source = "static"
version = "8.11.1"

[[steps]]
action = "download_file"
url = "https://curl.se/download/curl-{version}.tar.gz"
checksum = "a889ac9dbba3644271bd9d1302b5c22a088893719b72be3487bc3d401e5c4e80"

[[steps]]
action = "extract"
archive = "curl-{version}.tar.gz"
format = "tar.gz"
strip_dirs = 1

[[steps]]
action = "setup_build_env"

[[steps]]
action = "configure_make"
source_dir = "."
configure_args = ["--with-openssl", "--with-zlib", "--disable-silent-rules"]
executables = ["curl"]

[[steps]]
action = "install_binaries"
install_mode = "directory"
binaries = ["bin/curl"]

[verify]
command = "curl --version"
pattern = "curl {version}"
```

### 2. Verification Test Update
**Path:** `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/test/scripts/verify-tool.sh`

**Add function:**
```bash
verify_curl() {
    echo "Testing: curl --version"
    curl --version
    
    # Verify OpenSSL backend
    echo ""
    echo "Testing: curl shows OpenSSL in version"
    if curl --version | grep -i openssl; then
        echo "OpenSSL backend detected"
    else
        echo "ERROR: OpenSSL not found in curl --version"
        return 1
    fi
    
    # Verify zlib support
    echo ""
    echo "Testing: curl shows zlib in version"
    if curl --version | grep -i "libz\|zlib"; then
        echo "zlib support detected"
    else
        echo "WARNING: zlib not detected in curl --version"
    fi
    
    # Test HTTPS functionality
    echo ""
    echo "Testing: curl can fetch HTTPS URL"
    if curl -sS --max-time 10 https://example.com > /dev/null; then
        echo "HTTPS request succeeded"
    else
        echo "ERROR: HTTPS request failed"
        return 1
    fi
}
```

**Add case in switch statement:**
```bash
curl)
    verify_curl
    ;;
```

### 3. CI Workflow Update
**Path:** `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/.github/workflows/build-essentials.yml`

**Add test job after test-configure-make:**
```yaml
# Test curl recipe with openssl and zlib dependencies
test-curl:
  name: "${{ matrix.platform.name }}: curl (openssl+zlib)"
  runs-on: ${{ matrix.platform.runner }}
  strategy:
    fail-fast: false
    matrix:
      platform:
        - { runner: ubuntu-latest, name: "Linux x86_64", os: linux, arch: x86_64 }
        - { runner: macos-15-intel, name: "macOS Intel", os: macos, arch: x86_64 }
        - { runner: macos-14, name: "macOS Apple Silicon", os: macos, arch: arm64 }

  steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version-file: 'go.mod'
        cache-dependency-path: go.sum

    - name: Build tsuku
      run: go build -o tsuku ./cmd/tsuku

    - name: Install curl (validates openssl+zlib chain)
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: ./tsuku install --force curl

    - name: "Verify curl functionality"
      run: ./test/scripts/verify-tool.sh curl

    - name: "Verify OpenSSL backend"
      run: |
        if curl --version | grep -i openssl; then
          echo "✓ OpenSSL backend verified"
        else
          echo "✗ OpenSSL not found in curl --version"
          exit 1
        fi

    - name: "Test HTTPS request"
      run: |
        if curl -sS --max-time 10 https://example.com > /dev/null; then
          echo "✓ HTTPS request succeeded"
        else
          echo "✗ HTTPS request failed"
          exit 1
        fi
```

## Implementation Steps

### Phase 1: Create Recipe File
1. Create `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/c/curl.toml` with the structure above
2. Verify checksum is correct: `a889ac9dbba3644271bd9d1302b5c22a088893719b72be3487bc3d401e5c4e80`
3. Set dependencies: `["openssl", "zlib"]`
4. Configure flags: `["--with-openssl", "--with-zlib", "--disable-silent-rules"]`

### Phase 2: Update Verification Script
1. Add `verify_curl()` function to `test/scripts/verify-tool.sh`
2. Test that:
   - `curl --version` shows version number
   - Output contains "OpenSSL" (validates TLS backend)
   - Output contains "zlib" or "libz" (validates compression support)
   - `curl https://example.com` succeeds (functional TLS test)

### Phase 3: Update CI Workflow
1. Add `test-curl` job to `.github/workflows/build-essentials.yml`
2. Test on 3 platforms (Linux x86_64, macOS Intel, macOS Apple Silicon)
3. Verify:
   - Recipe installs successfully
   - Verification script passes
   - OpenSSL backend present in version output
   - HTTPS requests work

### Phase 4: Local Testing
1. Build tsuku binary: `go build -o tsuku ./cmd/tsuku`
2. Install curl recipe: `./tsuku install curl`
3. Verify installation:
   - Check `~/.tsuku/tools/curl-8.11.1/bin/curl` exists
   - Run `curl --version` shows OpenSSL and zlib
   - Run `curl https://example.com` succeeds
4. Verify dependency chain:
   - Check openssl was installed to `~/.tsuku/tools/openssl-*/`
   - Check zlib was installed to `~/.tsuku/tools/zlib-*/`
   - Verify no system OpenSSL/zlib was used

### Phase 5: Integration Testing
1. Test on clean system without system curl
2. Verify tsuku-provided curl works independently
3. Test dependency resolution:
   - If openssl already installed, reuse it
   - If openssl not installed, install it automatically
4. Verify relocation works (binaries reference tsuku paths, not system paths)

## Testing Strategy

### Unit Tests
**Scope:** Recipe parsing and validation
- Verify recipe loads without errors
- Verify checksum validation passes
- Verify dependencies are correctly declared

### Functional Tests
**Scope:** Build and installation
- **Test 1:** Build curl from source
  - Download tarball
  - Extract to work directory
  - Run configure with correct flags
  - Compile and install
  - Verify binary exists at expected path

- **Test 2:** Dependency resolution
  - Verify openssl is installed first
  - Verify zlib is installed first
  - Verify PKG_CONFIG_PATH includes dependency paths
  - Verify CPPFLAGS includes dependency include paths
  - Verify LDFLAGS includes dependency lib paths

- **Test 3:** TLS functionality
  - Run `curl --version` and parse output
  - Verify "OpenSSL" appears in version string
  - Verify "zlib" or "libz" appears in features
  - Fetch `https://example.com` and verify success
  - Verify certificate validation works

### Platform Tests
**Scope:** Cross-platform compatibility
- **Linux x86_64:** Build and test on ubuntu-latest runner
- **macOS Intel:** Build and test on macos-15-intel runner
- **macOS Apple Silicon:** Build and test on macos-14 runner

**Note:** arm64_linux excluded (Homebrew doesn't publish bottles for this platform per design doc)

### Relocation Tests
**Scope:** Verify no system dependencies
- Run `verify-relocation.sh curl` to check RPATH/install_name
- Run `verify-no-system-deps.sh curl` to check linked libraries
- Verify curl binary only links to:
  - tsuku-provided openssl (libssl.so/dylib, libcrypto.so/dylib)
  - tsuku-provided zlib (libz.so/dylib)
  - System C runtime (libc, libSystem)

## Risks and Mitigations

### Risk 1: Configure Script Doesn't Detect OpenSSL
**Impact:** Build fails or falls back to different TLS backend
**Likelihood:** Medium
**Mitigation:**
- `setup_build_env` sets PKG_CONFIG_PATH to include openssl's pkgconfig directory
- Configure script should auto-detect via pkg-config
- If detection fails, check that openssl recipe installs `.pc` files
- Fallback: Use `CPPFLAGS` and `LDFLAGS` to explicitly point to openssl paths

### Risk 2: curl Requires Additional Dependencies
**Impact:** Build fails with missing library errors
**Likelihood:** Low
**Mitigation:**
- Start with minimal configure flags (only --with-openssl --with-zlib)
- Add `--without-*` flags to disable optional features that need extra deps
- Example: `--without-libpsl`, `--without-brotli`, `--without-nghttp2`
- Reference: The design doc shows this pattern is common

### Risk 3: Checksum Mismatch
**Impact:** Download fails, blocking installation
**Likelihood:** Low
**Mitigation:**
- Verify checksum is correct before committing recipe
- Downloaded and verified: `a889ac9dbba3644271bd9d1302b5c22a088893719b72be3487bc3d401e5c4e80`
- curl.se provides stable checksums that don't change

### Risk 4: HTTPS Test Fails Due to Network Issues
**Impact:** CI flakiness, false failures
**Likelihood:** Medium
**Mitigation:**
- Use reliable test URL (example.com is stable)
- Add timeout to prevent hanging (--max-time 10)
- Use `-sS` flags (silent but show errors)
- If flakiness occurs, consider mocking or using retry logic

### Risk 5: Platform-Specific Build Issues
**Impact:** Recipe works on some platforms but not others
**Likelihood:** Medium
**Mitigation:**
- Test on all 3 platforms before merging
- Check CI logs for platform-specific configure/compile errors
- Use `--disable-silent-rules` to see full compile commands
- Reference ncurses recipe which already works on all platforms

### Risk 6: Runtime Linking Issues
**Impact:** curl binary exists but crashes or can't find libraries
**Likelihood:** Low
**Mitigation:**
- Use `verify-no-system-deps.sh` to check linked libraries
- Verify RPATH/install_name points to tsuku-provided libs
- Test execution, not just compilation
- Functional test (HTTPS request) catches runtime issues

## Success Criteria

### Build Success
- [ ] Recipe downloads curl-8.11.1.tar.gz without checksum errors
- [ ] Source extracts to work directory
- [ ] Configure script runs successfully with --with-openssl and --with-zlib
- [ ] Make compiles curl without errors
- [ ] Binary installed to `$TSUKU_HOME/tools/curl-8.11.1/bin/curl`

### Dependency Integration
- [ ] openssl recipe installed automatically (if not present)
- [ ] zlib recipe installed automatically (if not present)
- [ ] setup_build_env sets PKG_CONFIG_PATH to include deps
- [ ] setup_build_env sets CPPFLAGS to include deps
- [ ] setup_build_env sets LDFLAGS to include deps
- [ ] Configure script detects openssl via pkg-config or environment flags
- [ ] Configure script detects zlib via pkg-config or environment flags

### TLS Verification
- [ ] `curl --version` shows version "curl 8.11.1"
- [ ] `curl --version` output contains "OpenSSL" (validates TLS backend)
- [ ] `curl --version` output contains "zlib" or "libz" (validates compression)
- [ ] `curl --version` shows "SSL" in Features list
- [ ] `curl https://example.com` succeeds (validates functional TLS)
- [ ] HTTPS request completes in < 10 seconds
- [ ] Certificate validation works (no --insecure needed)

### Platform Compatibility
- [ ] Recipe builds successfully on Linux x86_64
- [ ] Recipe builds successfully on macOS Intel (x86_64)
- [ ] Recipe builds successfully on macOS Apple Silicon (arm64)
- [ ] All verification tests pass on all platforms

### Relocation Verification
- [ ] curl binary has relocatable RPATH (Linux) or install_name (macOS)
- [ ] curl links only to tsuku-provided openssl and zlib
- [ ] curl does NOT link to system OpenSSL (/usr/lib, /usr/local/lib)
- [ ] curl does NOT link to system zlib
- [ ] curl works when executed from `$TSUKU_HOME/bin/curl` symlink

### CI Integration
- [ ] CI workflow includes test-curl job
- [ ] CI tests all 3 platforms
- [ ] CI verifies OpenSSL backend present
- [ ] CI performs functional HTTPS test
- [ ] CI checks pass before merge allowed

## Implementation Sequence

1. **Create Recipe** (15 min)
   - Write curl.toml with proper structure
   - Verify checksum matches downloaded tarball
   - Set dependencies and configure args

2. **Update Verification Script** (10 min)
   - Add verify_curl() function
   - Test OpenSSL and zlib detection
   - Add HTTPS functional test

3. **Update CI Workflow** (15 min)
   - Add test-curl job
   - Configure matrix for 3 platforms
   - Add OpenSSL and HTTPS verification steps

4. **Local Testing** (20 min)
   - Build tsuku binary
   - Test recipe installation
   - Verify dependencies installed
   - Check curl functionality

5. **Commit and Push** (5 min)
   - Stage changes
   - Create commit following conventional commits
   - Push to feature branch

6. **Monitor CI** (10 min)
   - Wait for CI jobs to complete
   - Check for platform-specific failures
   - Fix any issues discovered

**Total Estimated Time:** ~75 minutes

## Dependencies

**Required Before Implementation:**
- ✅ openssl recipe (Issue #552) - DONE
- ✅ zlib recipe (Issue #540) - DONE
- ✅ setup_build_env action (Issue #551) - DONE
- ✅ configure_make action - EXISTS
- ✅ make recipe (Issue #541) - DONE
- ✅ zig recipe (Issue #542) - DONE
- ✅ pkg-config recipe (Issue #548) - DONE

**Blocks:**
- Issue #557: readline recipe (depends on ncurses from M2)
- Issue #559: git recipe (depends on curl)

## References

- Design Document: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/docs/DESIGN-dependency-provisioning.md`
- ncurses Recipe: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/recipe/recipes/n/ncurses.toml`
- gdbm-source Test Recipe: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/testdata/recipes/gdbm-source.toml`
- setup_build_env Action: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/setup_build_env.go`
- configure_make Action: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/internal/actions/configure_make.go`
- Build Essentials Workflow: `/home/dgazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/.github/workflows/build-essentials.yml`

## Sources

Build configuration information gathered from:
- [How to build OpenSSL, zlib and cURL libraries on Linux](https://developers.lseg.com/en/article-catalog/article/how-to-build-openssl-and-curl-libraries-on-linux)
- [build and install curl from source](https://curl.se/docs/install.html)
- [How to cross compile CURL library with SSL and ZLIB support](https://www.matteomattei.com/how-to-cross-compile-curl-library-with-ssl-and-zlib-support/)

curl 8.11.1 SHA256 checksum verified by direct download and computation.
