# Issue 279 Implementation Plan

## Summary

Refactor the `fetch_file` tool handler to take a relative file path instead of a full URL, constructing GitHub raw content URLs from repo/tag context stored during recipe generation.

## Approach

The design doc specifies that `fetch_file` should accept a path relative to the repo root (e.g., `INSTALL.md`, `docs/usage.md`) and construct the full URL internally. This requires:
1. Storing repo and tag context in the client or passing it through the generation flow
2. Changing the tool schema from `url` to `path`
3. Adding validation for text content types
4. Providing helpful error messages for 404s

### Alternatives Considered
- **Keep URL-based approach**: Simpler but exposes implementation details to the LLM and allows fetching arbitrary URLs, which is less secure and less aligned with the design doc.

## Files to Modify
- `internal/llm/tools.go` - Change `FetchFileInput` to use `path` instead of `url`, update tool schema
- `internal/llm/client.go` - Refactor `fetchFile` to accept path and construct URL from repo/tag context, add content-type validation

## Files to Create
- None

## Implementation Steps
- [ ] 1. Add context struct to store repo/tag during generation, modify GenerateRecipe to track active release tag
- [ ] 2. Update FetchFileInput schema from `url` to `path`
- [ ] 3. Update tool schema description to match design doc
- [ ] 4. Refactor fetchFile method signature and implementation
- [ ] 5. Add content-type validation (text/* only, reject binaries)
- [ ] 6. Update executeToolUse to pass context to fetchFile
- [ ] 7. Update existing tests for the new signature
- [ ] 8. Add unit test for fetch_file with mock HTTP server
- [ ] 9. Add test for 404 handling with helpful error message

## Testing Strategy
- Unit tests: Mock HTTP server returning file content, 404 errors, and binary content types
- Integration test: Existing integration test should still work as it tests the overall flow

## Risks and Mitigations
- **Breaking change to tool schema**: The LLM may already be generating prompts expecting `url` parameter. Mitigation: This is early in development (Slice 1), and changing the schema now is the right time.
- **Hardcoded tag assumption**: Need to track which release we're analyzing. Mitigation: Use the first release tag from the request as the default for fetching files.

## Success Criteria
- [ ] Tool schema uses `path` parameter matching design doc
- [ ] fetchFile constructs raw.githubusercontent.com URLs correctly
- [ ] 404 errors return helpful message to LLM
- [ ] Binary content types are rejected
- [ ] All unit tests pass
- [ ] 60 second timeout is configured

## Open Questions
None - the design doc is clear on the requirements.
