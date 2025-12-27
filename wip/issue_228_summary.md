# Issue 228 Summary

## What Was Implemented
Implemented platform-aware recipe support with a complementary hybrid constraint model. Recipes can now declare OS/architecture support using allowlists and denylists. Installation fails fast with clear error messages when attempting to install on unsupported platforms.

## Changes Made
- `internal/recipe/types.go`: Added 3 platform constraint fields to MetadataSection (SupportedOS, SupportedArch, UnsupportedPlatforms)
- `internal/recipe/platform.go`: Created platform validation logic with complementary hybrid approach
- `internal/recipe/platform_test.go`: Added 34+ unit tests covering all constraint combinations and edge cases
- `cmd/tsuku/install.go`: Added preflight platform check in runDryRun()
- `cmd/tsuku/eval.go`: Added preflight platform check before executor creation
- `cmd/tsuku/info.go`: Integrated platform constraint display in both JSON and human-readable output
- `internal/recipe/recipes/b/btop.toml`: Added `supported_os = ["linux"]` constraint
- `internal/recipe/recipes/h/hello-nix.toml`: Added `supported_os = ["linux"]` constraint

## Key Decisions

**1. Complementary Hybrid Approach**
- Combines coarse allowlists (`supported_os`, `supported_arch`) with fine-grained denylist (`unsupported_platforms`)
- Computation: `(supported_os × supported_arch) - unsupported_platforms`
- Rationale: Scales to tsuku's mission of supporting "all tools in the world" where most tools work on most platforms
- Default behavior (nil fields) = universal support for backwards compatibility

**2. Nil vs Empty Slice Semantics**
- `nil` = use defaults (all OS, all arch, no exceptions)
- `[]` = explicit empty set (no platforms)
- Rationale: Allows authors to explicitly restrict recipes while maintaining backwards compatibility

**3. Preflight Validation**
- Check platform before executor creation (in install and eval commands)
- Rationale: Fail-fast UX without wasted downloads, work directory creation, or dependency resolution

**4. Simple Error Messages**
- Show current platform and constraints (allowlist + denylist)
- No alternative suggestions or upstream links
- Rationale: Ship quickly with sufficient information; enhancements can be added later without breaking changes

## Trade-offs Accepted

**1. Validation complexity**
- Must compute Cartesian product and subtract exceptions
- Acceptable: Implementation detail, not user-facing; computation is fast

**2. Edge case validation required**
- Must detect no-op constraints (warnings) and empty result sets (errors)
- Acceptable: Follows existing preflight warning pattern; prevents broken recipes

**3. Allowlists can become stale**
- If new platforms are added to Go runtime, recipes with explicit `supported_os`/`supported_arch` won't automatically support them
- Acceptable: New platforms added infrequently (years between new GOOS/GOARCH); CI tests on new platforms will reveal gaps

**4. No alternative suggestions**
- Error messages don't suggest alternative tools
- Acceptable: Users can run `tsuku search` manually; keeps initial implementation simple; can enhance later

## Test Coverage
- **New tests added**: 34 test cases
  - 16 for `SupportsPlatform()` covering all constraint patterns
  - 7 for `ValidatePlatformConstraints()` edge cases
  - 4 for `GetSupportedPlatforms()`
  - 4 for `FormatPlatformConstraints()`
  - 3 for `UnsupportedPlatformError` formatting
- **Coverage**: All new code paths covered with comprehensive test cases
- **Build status**: All tests passing, go vet clean

## Known Limitations

**1. Cannot express complex version-specific constraints**
- Cannot specify "macOS 12+" or "Linux kernel 5.10+"
- Mitigation: Can be added as separate fields if use cases emerge

**2. No force-install override**
- Users cannot bypass platform check to attempt installation anyway
- Mitigation: Can add `--force` flag in future if needed

**3. Gradual rollout**
- Existing recipes without constraints won't benefit until updated
- Mitigation: Start with known-failing recipes (btop, hello-nix); document pattern for community contributions

**4. No automatic platform detection**
- Recipe authors must manually add constraints
- Mitigation: Document pattern; CI can eventually warn when recipes fail only on certain platforms but lack metadata

## Future Improvements
- Add platform constraint display to website (deferred to separate issue)
- CI test matrix integration to automatically skip unsupported combinations
- Consider `--force` flag to bypass platform check for advanced users
- Enhance error messages with alternative tool suggestions (requires similarity/tagging system)
- Add support for minimum OS version constraints if use cases emerge
