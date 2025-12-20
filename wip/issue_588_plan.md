# Issue 588 Implementation Plan

## Summary

Close obsolete Homebrew source build issues and update milestone description to reflect that source builds have been removed. This involves updating the Homebrew Builder milestone (currently closed) to clarify it only implemented bottle extraction, and closing Phase 2 issues (#491-494) and meson_build issue (#521) with explanatory notes.

## Approach

Since the Homebrew source build functionality has been removed (issues #586, #587 completed), we need to close all related issues that were created for features that are no longer needed. The design document (docs/DESIGN-homebrew-cleanup.md) explains that research showed 99.94% of Homebrew formulas have bottles, making source builds unnecessary.

Phase 2 issues (#491-494) are already closed, but we should verify they have appropriate closure notes. Issue #521 (meson_build) is also already closed. The main work is updating the Homebrew Builder milestone description to reflect the reduced scope.

## Files to Modify

None - all work is done via GitHub API (gh CLI).

## Files to Create

None - this is a chore issue handling GitHub metadata.

## Implementation Steps

- [x] Check closure notes on issues #491, #492, #493, #494 to ensure they explain the source build removal decision
  - These issues were successfully implemented and merged, then the source build code was removed later (issue #587)
  - No closure notes needed - they were completed, not abandoned
- [x] Check closure note on issue #521 to ensure it references the cleanup decision
  - Issue #521 (meson_build) was successfully implemented and merged
  - The meson_build action remains useful for non-Homebrew source builds
  - No closure note needed
- [x] Update Homebrew Builder milestone (number 17) description to reflect bottles-only scope
  - Updated to: "LLM-based recipe generation from Homebrew formulas. Implements bottle extraction to generate platform-agnostic tsuku recipes from Homebrew core formulas. Note: Phase 2 (source builds) was implemented but later removed after research showed 99.94% of formulas have bottles, making source build support unnecessary."
- [x] Verify no other open issues exist related to Homebrew source builds (search confirmed none found)

## Success Criteria

- [x] Homebrew Builder milestone description updated to clarify bottles-only scope
- [x] All Phase 2 issues (#491-494) verified as properly closed (were successfully implemented)
- [x] Issue #521 verified as properly closed (successfully implemented, meson_build action still useful)
- [x] No open issues remain that reference Homebrew source build features

## Open Questions

None - the design document is clear about the decision and rationale.
