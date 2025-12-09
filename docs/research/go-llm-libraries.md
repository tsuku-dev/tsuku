# Go LLM Libraries Provider Abstraction Research

## Common Patterns

### A. `llms.Model` Interface (LangChainGo)

```go
type Model interface {
    GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}
```
- Provider-agnostic
- Functional options for configuration
- Multi-modal support

### B. Provider-Specific Clients (go-openai)

```go
client := openai.NewClient("api-key")
resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
    Model: openai.GPT4,
    Messages: messages,
    Tools: tools,  // Provider-native schema
})
```
- Direct API mapping, zero overhead
- No provider portability

### C. Hybrid (Agency, go-llm)

Thin interface focused on conversation management, provider-specific LLM calls.

## Ecosystem Trend: Favor Thin Abstractions

> "Many enterprise teams end up sunsetting LLM native libraries like LangChain in favor of communicating with LLMs directly."

**Why Go Prefers Thin:**
- Performance (goroutine-per-request vs Python async)
- Type safety (compile-time validation)
- Philosophy (explicitness over magic)

## Where Conversation Loop Lives

**Key Finding:** Loops live in **application code**, NOT provider clients.

```go
// Application code or thin agent wrapper
for {
    resp := callLLM(ctx, messages)
    if resp.ToolCalls != nil {
        toolResults := executeTools(resp.ToolCalls)
        messages = append(messages, toolResults...)
        continue
    }
    if resp.StopReason == "end_turn" {
        return resp.Content
    }
}
```

## Recommendation for Tsuku

**Architecture:**
```
internal/llm/
├── provider.go       # Thin interface (GenerateWithTools)
├── types.go          # Shared types
├── anthropic/
│   ├── client.go     # Anthropic SDK wrapper
│   └── tools.go      # Anthropic tool schemas
└── gemini/
    ├── client.go     # Gemini SDK wrapper
    └── tools.go      # Gemini tool schemas

internal/builders/
└── github_release.go # Multi-turn loop lives HERE
```

**Interface:**
```go
type Provider interface {
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}
```

**Why This Works:**
- LLM client is thin (just wraps SDK)
- Loop is business logic (in builder)
- Easy to test (mock Generate())
- Clear boundaries

## Key Takeaway

> "Put loop in GitHubReleaseBuilder, NOT in internal/llm/client.go"

Thin provider interface, loop in business logic, provider-native tool schemas.
