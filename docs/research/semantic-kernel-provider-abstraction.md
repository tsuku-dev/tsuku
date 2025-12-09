# Semantic Kernel Provider Abstraction Research

## Architecture Pattern: Two-Layer Abstraction

**Layer 1: Microsoft.Extensions.AI** (Foundation)
- Core interface: `IChatClient` with 2 methods (`CompleteAsync`, `CompleteStreamingAsync`)
- Middleware pattern via `DelegatingChatClient`

**Layer 2: Semantic Kernel** (Orchestration)
- Agent patterns, plugins, prompt templates
- Consumes `IChatClient` for LLM interactions

## Where Agent Loop Lives

**Agent abstractions at SK layer**, not foundation:
- `ChatCompletionAgent` - Stateless (app manages history)
- `OpenAIAssistantAgent` - Stateful (service manages history)
- `AgentThread` - Manages conversation state

## Function Calling Abstraction

6-step unified process regardless of provider:
1. Serialize functions to JSON schema
2. Send messages + functions to model
3. Model processes input
4. Handle response
5. Invoke function
6. Return result to model

**Switching providers requires only initialization change:**
```csharp
// Azure OpenAI
kernelBuilder.AddAzureOpenAIChatCompletion(deployment, apiKey, endpoint);
// OpenAI
kernelBuilder.AddOpenAIChatCompletion(modelId, apiKey);
// Everything else stays the same
```

## Recommendation for Tsuku

**Adopt:**
- Single interface abstraction (`LLMClient`)
- Provider-specific implementations
- Factory pattern for provider selection
- Shared usage tracking

**Avoid:**
- Two-layer architecture (overkill for 2 providers)
- Middleware pipeline (unnecessary complexity)
- Agent orchestration (not needed for recipe generation)

## Key Takeaway

> "Abstract behind a domain-specific interface that returns exactly what you need."

Keep it simple: one interface, two implementations, shared types.
