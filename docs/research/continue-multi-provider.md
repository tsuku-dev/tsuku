# Continue IDE Multi-Provider Architecture Research

## Architecture Pattern: Shared Loop + Provider Adapters

Three-tier architecture:
1. **IDE Layer** (`extensions/vscode`, `extensions/intellij`) - IDE-specific integration
2. **Core Layer** (`core/`) - Platform-agnostic business logic, conversation loop, tool orchestration
3. **GUI Layer** (`gui/`) - React-based UI

Flow: `GUI <-> Extension <-> Core <-> LLM Providers`

## Where Agent Loop Lives

**Centralized in `core/llm/streamChat.ts`** - the `llmStreamChat` async generator function.

The loop is **provider-agnostic** - it delegates actual LLM communication to provider implementations via the `model.streamChat()` interface.

## Provider Interface Design

### Two-Layer Architecture

**Layer 1: Core Interface (`ILLM`)**
- `streamChat(messages, signal, options): AsyncGenerator<ChatMessage>`
- `streamComplete(prompt, signal, options): AsyncGenerator<string>`
- `chat(messages, signal, options): Promise<ChatMessage>`
- `countTokens(text): number`
- `supportsImages(): boolean`
- `supportsFim(): boolean`
- 12+ methods total

**Layer 2: Base Implementation (`BaseLLM`)**
- Template method pattern - defines skeleton, providers override `_streamChat()`
- Integration with OpenAI adapter layer
- Common functionality (error handling, usage tracking)

**Provider Registration:**
```typescript
export const LLMClasses = [
  Anthropic, OpenAI, Gemini, Ollama, /* ... 60+ providers */
];
```

## Tool Calling Abstraction

**Unified Tool Definition:**
```typescript
interface Tool {
  type: "function",
  function: {
    name: string,
    description: string,
    parameters: JSONSchema
  }
}
```

**Provider-Specific Adaptation:**
- OpenAI: Uses native function calling format directly
- Anthropic: Converts to `tool_use` content blocks
- Other Providers: Adapters normalize to OpenAI-compatible format

**Key Insight:** Tool execution is **separate from LLM interaction** - the core orchestrates the loop, providers just return tool call requests.

## OpenAI Adapter Pattern

The `@continuedev/openai-adapters` package provides decorators that convert provider APIs to OpenAI format.

**Routing Logic:**
```typescript
if (shouldUseOpenAIAdapterFor(provider)) {
  // Route through adapter layer
} else {
  // Use provider's native implementation
}
```

**AnthropicApi Adapter:**
- System messages -> `system` parameter
- Tool calls -> `tool_use` content blocks
- Streaming events -> OpenAI chunk format

## Pros and Cons

**Pros:**
- 60+ providers supported
- Clean separation: core loop is provider-agnostic
- Tool execution separated from LLM calls
- Extensible (add provider = extend BaseLLM, register in array)
- Unified tool definition across all providers

**Cons:**
- Complexity (3-tier architecture, Redux state, protocols)
- Known tool execution bugs (loops, session termination)
- TypeScript-specific patterns
- Feature parity challenges (not all providers support all features)
- Provider-specific workarounds needed

## Key Lessons for Go CLI

**What to Adopt:**

1. **Shared Loop Pattern**
   - Conversation loop in platform-agnostic core
   - Providers implement minimal interface
   - Core orchestrates: input -> LLM -> tools -> LLM -> output

2. **Minimal Interface** (vs ILLM's 12+ methods)
   ```go
   type Provider interface {
       StreamChat(ctx context.Context, messages []Message, tools []Tool) (<-chan Delta, error)
       CountTokens(text string) int
   }
   ```

3. **Unified Tool Definition**
   ```go
   type Tool struct {
       Name        string
       Description string
       Parameters  JSONSchema
       Execute     func(args map[string]any) (string, error)
   }
   ```

4. **Tool Execution Separate from LLM**
   ```go
   for {
       response := provider.StreamChat(messages, tools)
       if response.ToolCalls != nil {
           result := ExecuteTool(call.Name, call.Args)  // Separate
           messages = append(messages, ToolResultMessage(result))
           continue
       }
       return response.Content
   }
   ```

**What NOT to Copy:**

1. Don't build adapter layer - 2 providers = explicit conversion is simpler
2. Don't separate into 3 tiers - CLI has no GUI
3. Don't use template method pattern - adds complexity for 2 providers
4. Don't support 60+ providers - stick to scope

## Key Takeaway

> "Adopt Continue's shared loop + provider interface + tool abstraction pattern, but simplify aggressively: no adapter layer, no multi-tier architecture, minimal provider interface."
