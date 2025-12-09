# Issue 282 Summary

## What Was Implemented

Extended the `--from` flag in `tsuku create` to support the `github:owner/repo` format, enabling LLM-based recipe generation from GitHub release assets.

## Changes Made

- `cmd/tsuku/create.go`:
  - Added `parseFromFlag()` function to parse `builder:sourceArg` format
  - Updated help text with GitHub builder syntax and examples
  - Modified `runCreate()` to:
    - Register GitHubReleaseBuilder when github: prefix detected
    - Skip toolchain check for github builder (not needed)
    - Skip CanBuild check for github builder (uses SourceArg instead)
    - Pass SourceArg to BuildRequest
    - Print warnings to stderr (for LLM cost info)

- `cmd/tsuku/create_test.go` (new):
  - Added unit tests for `parseFromFlag()` (9 test cases)
  - Added unit tests for `normalizeEcosystem()` (20 test cases)

## Key Decisions

- **Backward compatibility**: Existing `--from crates.io` syntax continues to work unchanged
- **Case-insensitive prefix**: `github:`, `GitHub:`, `GITHUB:` all work, but sourceArg case is preserved
- **Lazy registration**: GitHubReleaseBuilder only instantiated when needed (avoids API key check for other builders)

## Trade-offs Accepted

- **No integration tests**: Full integration testing deferred to #283 (ground truth validation) since the GitHub Release Builder has its own comprehensive test suite

## Test Coverage

- New tests added: 2 test functions with 29 test cases total
- All existing tests continue to pass
- Unit tests cover parsing logic edge cases

## Known Limitations

- Requires `ANTHROPIC_API_KEY` environment variable for github builder
- No budget confirmation prompts (per design doc, deferred to later slice)

## Future Improvements

- Add bash completion for github:owner/repo suggestions
- Support other builders with sourceArg format if needed
