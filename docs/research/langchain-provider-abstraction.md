# LangChain Provider Abstraction Research

## Architecture Pattern: Layered with Standardized Interface

LangChain uses a modular package structure:
- `langchain-core` - Base abstractions (`BaseChatModel`, Messages, LCEL)
- `langchain-openai`, `langchain-anthropic` - Provider implementations
- `langchain` - High-level components (Chains, Agents, AgentExecutor)

## Where Conversation Loop Lives

**Shared orchestration layer** - The loop lives in `AgentExecutor` or `LangGraph`:

```python
# AgentExecutor implements the loop
while not should_continue:
    output = agent.plan(intermediate_steps, messages)
    if isinstance(output, AgentFinish):
        return output
    observation = tool.run(output.tool_input)
    intermediate_steps.append((output, observation))
```

## Tool Use Abstraction (Key Innovation)

Three-part standardization:

1. **`bind_tools()` method** - Attaches tools to any chat model, handles provider format conversion
2. **`AIMessage.tool_calls`** - Unified output format regardless of provider
3. **`create_tool_calling_agent()`** - Universal agent builder

```python
# Works with ANY provider
openai_model = ChatOpenAI().bind_tools([get_weather])
anthropic_model = ChatAnthropic().bind_tools([get_weather])
```

## Recommendation for Tsuku

**Adopt:**
- Standardized tool call format (`ToolCall` struct with Name, Args, ID)
- Conversation loop in shared code, not per-provider
- Single interface with provider implementations

**Avoid:**
- Over-abstraction (don't create multiple packages for 2 providers)
- Complex chains/agents abstractions (YAGNI)
- Magic/hidden behavior

## Key Takeaway

> "LangChain's success comes from the standardized interface, not the layers."

For 2 providers: thin interface, shared loop, standardized tool call format.
