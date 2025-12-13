# Security Review: Plan-Based Installation

## Executive Summary

This security review analyzes the plan-based installation feature for tsuku, which enables installing tools from pre-computed JSON plans. The feature allows offline installation and reproducible builds but introduces new attack vectors around plan file trust.

**Overall Assessment**: The security posture is **ADEQUATE** for the intended use cases with **MODERATE residual risk** that users must accept. Critical mitigations are in place, but several attack vectors warrant deeper consideration.

**Key Findings**:
1. One critical gap: Time-of-check/time-of-use (TOCTOU) vulnerability in download verification
2. Several "not applicable" justifications need reconsideration
3. Residual supply chain risk is correctly identified but underestimated
4. Missing consideration of plan lifetime security implications
5. Cache security assumptions need hardening

## Detailed Analysis

### 1. Attack Vectors Assessment

#### 1.1 Identified and Mitigated

##### Download Verification (Checksums)
**Status**: Partially mitigated with critical gap

**Current Mitigations**:
- All downloads verify SHA256 checksums against plan
- ChecksumMismatchError fails installation with clear message
- HTTPS enforcement on all URLs

**CRITICAL GAP IDENTIFIED**: Time-of-check/time-of-use (TOCTOU) vulnerability
- **Issue**: The current implementation downloads the file FIRST (via `download` action), THEN computes checksum for verification
- **Code evidence**: `executor.go:324-326` shows download executes, then verification happens afterward
- **Attack scenario**:
  1. Plan specifies legitimate URL with expected checksum
  2. Between download completion and checksum verification, attacker replaces file in work directory
  3. Verification fails, but malicious file already exists on disk
  4. If verification failure handling is weak, partial state remains

**Recommendation**:
- Compute checksum during download (streaming hash), not after
- Use atomic operations: only move to final destination after verification passes
- Already implemented in PreDownloader (`predownload.go:122-123` uses `io.TeeReader`) but NOT in ExecutePlan flow
- **ACTION REQUIRED**: Refactor `executeDownloadWithVerification` to use streaming verification

##### Execution Isolation
**Status**: Adequate

**Current Mitigations**:
- Work directories created with mode 0700 (good)
- Only primitive actions allowed
- Actions constrained to $TSUKU_HOME

**Additional Consideration**:
- Work directory is in `/tmp` (mode 0700 is correct)
- But cleanup on failure? Review needed
- Partial installations leave artifacts in work directory

#### 1.2 Newly Identified Attack Vectors

##### 1.2.1 Plan Substitution Attack
**Vector**: Attacker substitutes plan file during `tsuku install --plan` execution

**Attack Scenario**:
1. User runs `tsuku install --plan plan.json`
2. Plan is loaded and validated
3. Between validation and execution, attacker modifies plan.json
4. Modified plan executes with different steps

**Current Mitigation**: NONE - plan is read once, parsed, validated, then executed
**Risk Level**: LOW (requires local file access during execution window)
**Recommendation**:
- Plan is already loaded into memory before execution starts
- Document that plan files should not be on shared/writable filesystems
- Consider adding plan hash to execution context for audit trail

##### 1.2.2 Symlink Attack on Cache Directory
**Vector**: Attacker creates symlink in cache directory to redirect writes

**Attack Scenario**:
1. Attacker creates symlink: `$TSUKU_HOME/cache/downloads/HASH.data -> /victim/file`
2. User runs plan-based installation
3. Download overwrites victim file

**Current Mitigation**: YES - `download_cache.go:54-59` checks for symlinks with `containsSymlink()`
**Risk Level**: LOW (already mitigated)
**Status**: ADEQUATE

##### 1.2.3 Cache Poisoning via Race Condition
**Vector**: Multiple concurrent installations race to populate cache

**Attack Scenario**:
1. User A starts install, begins downloading legitimate file to cache
2. User B (attacker) simultaneously downloads malicious file with same URL hash
3. Race condition: which file wins the cache write?

**Current Mitigation**: Atomic write pattern in `download_cache.go:146-154`
- Writes to `.tmp` file first
- Renames to final location (atomic on POSIX)
**Risk Level**: LOW (atomic rename prevents corruption)
**Status**: ADEQUATE

##### 1.2.4 Plan Format Version Downgrade Attack
**Vector**: Attacker provides plan with older format version to bypass new security checks

**Current Mitigation**: `ValidatePlan` rejects format version < 2 (`plan.go:179-185`)
**Risk Level**: LOW (already mitigated)
**Status**: ADEQUATE

However, **FUTURE RISK**: As new format versions are added, compatibility logic may inadvertently allow downgrades.
**Recommendation**: Document that format version must NEVER go backward, only forward

##### 1.2.5 Dependency Confusion via Plans
**Vector**: Plan specifies URLs for dependencies that shadow legitimate tools

**Attack Scenario**:
1. Plan for `tool-a` includes steps that download binaries named like other tools
2. If binaries are placed in PATH, they shadow legitimate tools
3. User runs what they think is `git` but it's attacker's binary from `tool-a` plan

**Current Mitigation**:
- Plans only specify installation for ONE tool
- Binaries go to tool-specific directory: `$TSUKU_HOME/tools/<tool>-<version>/`
- Symlinks from `$TSUKU_HOME/bin/` are controlled by `install_binaries` action
**Risk Level**: MEDIUM - depends on recipe validation
**Gap**: Plan validation doesn't check that `install_binaries` names match tool metadata
**Recommendation**:
- Add validation: warn if `install_binaries` specifies names that differ from tool name
- Not a hard error (some tools install multiple binaries), but should be auditable

##### 1.2.6 URL Redirection Attack
**Vector**: HTTPS URL redirects to malicious content

**Attack Scenario**:
1. Plan specifies `https://releases.example.com/v1.0.0/tool.tar.gz` with checksum X
2. Attacker compromises DNS or performs BGP hijacking
3. URL redirects to malicious server
4. Malicious server returns content with checksum X (pre-computed to match)

**Current Mitigation**:
- Checksum verification prevents content tampering
- HTTPS prevents in-transit tampering
- SSRF protection in `httputil.NewSecureClient()`
**Risk Level**: LOW (checksum verification defeats this)
**Status**: ADEQUATE

BUT: **What if attacker controls the plan generation environment?**
- If `tsuku eval` runs on compromised system, generated plan will have malicious URL + matching checksum
- This is the "Upstream Trust Inheritance" assumption
- Risk is correctly identified in assumptions section

##### 1.2.7 Offline Mode Cache Trust Exploitation
**Vector**: In offline mode, cached artifacts are trusted solely on checksum

**Attack Scenario**:
1. Attacker gains access to user's `$TSUKU_HOME/cache/downloads/`
2. Replaces cached files with malicious content
3. Computes new checksums and updates `.meta` files
4. User runs offline installation, trusts cache

**Current Mitigation**:
- Cache directory created with mode 0700 (user-only access)
- `download_cache.go:127` sets secure permissions
**Risk Level**: MEDIUM (depends on filesystem permissions being maintained)
**Gap**: No integrity protection for cache metadata
**Recommendation**:
- Cache .meta files should be signed or use HMAC
- Key derived from $TSUKU_HOME permissions (prevents tampering without user access)
- **Defer to future work** but document as known limitation

##### 1.2.8 Resource Exhaustion via Plan Size
**Vector**: Malicious plan specifies enormous downloads to exhaust disk space

**Attack Scenario**:
1. Attacker provides plan with 100+ download steps
2. Each download is 1GB+
3. Plan execution fills user's disk

**Current Mitigation**: NONE
**Risk Level**: LOW-MEDIUM (denial of service, not privilege escalation)
**Existing Controls**:
- PreDownloader supports context cancellation (user can Ctrl+C)
- No automatic disk space checks
**Recommendation**:
- Document as user responsibility to review plans before execution
- Consider adding `--dry-run` mode for plans to show what would be downloaded
- Defer disk quota enforcement to future work

##### 1.2.9 Template Variable Injection
**Vector**: Plan params contain unchecked template variables that expand to malicious paths

**Attack Scenario**:
1. Malicious plan includes step with params: `{"file": "{version}/../../../etc/passwd"}`
2. Template expansion in `ExpandVars` creates path traversal
3. Action operates on unintended file

**Current Mitigation**:
- Plans store RESOLVED steps, not templates
- Template expansion happens during `tsuku eval`, not during plan execution
- ExecutePlan receives pre-expanded parameters
**Risk Level**: LOW (templates already resolved in plans)
**Gap Check**: Verify that plan execution does NOT re-expand templates
- Reviewing `executor.go:280-338`: ExecutePlan creates ExecutionContext with version from plan, passes step.Params directly to actions
- **CONFIRMED**: No re-expansion during plan execution
**Status**: ADEQUATE

##### 1.2.10 Plan Lifetime Expiration Attack
**Vector**: Long-lived plans reference URLs that have been compromised over time

**Attack Scenario**:
1. Team generates plan for `tool@1.0.0` in January 2025
2. Plan stored in repository, used for months
3. In June 2025, upstream GitHub repository is compromised
4. Attacker re-tags release to serve malicious content
5. Team uses old plan with stale checksum

**Current Mitigation**: Checksum verification prevents this
- If upstream content changes, checksum won't match
- Installation fails with ChecksumMismatchError
**Risk Level**: LOW (checksums prevent execution)
**But**: Creates operational burden - plans fail when upstreams change
**Assumption**: "Plans are intended for short-term use (hours to days for CI workflows), not long-term archival"
**Status**: ADEQUATE for stated scope
**Recommendation**: Document plan expiration expectations clearly

### 2. Mitigation Sufficiency Analysis

#### Download Verification
**Status**: Insufficient - TOCTOU gap

**Recommendation**: HIGH PRIORITY fix
- Refactor `executeDownloadWithVerification` to compute checksum during download
- Model after PreDownloader pattern (streaming hash with TeeReader)
- Ensure atomic write-after-verify

#### Execution Isolation
**Status**: Adequate with minor gaps

**Current State**:
- Mode 0700 on work directories ✓
- Primitive-only actions ✓
- $TSUKU_HOME containment ✓

**Minor Gap**: Cleanup on failure
- If ExecutePlan fails mid-execution, work directory may contain partial artifacts
- Not a security issue (mode 0700 prevents other users), but could leak disk space
**Recommendation**: Ensure `Cleanup()` is called via defer

#### Supply Chain Risks
**Status**: Correctly identified, mitigations are minimal

**Assessment**: The design correctly states this is "residual risk" that users must accept
- Plans are trusted input (like running a script)
- No automatic verification of plan source
- Future enhancement: plan signing

**But**: This deserves MORE emphasis in documentation
- Users must treat plans like executable code
- Plans should be reviewed before use
- Plans from untrusted sources are dangerous

**Recommendation**:
- Add explicit warning in `tsuku install --plan --help`
- Add confirmation prompt for plans from outside $TSUKU_HOME (optional)
- Document plan verification workflow (e.g., how to audit a plan)

#### User Data Exposure
**Status**: Adequate

No credentials in plans ✓
Public URLs only ✓
No new exposure beyond normal install ✓

#### Cache Security
**Status**: Good with minor gap

**Strengths**:
- Mode 0700 on cache directory
- Symlink checks before read/write
- Atomic write pattern
- Permission validation

**Gap**: Cache metadata integrity
- .meta files are JSON, not signed
- Attacker with user access can modify metadata
- Not a privilege escalation (already has user access)
- But could cause cache corruption or bypass checksum checks

**Recommendation**: Defer to future work, document limitation

### 3. Residual Risk Assessment

#### Correctly Identified Residual Risks

1. **Supply Chain - Malicious Plans**: Users who accept plans from untrusted sources may install malicious software
   - **Assessment**: CORRECT
   - **Severity**: HIGH
   - **User Decision**: Yes, but needs better documentation

2. **Upstream Trust Inheritance**: Plans generated from compromised recipe inherit compromise
   - **Assessment**: CORRECT
   - **Severity**: HIGH
   - **Mitigation**: None technical, relies on recipe integrity

3. **Cached Artifact Trust**: In offline mode, no external verification possible
   - **Assessment**: CORRECT
   - **Severity**: MEDIUM
   - **Mitigation**: Cache permissions, but no integrity checks

#### Underestimated Residual Risks

1. **Plan Verification Burden**: Users are expected to "review plans before execution"
   - **Issue**: Plans are JSON with resolved steps - not human-friendly
   - **Gap**: No tooling to audit plans (e.g., `tsuku plan audit <file>` to show what it does)
   - **Recommendation**: Provide plan auditing tools, not just documentation

2. **Multi-Tool Installations**: Design explicitly excludes multi-tool plans
   - **But**: Complex tools may require dependencies
   - **Gap**: Scope states "dependency resolution from plans (plan-based installation assumes dependencies are already installed via normal flow)"
   - **Risk**: User might accept a plan for "tool-a" that fails because "tool-b" dependency is missing
   - **Not a security issue**, but UX problem that could lead to installing from untrusted sources out of frustration

3. **Plan Format Evolution**: As format versions increase, backward compatibility could weaken security
   - **Current**: Format version 2 required
   - **Risk**: Future versions might relax validation rules for compatibility
   - **Recommendation**: Formal policy: format versions only add restrictions, never remove them

### 4. "Not Applicable" Justification Review

#### Download Verification - NOT APPLICABLE?
**Claim**: "Analysis: All downloads during plan execution verify checksums against the plan."

**Review**: This is APPLICABLE and there's a gap
- Verification happens AFTER download completes
- TOCTOU vulnerability between download and verification
- **Change status**: APPLICABLE, needs improvement

#### Execution Isolation - Work Directory Containment
**Claim**: "Actions are constrained to $TSUKU_HOME directory structure"

**Review**: Not entirely accurate
- Work directory is in `/tmp`, not $TSUKU_HOME
- Actions receive ExecutionContext with `WorkDir` (in /tmp) and `InstallDir` ($TSUKU_HOME)
- Primitive actions could potentially write outside these if implemented incorrectly
- **Verification needed**: Audit all primitive actions for path validation

**Recommendation**:
- Add path validation helper: `validatePathWithinDir(path, allowedDir)`
- All file-writing primitives should use this
- Not a critical gap (primitives are internal code), but defense in depth

#### Supply Chain - HTTPS Enforcement
**Claim**: "HTTPS enforcement on all URLs (existing download action protection)"

**Review**: CONFIRMED
- `download.go:166-168`: Rejects non-HTTPS URLs
- `predownload.go:56-58`: Same enforcement
- **Status**: ADEQUATE

#### User Data Exposure - "No additional data exposure beyond normal installation"
**Claim**: Plans don't expose more data than normal install

**Review**: MOSTLY TRUE, but one consideration
- Plans contain `recipe_hash` and `recipe_source`
- If `recipe_source` is a file path, it could leak information about user's filesystem structure
- Example: `"recipe_source": "/home/alice/secret-project/recipes/tool.toml"`
- **Risk**: LOW (plans are trusted input, not public)
- **Recommendation**: Sanitize `recipe_source` in plan export to use basename only

### 5. Security Assumptions Validation

#### Assumption 1: "Plans are intended for short-term use (hours to days for CI workflows)"
**Validation**: REASONABLE

**Implications**:
- Checksums may become stale as upstreams change (not a security issue, just operational)
- Recipe format evolution could invalidate old plans
- Security advisories won't apply to frozen plans

**Risk**: Users may archive plans for long-term use despite this assumption
**Recommendation**:
- Add `generated_at` timestamp check (warn if plan > 30 days old)
- Optional flag: `--allow-stale-plan` to override

#### Assumption 2: "Plans are strictly platform-specific"
**Validation**: CORRECT

**Enforcement**: `validateExternalPlan` checks platform match (design doc section 358)
**Status**: ADEQUATE

#### Assumption 3: "Cached artifacts trusted based on checksum verification alone"
**Validation**: REASONABLE but risky in offline mode

**Gap**: If attacker has user-level access, can poison cache
**Mitigation**: Mode 0700 on cache directory
**Residual Risk**: User-level compromise allows cache tampering
**Status**: ADEQUATE for threat model (assumes user environment is secure)

#### Assumption 4: "Upstream trust inheritance"
**Validation**: CORRECT and critical

**This is the CORE security assumption**: Plans are only as trustworthy as their generation environment

**Implications**:
- Compromised build server generates compromised plans
- Compromised developer machine generates compromised plans
- No technical mitigation (signing deferred to future work)

**Recommendation**: Document plan generation security:
- Generate plans on trusted, isolated systems
- Treat plan generation like code signing ceremony
- Store plans in version control for audit trail

### 6. Missing Security Considerations

#### 6.1 Plan Provenance and Audit Trail
**Gap**: No mechanism to track where a plan came from

**Use Cases**:
- Post-incident forensics: "Which plans were compromised?"
- Compliance: "Can we prove this plan came from our CI system?"

**Current State**: Plan has `generated_at` timestamp and `recipe_hash`, but no:
- Generator identity
- Build system metadata
- Chain of custody

**Recommendation**:
- Add optional `metadata` field to plan format
- Include: hostname, username, CI job ID, etc.
- Not required for security, but aids auditability

#### 6.2 Plan Signing (Deferred)
**Status**: Correctly identified as future enhancement

**Current Risk**: Anyone can craft a valid-looking plan
**Mitigation**: User review (but see burden discussion above)

**Recommendation for future work**:
- Asymmetric signatures (GPG, signify, minisign)
- Key distribution via $TSUKU_HOME/trusted-keys/
- Optional enforcement: `--require-signed-plans`

#### 6.3 Checksum Algorithm Agility
**Current State**: SHA256 hardcoded in many places

**Future Risk**: If SHA256 is broken, migration path is unclear
**Recommendation**:
- Plan format already has `FormatVersion` for evolution
- Document that checksum algorithm changes require format version bump
- Keep SHA256 for now (no known practical attacks)

#### 6.4 Denial of Service via Action Composition
**Vector**: Plan with many extract/download steps consumes resources

**Example**:
```json
{
  "steps": [
    {"action": "download", "params": {"url": "https://example.com/1.tar.gz"}},
    {"action": "extract", "params": {"src": "1.tar.gz"}},
    {"action": "download", "params": {"url": "https://example.com/2.tar.gz"}},
    {"action": "extract", "params": {"src": "2.tar.gz"}},
    // ... repeated 1000 times
  ]
}
```

**Current Mitigation**: User can Ctrl+C
**Risk Level**: LOW (annoyance, not privilege escalation)
**Recommendation**: Document that plan review should check step count

#### 6.5 Binary Verification Post-Installation
**Gap**: After installation, no ongoing verification of binaries

**Use Case**: Detect if installed binaries are modified after installation
**Current State**:
- state.json stores plan (includes checksums of archives)
- But no checksums of final extracted binaries

**Future Enhancement**:
- `tsuku verify <tool>` could re-run plan to compute expected state
- Compare against installed files
- Detect tampering

**Defer to future work**

### 7. Threat Model Validation

#### Assumed Threat Model (Inferred)

**In Scope**:
- Network attacker (MITM, compromised mirrors)
- Untrusted plan sources
- Multi-user systems (cache isolation)

**Out of Scope**:
- User-level compromise (attacker with user's permissions)
- Kernel-level attacks
- Hardware attacks

**Assessment**: REASONABLE for a package manager

**But**: One gray area - Local privilege escalation
- tsuku runs in user space, no sudo
- But if a compromised tool is installed, it inherits user privileges
- This is inherent to any package manager, not specific to plan-based installation

**Clarification Needed**: Document threat model explicitly
- What attackers are we defending against?
- What attacks are user responsibility?

### 8. Recommendations Summary

#### Critical (Fix Before Release)

1. **Fix TOCTOU in Download Verification**
   - Compute checksum during download, not after
   - Use streaming verification (TeeReader pattern)
   - Atomic write-after-verify

#### High Priority (Should Address)

2. **Enhanced Plan Security Warnings**
   - Add explicit warning to `--plan` flag help text
   - Document plan review workflow
   - Consider confirmation prompt for external plans

3. **Plan Audit Tooling**
   - Implement `tsuku plan audit <file>` command
   - Human-readable summary of plan actions
   - Security-relevant information highlighted (URLs, binaries to install)

4. **Path Validation Hardening**
   - Audit all primitive actions for path traversal
   - Add `validatePathWithinDir` helper
   - Ensure all file operations are contained

#### Medium Priority (Improve Over Time)

5. **Plan Lifetime Warnings**
   - Check `generated_at` timestamp
   - Warn on stale plans (> 30 days)
   - Optional `--allow-stale-plan` flag

6. **Cache Metadata Integrity**
   - Sign or HMAC cache .meta files
   - Detect tampering in multi-user scenarios
   - Defer if single-user systems only

7. **Plan Provenance Metadata**
   - Add optional metadata field to plans
   - Record generation context (hostname, CI job, etc.)
   - Aid forensics and audit

8. **Sanitize recipe_source in Plans**
   - Use basename only, not full path
   - Prevent leaking filesystem structure

#### Low Priority / Future Work

9. **Plan Signing** (already deferred)
10. **Disk Quota Enforcement**
11. **Binary Verification Post-Install**

## Conclusion

The plan-based installation security model is **fundamentally sound** but has **one critical gap** (TOCTOU in verification) and several areas for hardening.

**Risk Level**: MEDIUM-HIGH without TOCTOU fix, MEDIUM with fix

**Residual Risk**: Correctly identified, but documentation must emphasize that plans are executable code and must be treated with corresponding caution.

**Release Readiness**:
- **Block on**: TOCTOU fix (#1)
- **Strongly recommend**: Enhanced warnings and audit tooling (#2-3)
- **Nice to have**: Other recommendations

**Key Insight**: The security model shifts trust from "recipes + upstream URLs" to "pre-computed plans". This is a significant change in the trust model that needs clear user communication. Users must understand that accepting a plan is like running a script - review is essential.
