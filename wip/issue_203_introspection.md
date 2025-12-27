# Issue 203 Introspection

## Staleness Signals

```json
{
  "introspection_recommended": true,
  "issue": {
    "number": 203,
    "title": "feat(verify): post-install checksum pinning (Layer 3)",
    "created_at": "2025-12-06T20:06:57Z",
    "age_days": 20,
    "milestone": "Defense-in-Depth Verification"
  },
  "signals": {
    "issue_age_days": 20,
    "sibling_issues_closed_since_creation": 2,
    "milestone_position": "middle",
    "files_modified_since_creation": [
      "docs/DESIGN-version-verification.md"
    ]
  }
}
```

## Assessment

### Spec Validity
The issue spec remains valid. The upstream design (DESIGN-version-verification.md Phase 6) was reviewed and found to provide accurate high-level direction.

### Recent Changes Impact
- `docs/DESIGN-version-verification.md` was modified since issue creation
- Review of changes: No material impact on issue 203 scope
- Layer 2 (Version Verification) was completed (sibling issues #196-#201)
- This does not affect Layer 3 implementation

### Design Document Created
A detailed tactical design was just created: `docs/DESIGN-checksum-pinning.md`

This design:
- Analyzed current `VersionState` schema (found `Plan.Steps[].Checksum` already stores download checksums)
- Identified 6 decision points not covered in upstream design
- Made explicit choices with documented tradeoffs
- Defined implementation phases

### Key Decisions Made
1. Checksum **binaries only** (not all files) - already tracked in `VersionState.Binaries`
2. Compute **after all actions complete** - captures final state
3. Store in `VersionState.BinaryChecksums map[string]string` - simple, backward compatible
4. **Graceful verification** - old installations without checksums show "SKIPPED"
5. Use **SHA256** - already used throughout codebase

## Recommendation: Proceed

The issue spec is valid and has been elaborated into a detailed design document that was approved by the user. No clarification or amendment needed.

## Files to Modify
Based on design analysis:
- `internal/install/state.go` - Add `BinaryChecksums` field
- `internal/install/checksum.go` (new) - Checksum computation functions
- `internal/install/manager.go` - Integration point
- `cmd/tsuku/verify.go` - Verification step
- Tests for all above
