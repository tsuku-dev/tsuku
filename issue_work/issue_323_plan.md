# Issue 323 Implementation Plan

## Goal
Define the `Provider` interface and associated types for LLM abstraction in `internal/llm/provider.go`.

## Design Reference
- Design doc: `docs/DESIGN-llm-slice-3-repair-loop.md`
- Section: "1. Provider Interface (`internal/llm/provider.go`)"

## Implementation Steps

### Step 1: Create `internal/llm/provider.go`

Create new file with the following types:

1. **`Provider` interface**
   - `Name() string` - returns provider identifier
   - `Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)` - single-turn completion

2. **`CompletionRequest` struct**
   - `SystemPrompt string`
   - `Messages []Message`
   - `Tools []ToolDef`
   - `MaxTokens int`

3. **`CompletionResponse` struct**
   - `Content string`
   - `ToolCalls []ToolCall`
   - `StopReason string`
   - `Usage Usage` (reuse existing type from `cost.go`)

4. **`Message` struct**
   - `Role Role`
   - `Content string`
   - `ToolCalls []ToolCall` (for assistant messages)
   - `ToolResult *ToolResult` (for user messages with tool results)

5. **`Role` type**
   - `RoleUser Role = "user"`
   - `RoleAssistant Role = "assistant"`

6. **`ToolCall` struct**
   - `ID string`
   - `Name string`
   - `Arguments map[string]any`

7. **`ToolResult` struct**
   - `CallID string`
   - `Content string`
   - `IsError bool`

8. **`ToolDef` struct**
   - `Name string`
   - `Description string`
   - `Parameters map[string]any` (JSON Schema)

### Step 2: Add Documentation Comments

All exported types must have documentation comments explaining their purpose, as specified in the acceptance criteria.

### Step 3: Verify Compatibility

Ensure the `Usage` type from `cost.go` is compatible with the design (it already has `InputTokens` and `OutputTokens`).

## Files Changed
- `internal/llm/provider.go` (new)

## Testing
No unit tests needed for this issue - it only defines types and interfaces. Tests will be added in subsequent issues that implement the interface.

## Acceptance Criteria Checklist
- [ ] `Provider` interface defined with `Name()` and `Complete()` methods
- [ ] `CompletionRequest` and `CompletionResponse` types defined
- [ ] `Message`, `ToolCall`, `ToolResult`, `ToolDef` types defined
- [ ] `Usage` type includes input/output token counts (already exists)
- [ ] All types have documentation comments
