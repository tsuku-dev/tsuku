# Issue 493 Summary

## What Was Implemented

Added resource, patch, and inreplace support to the Homebrew builder's source build mode, along with the corresponding execution actions.

## Changes Made

- `internal/recipe/types.go`: Added Resource, Patch, and TextReplace types with TOML serialization
- `internal/recipe/types_test.go`: Added tests for Resource/Patch unmarshaling and ToTOML roundtrip
- `internal/builders/homebrew.go`:
  - Extended sourceRecipeData with resources, patches, and inreplace fields
  - Updated extract_source_recipe tool definition with new parameters
  - Added validation functions for resource/patch/inreplace data
  - Updated LLM system prompt with guidance on extracting resources/patches
  - Modified generateSourceRecipeOutput to convert data to recipe fields
  - Modified buildSourceSteps to emit text_replace steps for inreplace
- `internal/builders/homebrew_test.go`: Added 25+ tests for validation and generation
- `internal/actions/text_replace.go`: New action for literal/regex text replacement
- `internal/actions/text_replace_test.go`: 14 tests covering all replacement scenarios
- `internal/actions/apply_patch.go`: New action for applying patches
- `internal/actions/apply_patch_test.go`: 10 tests covering patch application
- `internal/actions/dependencies.go`: Added entries for new actions

## Key Decisions

- **Resources and patches stored in recipe fields**: Instead of generating steps, resources/patches are stored in dedicated recipe sections (`[[resources]]`, `[[patches]]`) for clarity
- **Inreplace as text_replace steps**: Inreplace operations are emitted as recipe steps since they're executed during the build process
- **Security validation**: All paths validated to prevent traversal, all URLs must be HTTPS
- **Inline vs URL patches**: Both supported with mutual exclusion enforced

## Trade-offs Accepted

- **patch command dependency**: apply_patch relies on system patch utility; if unavailable, provides clear error message
- **Regex support optional**: text_replace defaults to literal matching; regex opt-in prevents accidental metacharacter issues

## Test Coverage

- New tests added: 50+
- Recipe types: 6 new tests for Resource/Patch serialization
- Homebrew builder: 25+ new tests for validation and generation
- Actions: 24 new tests for text_replace and apply_patch

## Known Limitations

- No resource staging action yet (resources are declared in recipe but not downloaded/staged during execution)
- patch command must be installed on the system for apply_patch to work
- Inline patches with special TOML characters may need escaping

## Future Improvements

- Add resource_stage action for downloading and extracting resource archives
- Add checksum verification for patches and resources
- Support for git-format patches
