# Issue 36 Implementation Plan

## Summary

Add `TSUKU_HOME` environment variable support by modifying `DefaultConfig()` to check for this env var before falling back to `~/.tsuku`, and update hardcoded paths in `findPythonStandalone` and `gem_install.go` to use config-based paths.

## Approach

Follow the established pattern used by `TSUKU_REGISTRY_URL` in `internal/registry/registry.go` - check for an environment variable and fall back to the default value. Since `DefaultConfig()` is the single source of truth for the tsuku home directory, updating it will automatically flow to all callers.

For the hardcoded paths in actions, add a `ToolsDir` field to `ExecutionContext` to provide access to the tools directory without requiring direct access to config.

### Alternatives Considered

- **Add TSUKU_HOME to config.go only**: Would not fix hardcoded paths in actions. The actions would still use `os.UserHomeDir()` directly.
- **Pass full Config to ExecutionContext**: Too heavy - actions only need ToolsDir. Passing the full config would create unnecessary coupling.

## Files to Modify

- `internal/config/config.go` - Add `TSUKU_HOME` env var check in `DefaultConfig()`
- `internal/config/config_test.go` - Add test for env var override behavior
- `internal/actions/action.go` - Add `ToolsDir` field to `ExecutionContext`
- `internal/actions/gem_install.go:131-132` - Use `ctx.ToolsDir` instead of hardcoded path
- `internal/install/manager.go:318-325` - Use `cfg.ToolsDir` instead of hardcoded path (requires passing config to `findPythonStandalone`)
- `internal/executor/executor.go` - Set `ToolsDir` in `ExecutionContext`

## Files to Create

None

## Implementation Steps

- [x] Add `TSUKU_HOME` env var support to `DefaultConfig()`
- [x] Add `ToolsDir` field to `ExecutionContext` in `action.go`
- [x] Update executor to set `ToolsDir` in context
- [x] Update `gem_install.go` to use `ctx.ToolsDir`
- [x] Update `findPythonStandalone` in `manager.go` to use config
- [x] Add unit tests for `TSUKU_HOME` env var behavior
- [x] Verify all tests pass

Mark each step [x] after it is implemented and committed. This enables clear resume detection.

## Testing Strategy

- Unit tests: Add tests in `config_test.go` that:
  1. Set `TSUKU_HOME` env var and verify `DefaultConfig()` uses it
  2. Verify unset env var falls back to `~/.tsuku`
  3. Verify all derived paths (ToolsDir, etc.) are based on the custom home

- Manual verification:
  1. `TSUKU_HOME=/tmp/test-tsuku tsuku install <tool>` should install to `/tmp/test-tsuku/tools/`

## Risks and Mitigations

- **Breaking existing installations**: Mitigated by only checking env var, not changing default behavior
- **Executor doesn't have config access**: Mitigated by adding ToolsDir to ExecutionContext

## Success Criteria

- [x] `TSUKU_HOME` environment variable is checked in `DefaultConfig()` before falling back to `~/.tsuku`
- [x] Hardcoded path in `internal/install/manager.go:325` uses config instead
- [x] Hardcoded path in `internal/actions/gem_install.go:132` uses config instead
- [x] Unit tests verify env var override behavior
- [x] All existing tests pass
- [x] Build succeeds

## Open Questions

None - implementation path is clear.
