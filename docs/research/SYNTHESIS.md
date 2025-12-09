# Multi-Provider LLM Architecture Research Synthesis

## Research Scope

Analyzed 9 tools/frameworks:
- **LLM Frameworks**: LangChain, Semantic Kernel, LlamaIndex
- **Go Ecosystem**: langchaingo, go-openai patterns
- **SDK Comparison**: Anthropic vs Google GenAI SDKs
- **Coding Agents**: Aider, Open Interpreter, goose, Continue

## Consensus Findings

### 1. Conversation Loop Location

**Strong consensus: Loop lives in orchestration layer, NOT in provider implementations.**

| Tool | Loop Location |
|------|---------------|
| LangChain | `AgentExecutor` (application code) |
| LlamaIndex | `AgentRunner` (application code) |
| Aider | `base_coder.py` (shared client layer) |
| Open Interpreter | `core/core.py` (centralized) |
| goose | `crates/goose/src/agents/` (orchestration) |
| Continue | `core/llm/streamChat.ts` (core layer) |
| Go libraries | Application/builder code |

**Pattern**: Provider implementations handle single request/response. The multi-turn loop lives in business logic.

### 2. Provider Interface Design

**Consensus: Thin interface, typically 1-3 methods.**

```go
// Typical minimal interface
type Provider interface {
    Generate(ctx, messages, tools) (*Response, error)
    // OR streaming variant
    StreamChat(ctx, messages, tools) (<-chan Chunk, error)
}
```

| Tool | Interface Size |
|------|----------------|
| Aider | 1 method (via LiteLLM `completion()`) |
| goose | 2 methods (`complete`, `stream`) |
| Open Interpreter | 1 method (via LiteLLM) |
| Go libraries | 1-2 methods typical |

**Anti-pattern**: Thick interfaces (Continue's ILLM has 12+ methods, adds complexity).

### 3. Tool/Function Calling Abstraction

**Two dominant patterns:**

**Pattern A: Unified Internal Format (LangChain, Continue)**
- Define tools once in internal format
- Convert at call boundary to provider-specific format
- More abstraction, better for many providers

**Pattern B: Provider-Native Schemas (Aider, goose)**
- Use provider's native format directly
- Less abstraction, simpler for few providers
- Aider notably avoids function calling entirely (prompt engineering)

**Recommendation for 2 providers**: Pattern B (provider-native) is simpler.

### 4. Anthropic vs Google SDK Differences

**Critical finding: Fundamental architectural differences.**

| Aspect | Anthropic | Google GenAI |
|--------|-----------|--------------|
| State | Stateless | Stateful sessions |
| History | Manual (send full each turn) | Automatic (managed by Chat) |
| Tool ID | Unique ID for correlation | Name-based correlation |

**Implication**: A thick shared abstraction leaks. Thin interface with provider-specific implementations handles this better.

### 5. When Thick Abstraction Works

**Use thick abstraction (shared loop in abstraction layer) when:**
- Supporting 10+ providers (LangChain, Continue)
- Feature parity is achievable
- Team has resources for abstraction maintenance

**Use thin abstraction (loop in builder) when:**
- Supporting 2-3 providers (tsuku's case)
- Providers have fundamental differences
- Simplicity is valued

## Recommendation for Tsuku

Based on research consensus, the optimal architecture is:

### Hybrid Approach: Thin Interface + Shared Loop in Builder

```
internal/
  llm/
    provider.go       # Thin interface (1-2 methods)
    anthropic/
      client.go       # Anthropic implementation
    gemini/
      client.go       # Gemini implementation
  builders/
    github_release.go # Multi-turn loop lives HERE
```

**Interface:**
```go
type Provider interface {
    Generate(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
    Messages []Message
    Tools    []Tool       // Provider-native format
    MaxTurns int
}

type Response struct {
    Content   string
    ToolCalls []ToolCall  // Provider-native format
    Usage     *Usage
    StopReason string
}
```

**Loop in Builder:**
```go
func (b *GitHubReleaseBuilder) generateRecipe(ctx context.Context) (*Recipe, error) {
    messages := b.buildInitialMessages()

    for turn := 0; turn < b.maxTurns; turn++ {
        resp, err := b.provider.Generate(ctx, &Request{
            Messages: messages,
            Tools:    b.tools,
        })
        if err != nil {
            return nil, err
        }

        if resp.StopReason == "end_turn" {
            return b.parseRecipe(resp.Content)
        }

        // Handle tool calls
        messages = append(messages, b.toAssistantMessage(resp))
        results := b.executeTools(resp.ToolCalls)
        messages = append(messages, b.toToolResultMessage(results))
    }
    return nil, errors.New("max turns exceeded")
}
```

### Why This Works

1. **Provider implementations are simple** - just wrap SDK, no loop logic
2. **Loop is business logic** - can customize per use case
3. **Easy to test** - mock `Generate()` method
4. **Clear boundaries** - provider knows nothing about multi-turn
5. **Handles SDK differences** - each provider uses native patterns

### What This Means for Original Options

The research supports a **hybrid** closest to Option 1A (Thin Interface):
- Provider implementations are thin (1A)
- Loop is NOT in provider (1A)
- But loop IS shared in builder code (not duplicated per provider)

**Revised recommendation**: Option 1A with loop in builder (not provider).

## Key Takeaways

1. **"Put loop in GitHubReleaseBuilder, NOT in internal/llm/client.go"** (Go libraries research)

2. **"Thin provider interface, loop in business logic, provider-native tool schemas"** (Go libraries research)

3. **"Go's interface system allows cleaner design with explicit provider implementations rather than a magic translation layer"** (Open Interpreter research)

4. **"Start simple - interface + 2 implementations. Provider logic isolated. Avoid over-abstraction for 2 providers."** (goose research)

5. **"Abstract what's the same, parameterize what's different, encapsulate what varies."** (SDK comparison)
