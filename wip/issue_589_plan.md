# Issue 589 Implementation Plan

## Summary

Consolidate DESIGN-homebrew-builder.md and DESIGN-homebrew-cleanup.md into a single authoritative DESIGN-homebrew.md document describing the current state of Homebrew integration in tsuku (bottle-only approach with LLM-based builder).

## Approach

Create a new consolidated design document that merges relevant content from both existing docs, focusing on the current state (bottles only) while preserving historical context about the scope reduction decision. Update cross-references in other design docs that mention homebrew_bottle (which is now renamed to homebrew in the codebase).

## Files to Create

- `docs/DESIGN-homebrew.md` - Consolidated Homebrew design document describing:
  - The `homebrew` action (bottle installation)
  - The HomebrewBuilder (LLM-based bottle recipe generation)
  - Platform support and bottle tag mapping
  - Dependency discovery and generation workflow
  - Security considerations
  - Historical context about source build removal

## Files to Delete

- `docs/DESIGN-homebrew-builder.md` - Content merged into DESIGN-homebrew.md
- `docs/DESIGN-homebrew-cleanup.md` - Work complete, content merged into DESIGN-homebrew.md

## Files to Modify

- `docs/DESIGN-dependency-provisioning.md` - Update references from `homebrew_bottle` to `homebrew`
- `docs/DESIGN-relocatable-library-deps.md` - Update references from `homebrew_bottle` to `homebrew`

## Implementation Steps

- [ ] Review both existing design docs to identify content to consolidate
- [ ] Create `docs/DESIGN-homebrew.md` with consolidated content covering:
  - Current state: `homebrew` action (renamed from `homebrew_bottle`)
  - HomebrewBuilder architecture and workflow
  - Platform support (macOS ARM64/x86_64, Linux ARM64/x86_64)
  - Bottle tag mapping at runtime
  - Dependency discovery and tree traversal
  - LLM conversation loop and tools
  - Validation and repair loop
  - Security considerations (bottles only, no source builds)
  - Historical note about source build removal and rationale
- [ ] Update `docs/DESIGN-dependency-provisioning.md` to use `homebrew` instead of `homebrew_bottle`
- [ ] Update `docs/DESIGN-relocatable-library-deps.md` to use `homebrew` instead of `homebrew_bottle`
- [ ] Delete `docs/DESIGN-homebrew-builder.md`
- [ ] Delete `docs/DESIGN-homebrew-cleanup.md`
- [ ] Verify no broken references remain in docs/

## Success Criteria

- [ ] `docs/DESIGN-homebrew.md` exists and contains consolidated information
- [ ] Document accurately describes current state (bottles only, `homebrew` action)
- [ ] Document includes historical context about source build removal
- [ ] All cross-references to `homebrew_bottle` updated to `homebrew`
- [ ] No references to deleted design docs remain
- [ ] `docs/DESIGN-homebrew-builder.md` and `docs/DESIGN-homebrew-cleanup.md` are deleted

## Open Questions

None - this is straightforward documentation consolidation.
