# Open Interpreter Multi-Provider Architecture Research

## Architecture Pattern: Delegation via LiteLLM

Like Aider, Open Interpreter delegates to **LiteLLM** for provider abstraction.

**Configuration Interface:**
```python
interpreter.llm.model = "claude-3-sonnet"
interpreter.llm.api_key = "..."
interpreter.llm.context_window = 200000
interpreter.llm.supports_functions = True  # Feature flag
interpreter.llm.supports_vision = True     # Feature flag
```

## Where Agent Loop Lives

**Centralized in `interpreter/core/core.py`** - completely decoupled from providers.

**Three-Role Message System:**
- `user` - User input
- `assistant` - LLM responses
- `computer` - Code execution results

**Loop API:**
```python
interpreter.chat(message)           # Single message
interpreter.chat()                  # Interactive loop
interpreter.chat(message, stream=True)  # Streaming
interpreter.messages                # History access
```

## Tool/Function Calling Abstraction

**Two Strategies:**

1. **Native Function Calling** (when `supports_functions=True`)
   - LiteLLM translates tools to provider format
   - Normalizes responses back to OpenAI format

2. **Prompt-Based Fallback** (when `supports_functions=False`)
   - Instructions in system prompt for code execution
   - Parses structured text output for code blocks
   - Enables models without native tool use

## Pros and Cons

**Pros:**
- Clean separation of concerns
- Rapid provider addition
- Unified developer experience
- Local model support (Ollama, etc.)
- Generator-based streaming

**Cons:**
- Python-only (LiteLLM is Python)
- Abstraction leaks (function calling bugs)
- External dependency risk
- Feature lag behind providers

## Key Lessons for Go CLI

1. **Build thin custom abstraction** (no LiteLLM-for-Go exists)
   ```go
   type LLMProvider interface {
       Chat(ctx context.Context, messages []Message) (*Response, error)
       StreamChat(ctx context.Context, messages []Message) (<-chan StreamChunk, error)
       SupportsTools() bool
   }
   ```

2. **Use channels for streaming** (Go equivalent of generators)

3. **Feature flags for capabilities**
   ```go
   type ProviderCapabilities struct {
       SupportsTools     bool
       SupportsVision    bool
       SupportsStreaming bool
       MaxContextTokens  int
   }
   ```

4. **Provider attribution in errors**
   ```go
   type LLMError struct {
       Provider string
       Original error
   }
   ```

5. **Progressive enhancement**
   - Phase 1: Single provider
   - Phase 2: Configuration
   - Phase 3: Second provider
   - Phase 4: Profiles

## Key Takeaway

> "Go's interface system allows cleaner, more maintainable design with explicit provider implementations rather than a magic translation layer."
