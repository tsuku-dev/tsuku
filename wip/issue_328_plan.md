# Issue 328 Implementation Plan

## Summary

Implement `GeminiProvider` that uses the Google GenAI Go SDK to call Gemini 2.0 Flash with function calling support, implementing the existing `Provider` interface defined in `provider.go`.

## Approach

Create a single-turn provider implementation that converts between our common types (`CompletionRequest`, `ToolDef`, `Message`) and the Gemini SDK's native types (`genai.Content`, `genai.FunctionDeclaration`). The provider handles only a single API call - conversation loops remain in the builder layer.

### Alternatives Considered
- **Direct API calls without SDK**: More control but requires maintaining HTTP client, auth, and JSON serialization. SDK is maintained by Google and handles these concerns.
- **Using vertexai package instead of genai**: Would require GCP project setup. The `genai` package works with simple API keys.

## Files to Create
- `internal/llm/gemini.go` - GeminiProvider implementation

## Files to Modify
- `go.mod` - Add google/generative-ai-go dependency

## Implementation Steps
- [ ] Add google/generative-ai-go dependency
- [ ] Create GeminiProvider struct with client and model name
- [ ] Implement NewGeminiProvider constructor (reads GOOGLE_API_KEY)
- [ ] Implement Name() method (returns "gemini")
- [ ] Implement convertTools() - ToolDef to genai.FunctionDeclaration
- [ ] Implement convertMessages() - Message to genai.Content
- [ ] Implement convertResponse() - genai.GenerateContentResponse to CompletionResponse
- [ ] Implement Complete() method orchestrating the above
- [ ] Write unit tests for type conversions
- [ ] Write integration test with skip if no API key

## Testing Strategy
- Unit tests: Test type conversion functions with mock data
- Integration tests: Test actual API calls with `GOOGLE_API_KEY` (skipped in CI)

## Risks and Mitigations
- **Different function calling syntax**: Gemini uses `functionCall`/`functionResponse` vs Claude's `tool_use`/`tool_result`. Mitigation: Provider abstraction handles the mapping.
- **Token counting differences**: Gemini may report tokens differently. Mitigation: Map to common Usage struct as best as possible.

## Success Criteria
- [ ] GeminiProvider implements Provider interface
- [ ] Unit tests pass for all type conversions
- [ ] Integration test with actual API call works (manual verification)
- [ ] `go test ./internal/llm/...` passes
