# Issue 375 Summary

## What Was Implemented

Hybrid recipe preview flow for LLM-generated recipes. When creating a recipe from GitHub releases, users now see a summary of downloads, actions, verification command, and cost before deciding to install, view full recipe, or cancel.

## Changes Made

- `internal/builders/builder.go`: Added `Cost` field to `BuildResult` struct
- `internal/builders/github_release.go`: Populate `Cost` field from LLM usage
- `cmd/tsuku/create.go`: Added preview flow with helper functions:
  - `previewRecipe()` - displays summary and handles prompt loop
  - `promptForApproval()` - handles v/i/c user input
  - `extractDownloadURLs()` - extracts download URLs from recipe steps
  - `describeStep()` - returns human-readable step descriptions
  - `formatRecipeTOML()` - formats recipe as TOML for display
- `cmd/tsuku/create_test.go`: Added unit tests for helper functions

## Key Decisions

- **Preview only for GitHub builder**: Ecosystem builders (cargo, npm, etc.) don't use LLM and write directly
- **v/i/c prompt**: Simple single-character options matching the design spec
- **Cost display precision**: Using $%.4f format for small cost values

## Trade-offs Accepted

- **No TTY check**: Preview prompts in non-TTY environments will fail; future --yes flag (issue #374) will address this
- **Simple containsString helper**: Implemented inline in tests rather than using a library

## Test Coverage

- New tests added: 3 test functions (TestDescribeStep, TestExtractDownloadURLs, TestFormatRecipeTOML)
- Covers core helper functions; preview/prompt functions require manual testing

## Known Limitations

- Preview is mandatory for GitHub builder (no bypass until --yes flag is implemented)
- Interactive prompt doesn't work in non-TTY environments

## Future Improvements

- Issue #374: Add --yes flag to bypass preview
- Issue #377: Add progress indicators during generation
