# Issue 330 Implementation Plan

## Summary

Add repair loop to GitHubReleaseBuilder that validates generated recipes in containers and retries with sanitized error feedback on validation failures.

## Dependencies (All Merged)

- #323: Provider interface and types
- #324: Circuit breaker
- #325: Error sanitizer
- #326: Error parser
- #327: Claude provider
- #329: Provider factory with failover

## Architecture Overview

Per the design doc (DESIGN-llm-slice-3-repair-loop.md), the chosen approach is:
- **1A**: Thin interface with conversation loop in builder
- **2A**: Per-provider circuit breakers
- **3B**: Continue conversation with error feedback

The GitHubReleaseBuilder will:
1. Use factory to get provider with failover
2. Run conversation loop directly (not through `llm.Client`)
3. Validate generated recipe in container
4. On failure: sanitize error, parse category, continue conversation
5. Max 2 repair attempts before failing

## Current State Analysis

### `internal/builders/github_release.go`
- Uses `llm.Client` directly (line 31)
- `Build()` calls `b.llmClient.GenerateRecipe()` and returns result
- No validation or repair loop

### `internal/llm/client.go`
- Multi-turn conversation loop in `GenerateRecipe()`
- Manages messages, tool calls, and tool results
- Uses `ClaudeProvider.Complete()` for each turn

### `internal/llm/factory.go`
- `NewFactory()` auto-detects providers from env vars
- `GetProvider()` returns available provider respecting breaker state
- `ReportSuccess()`/`ReportFailure()` update breaker state

### `internal/validate/executor.go`
- `Validate()` runs recipe in container
- Returns `ValidationResult{Passed, Skipped, ExitCode, Stdout, Stderr, Error}`

### `internal/validate/sanitize.go`
- `Sanitizer.Sanitize()` removes home paths, IPs, credentials
- Max length 2000 chars

### `internal/validate/errors.go`
- `ParseValidationError()` returns `ParsedError{Category, Message, Details, Suggestions}`
- Categories: binary_not_found, extraction_failed, verify_failed, etc.

## Implementation Steps

### Step 1: Add validation and factory dependencies to builder

Modify `GitHubReleaseBuilder` struct to include:
- `factory *llm.Factory` instead of `llmClient *llm.Client`
- `executor *validate.Executor` for container validation
- `sanitizer *validate.Sanitizer` for error sanitization

Update constructors to accept factory and executor.

### Step 2: Move conversation loop from Client to Builder

Extract the conversation loop logic from `client.go:GenerateRecipe()` into the builder:
- Build system prompt and user message (reuse from client.go)
- Manage messages slice
- Call `provider.Complete()` in loop
- Execute tool calls
- Return pattern when extract_pattern called

This gives the builder control over the conversation for repair.

### Step 3: Add validation after recipe generation

After `extract_pattern` is called and recipe generated:
1. Build asset URL for validation
2. Call `executor.Validate(ctx, recipe, assetURL)`
3. If `result.Passed` or `result.Skipped`, return success
4. If failed, proceed to repair loop

### Step 4: Implement repair loop

On validation failure:
1. Sanitize error: `sanitizer.Sanitize(result.Stderr + result.Stdout)`
2. Parse error: `validate.ParseValidationError(result.Stdout, result.Stderr, result.ExitCode)`
3. Build repair message with error feedback and suggestions
4. Append to conversation as user message
5. Continue conversation loop until new extract_pattern
6. Validate again
7. Repeat up to 2 repair attempts

### Step 5: Add provider failover on API errors

When `provider.Complete()` fails:
1. `factory.ReportFailure(provider.Name())`
2. Get new provider: `factory.GetProvider(ctx)`
3. If different provider, start fresh conversation with error context
4. On success: `factory.ReportSuccess(provider.Name())`

### Step 6: Update BuildResult with repair metadata

Add to `BuildResult`:
- `RepairAttempts int` - number of repair attempts made
- `Provider string` - which provider generated the recipe

### Step 7: Add unit tests

Test cases:
- Validation passes on first attempt (no repair)
- Validation fails, repair succeeds on attempt 1
- Validation fails, repair succeeds on attempt 2
- Validation fails after max attempts (error)
- Validation skipped (no container runtime)
- Provider failover during generation
- Provider failover during repair

Use mock provider and mock executor for unit tests.

## File Changes

### Modified Files

1. **`internal/builders/github_release.go`**
   - Change `llmClient *llm.Client` to `factory *llm.Factory`
   - Add `executor *validate.Executor`
   - Add `sanitizer *validate.Sanitizer`
   - Move conversation loop into builder
   - Add validation and repair loop
   - Update constructors

2. **`internal/builders/types.go`** (if exists)
   - Add `RepairAttempts` and `Provider` to `BuildResult`

### New Files

1. **`internal/builders/github_release_test.go`** (update existing)
   - Add test cases for repair loop
   - Add mock executor

## Constants

```go
const (
    // MaxRepairAttempts is the maximum number of times to retry after validation failure
    MaxRepairAttempts = 2
)
```

## Repair Prompt Template

```
The recipe you generated failed validation. Here is the error:

---
{sanitized_error}
---

Error analysis:
- Category: {error_category}
- Details: {parsed_details}

Please analyze what went wrong and call extract_pattern again with a corrected recipe.

Common fixes for {error_category}:
{suggestions}
```

## Testing Strategy

1. Unit tests with mock provider and executor
2. Manual integration test with real provider (optional, not in CI)
3. Rely on existing integration tests (#332) for end-to-end coverage

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Long conversations hit context limits | Max 2 repair attempts bounds size |
| Repair loop adds cost | Limited retries, continue conversation cheaper than restart |
| Provider switch loses context | Include original error context in new conversation |
| Container validation slow | Already validated in Slice 2, acceptable for repair |

## Exit Criteria

- [ ] Builder uses factory instead of direct client
- [ ] Validation runs after recipe generation
- [ ] Repair loop continues conversation with error feedback
- [ ] Max 2 repair attempts enforced
- [ ] Error sanitization applied before LLM calls
- [ ] Provider failover works during generation and repair
- [ ] Unit tests cover all repair scenarios
- [ ] All existing tests pass
