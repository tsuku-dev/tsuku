# LlamaIndex Provider Abstraction Research

## Architecture Pattern: Layered with Inheritance

- **Base Layer**: `LLM` class with `complete`, `chat`, `stream_complete`, `stream_chat`
- **Function Calling Layer**: `FunctionCallingLLM` extends `LLM`
- **Provider Layer**: `Anthropic`, `OpenAI`, `Vertex` inherit from `FunctionCallingLLM`

## Where Conversation Loop Lives

**Event-driven Workflow system**:

```python
class AgentWorkflow:
    def handle_llm_input(self, event):
        response = llm.chat_with_tools(tools, chat_history)
        if has_tool_calls(response):
            emit ToolCallEvent
        else:
            emit StopEvent

    def handle_tool_calls(self, event):
        results = execute_tools(event.tool_calls)
        emit InputEvent(results)  # Loop back
```

**State Management:**
- Serializable `Context` object persists history
- Stateless by default, explicit state passing

## Tool Use Abstraction

**Provider-Agnostic with Runtime Adaptation:**

```python
def chat_with_tools(
    self,
    tools: Sequence["BaseTool"],
    user_msg: Optional[str] = None,
    chat_history: Optional[List[ChatMessage]] = None,
    allow_parallel_tool_calls: bool = False,
) -> ChatResponse
```

- Schema auto-generation from tool definitions
- `get_tool_calls_from_response` abstracts parsing
- Falls back to text-based prompting for LLMs without native function calling

## Recommendation for Tsuku

**Adopt:**
- Interface-based abstraction (`LLM` interface)
- Tool abstraction with provider-specific conversion
- Simple agent loop (not complex workflow system)

```go
type LLM interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
    ChatWithTools(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
}
```

**Avoid:**
- Event-driven workflows (too complex for CLI)
- Class inheritance (use Go interfaces/composition)
- Implicit context objects (use explicit state)

## Key Takeaway

> "Provider-specific adapters hide format differences inside implementations."

Simple loop in application code, interface for providers, explicit state management.
