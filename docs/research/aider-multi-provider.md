# Aider Multi-Provider Architecture Research

## Architecture Pattern: Thin Wrapper over LiteLLM

Aider delegates all provider abstraction to **LiteLLM** (Python library supporting 100+ providers).

**Key Components:**
- `aider/models.py` - Model metadata, imports LiteLLM
- `aider/sendchat.py` - Wraps `litellm.completion()` calls
- `aider/coders/base_coder.py` - Core agentic loop

## Where Agentic Loop Lives

**Entirely in shared client layer** (`base_coder.py`), not provider-specific.

**Loop Components:**
1. `run()` - Continuous prompt loop
2. `run_one(user_message)` - Single iteration
3. **Reflection mechanism** - Errors trigger retry (up to `max_reflections`)
4. **Auto-lint/test** - Failures auto-feed back to LLM

## Tool/Function Calling Approach

**Aider does NOT use provider function calling APIs** for code editing.

Instead uses **prompt engineering with structured output formats**:
- `diff` (search/replace blocks) - GPT-4 Turbo
- `udiff` (unified diffs) - 3x improvement for GPT-4 Turbo
- `whole` - Entire file replacement
- `diff-fenced` - Default for Gemini

**Architect/Editor Split:**
- Architect model: Solves coding problem (e.g., o1)
- Editor model: Formats edits properly (e.g., GPT-4o, Sonnet)

## Pros and Cons

**Pros:**
- Minimal maintenance (LiteLLM handles API changes)
- Fast provider switching (`--model` flag)
- 100+ providers supported
- Unified cost tracking

**Cons:**
- ~3-4ms latency per call
- Feature lag (new features require LiteLLM update)
- Limited granular control
- Dependency risk (version conflicts)

## Key Lessons for Go CLI

1. **Thin wrapper pattern** - Sufficient for 2 providers
2. **Keep agentic loop provider-agnostic** - Loop in shared code
3. **Model-specific configuration, not provider-specific code**
4. **Prompt engineering > function calling** (more reliable across providers)
5. **Unified message format internally** - Transform at call boundary

**Recommended Structure:**
```
internal/
  llm/
    provider.go      # interface LLMProvider
    openai.go        # OpenAI implementation
    anthropic.go     # Anthropic implementation
    models.go        # Model metadata/configs
  agent/
    loop.go          # Shared agentic loop
  prompts/
    system.go        # Prompt templates
```

## Key Takeaway

> "Aider's success shows that a thin wrapper over mature abstraction is pragmatic. The complexity shifts to prompt engineering, not API plumbing."
