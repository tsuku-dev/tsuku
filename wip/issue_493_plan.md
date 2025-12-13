# Issue 493 Implementation Plan

## Summary

Add resource and patch support to the Homebrew builder's source build mode. Resources are additional downloads (like tree-sitter parsers for neovim), patches are source modifications from the formula-patches repo or inline DATA sections, and inreplace operations are Homebrew-specific text replacements.

## Approach

Extend the existing `extract_source_recipe` LLM tool and source recipe generation to include resources, patches, and inreplace mappings. The LLM parses the Ruby formula and extracts these elements, which are then converted to recipe steps that download, stage, and apply modifications before the build.

### Alternatives Considered
- **Parse Ruby with a Go Ruby parser**: More deterministic but extremely complex given Homebrew's DSL. Ruby AST libraries for Go are immature. The LLM approach aligns with the existing architecture.
- **Only support resources, not patches**: Would limit coverage. Many formulas use patches to fix platform-specific issues. Implementing all three together is more cohesive.

## Files to Modify

- `internal/builders/homebrew.go` - Extend sourceRecipeData struct, add LLM tool for resource/patch extraction, update recipe generation
- `internal/builders/homebrew_test.go` - Add unit tests for resource/patch parsing and validation
- `internal/recipe/types.go` - Add Resource and Patch types, extend ToTOML to serialize them

## Files to Create

- `internal/actions/resource_stage.go` - Action to download and stage resources to specified directories
- `internal/actions/resource_stage_test.go` - Tests for resource staging action
- `internal/actions/apply_patch.go` - Action to apply patches using `patch` command
- `internal/actions/apply_patch_test.go` - Tests for patch application action
- `internal/actions/text_replace.go` - Action to perform inreplace-style text substitutions
- `internal/actions/text_replace_test.go` - Tests for text replacement action

## Implementation Steps

- [x] Add Resource and Patch types to `internal/recipe/types.go`
- [x] Extend ToTOML to serialize resources and patches
- [x] Add tests for Resource and Patch serialization in `types_test.go`
- [ ] Extend `sourceRecipeData` struct in `homebrew.go` to include resources, patches, and inreplace
- [ ] Update `extract_source_recipe` tool definition to accept resources, patches, and inreplace
- [ ] Validate resource/patch data in `validateSourceRecipeData()`
- [ ] Add `resource_stage` action for downloading and staging resources
- [ ] Add tests for `resource_stage` action
- [ ] Add `apply_patch` action for applying patches
- [ ] Add tests for `apply_patch` action
- [ ] Add `text_replace` action for inreplace operations
- [ ] Add tests for `text_replace` action
- [ ] Update `buildSourceSteps()` to emit resource, patch, and inreplace steps
- [ ] Add unit tests for resource/patch extraction in `homebrew_test.go`
- [ ] Run full test suite and verify build
- [ ] Run golangci-lint to catch any issues

Mark each step [x] after it is implemented and committed.

## Testing Strategy

- **Unit tests**: Test each new action in isolation (resource_stage, apply_patch, text_replace)
- **Recipe generation tests**: Verify that sourceRecipeData with resources/patches generates correct steps
- **TOML round-trip tests**: Verify Resource/Patch types serialize and deserialize correctly
- **Manual verification**: Generate a recipe for neovim (has resources) to validate end-to-end

## Risks and Mitigations

- **patch command availability**: The `patch` utility may not be installed. Mitigation: Check for patch availability and provide helpful error message.
- **Resource URL validation**: Resources could point to arbitrary URLs. Mitigation: Validate URLs against allowlist (GitHub, official sources).
- **Inreplace complexity**: Some inreplace operations use regex. Mitigation: Support simple text replacement first, mark complex patterns as unsupported.

## Success Criteria

- [ ] Resources are extracted from Ruby formulas and staged before build
- [ ] Patches (both URL-based and inline) are downloaded and applied
- [ ] Inreplace operations are mapped to text_replace actions
- [ ] Unit tests pass for all new code
- [ ] Integration test with neovim formula succeeds (has 6 resources)
- [ ] Build passes, no golangci-lint errors

## Open Questions

None blocking - the design is clear from the design document.
