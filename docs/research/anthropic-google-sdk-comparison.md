# Anthropic vs Google AI SDK Comparison

## Executive Summary

Both SDKs support tool use but with **significantly different architectural patterns**:
- **Anthropic**: Stateless, message-history-based
- **Google GenAI**: Stateful chat session

**Recommendation:** Thin abstraction (provider-specific loops) fits better than thick shared loop.

## Side-by-Side Comparison

### Conversation State Management

| Aspect | Anthropic SDK | Google GenAI SDK |
|--------|---------------|------------------|
| **State** | Stateless | Stateful |
| **History** | Manual (append to `messages`) | Automatic (managed by `Chat`) |
| **API Pattern** | Send full history each turn | Send only new parts |

**Anthropic:**
```go
messages := []anthropic.MessageParam{...}
resp1, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
    Messages: messages,  // Full history
})
messages = append(messages, resp1.ToParam())
messages = append(messages, anthropic.NewUserMessage(toolResults...))
resp2, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
    Messages: messages,  // Full history again
})
```

**Google:**
```go
chat, _ := client.Chats.Create(ctx, "gemini-2.5-flash", config, nil)
resp1, _ := chat.SendMessage(ctx, genai.Part{Text: "query"})
resp2, _ := chat.SendMessage(ctx, functionResponse)  // Only new content
```

### Tool Definition Structure

| Aspect | Anthropic SDK | Google GenAI SDK |
|--------|---------------|------------------|
| **Schema** | JSON Schema in Go structs | `Schema` type with properties |
| **Tool Type** | `ToolParam` | `FunctionDeclaration` in `Tool` |

### Tool Correlation

| Aspect | Anthropic SDK | Google GenAI SDK |
|--------|---------------|------------------|
| **Call ID** | Unique `ID` on `ToolUseBlock` | Identified by `Name` |
| **Response** | `ToolResultBlock` with `ToolUseID` | `FunctionResponse` with `Name` |
| **Parallel** | Explicit via ID matching | Relies on order/name |

## Key Differences Requiring Abstraction

1. **Message Structure Incompatibility**
   - Anthropic: Union type content blocks, role alternation
   - Google: Parts array, `user`/`model` roles

2. **Tool Correlation Mechanism**
   - Anthropic: Tool use ID for unambiguous matching
   - Google: Function name for matching

3. **Streaming Differences**
   - Anthropic: Event-based streaming
   - Google: Go 1.23+ iterators

## Recommendation: Thin Abstraction

Given substantial differences, **provider-specific loop implementations**:

```go
type Client interface {
    GenerateRecipe(ctx context.Context, req *GenerateRequest) (*AssetPattern, *Usage, error)
}
```

**Why Thin Wins:**

1. **Natural Control Flow** - Each provider has its own optimal loop pattern
2. **Provider-Specific Optimizations** - Batch tool results (Anthropic), auto context (Google)
3. **Easier Testing** - Test each provider independently
4. **Evolution Path** - Add providers without affecting existing ones

## Implementation Pattern

```go
// anthropic.go
func (c *AnthropicClient) GenerateRecipe(ctx, req) (*AssetPattern, *Usage, error) {
    messages := c.buildInitialMessages(req)
    for turn := 0; turn < c.maxTurns; turn++ {
        resp, _ := c.client.Messages.New(ctx, anthropic.MessageNewParams{
            Messages: messages,
            Tools:    tools,
        })
        messages = append(messages, resp.ToParam())
        // ... handle tool calls
    }
}

// google.go
func (c *GoogleClient) GenerateRecipe(ctx, req) (*AssetPattern, *Usage, error) {
    chat, _ := c.client.Chats.Create(ctx, c.model, config, nil)
    resp, _ := chat.Send(ctx, initialPart)
    for turn := 0; turn < c.maxTurns; turn++ {
        // ... handle function calls with chat.Send()
    }
}
```

## Key Takeaway

> "Abstract what's the same, parameterize what's different, encapsulate what varies."

The shared interface (`Client`) provides consistency at the **business logic level**, while implementation details remain provider-specific.
