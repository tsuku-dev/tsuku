# Goose (Block) Multi-Provider Architecture Research

## Architecture Pattern: Trait-Based Provider Interface (Rust)

Goose is written in Rust (59%) with TypeScript UI (33%). Uses Rust traits for provider abstraction.

**Key Components:**
1. **Provider Trait** (`crates/goose/src/providers/base.rs`)
   - `complete()` - Non-streaming
   - `stream()` - Streaming
   - Metadata access (model, limits, pricing)

2. **ProviderRegistry** - Thread-safe singleton for provider discovery
   - Lazy initialization: `OnceCell<RwLock<ProviderRegistry>>`
   - Classifies: Preferred, Builtin, Declarative, Custom

3. **Provider Factory** - Configuration-driven instantiation

**20+ Providers:** OpenAI, Anthropic, Azure, Bedrock, Gemini, Groq, Ollama, etc.

## Where Agent Loop Lives

**In `crates/goose/src/agents/`** - Orchestration flow:

```
User Input → Agent → LLM (with tools) → Tool Call Response
     ↑                                         ↓
     └──── Tool Execution Results ← Executor ──┘
```

**Key Characteristics:**
- Autonomous execution (runs without intervention)
- Error resilience (errors sent back to LLM)
- Context revision (removes old info for token limits)
- Session persistence (JSONL with backup/recovery)

## Tool Calling Abstraction

Uses **Model Context Protocol (MCP)** - Anthropic's open standard.

**Flow:**
```
MCP Tool Definition → Provider-Specific Format
                     ↓
              (OpenAI functions)
                     OR
           (Anthropic tool_use blocks)
                     ↓
        Unified Tool Call Execution
```

Provider-specific handling in each provider module.

## Design Patterns

1. **Trait-Based Polymorphism** - Clean provider abstraction
2. **Factory Pattern** - Central registry, config-driven
3. **Strategy Pattern** - Provider-specific handling
4. **Adapter Pattern** - MCP → provider format translation

## Pros and Cons

**Pros:**
- Strong type safety (Rust)
- Performance (compiled binary)
- 20+ providers out-of-box
- Real production use (~5,000 weekly users at Block)
- Model-agnostic tools via MCP

**Cons:**
- Rust complexity
- Provider-specific quirks still surface
- Configuration complexity
- Tool calling reliability varies by provider

## Key Lessons for Go CLI

**What to Adopt:**
1. Interface-based abstraction
   ```go
   type Provider interface {
       Complete(ctx, messages, tools) (Response, error)
       Stream(ctx, messages, tools) (<-chan StreamChunk, error)
   }
   ```

2. Agent loop pattern (request → LLM → tools → repeat)
3. Session persistence (JSONL)
4. Error resilience (send errors back to LLM)
5. Usage tracking (token counts)

**What to Simplify (for 2 providers):**
1. No registry needed - direct instantiation
2. Simpler error handling
3. Less abstraction
4. Pick canonical tool format, convert other provider

**Recommended Structure:**
```
/internal/
  /provider/
    provider.go    # Interface
    openai.go
    anthropic.go
    message.go     # Shared types
  /agent/
    agent.go       # Loop orchestration
    session.go     # State management
  /tool/
    tool.go        # Tool interface
    executor.go    # Execution
```

## Key Takeaway

> "Start simple - interface + 2 implementations. Provider logic isolated. Internal message format with conversion methods. Avoid over-abstraction for 2 providers."
