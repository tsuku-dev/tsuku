# Issue 563 Implementation Plan

## Summary

Implement `tsuku check-deps <recipe>` command that resolves all dependencies transitively, classifies each as provisionable or system-required, checks their status, and outputs a colorized status report with non-zero exit on missing system dependencies.

## Approach

Use existing patterns from `info.go` for dependency resolution and `require_system.go` for system dependency detection. Create a new `check_deps.go` file following CLI command conventions. Classify dependencies by loading each dependency's recipe and checking if all steps are `require_system` actions (pattern from `isSystemDependencyPlan()`). Use ANSI escape codes for colorized output since no color library is currently in use.

### Alternatives Considered

- **Reuse install_deps logic directly**: Not chosen because `check-deps` is read-only and should not trigger installation, but the pattern from `isSystemDependencyPlan()` will be extracted and shared.
- **Add to info command as subcommand**: Not chosen because `info` focuses on a single tool while `check-deps` analyzes the full dependency tree.
- **Use a color library (fatih/color)**: Not chosen to avoid adding a new dependency when ANSI codes suffice for the simple use case.

## Files to Modify

- `cmd/tsuku/main.go` - Add `checkDepsCmd` to command registration

## Files to Create

- `cmd/tsuku/check_deps.go` - New command implementation (~200-250 lines)
- `cmd/tsuku/check_deps_test.go` - Unit tests for the command

## Implementation Steps

- [ ] Create `cmd/tsuku/check_deps.go` with command structure
  - Define `checkDepsCmd` cobra command with `check-deps <recipe>` usage
  - Add `--json` flag for JSON output mode
  - Add command to root in `init()`

- [ ] Implement `DepStatus` struct and classification logic
  - Create `DepStatus` struct with Name, Type, Status, Version, Required, InstallGuide fields
  - Create `classifyDependency()` function that loads a dependency's recipe and returns "system-required" if all steps are `require_system`, otherwise "provisionable"
  - Handle recipe-not-found case as "unknown" type

- [ ] Implement status detection for each dependency type
  - For system-required deps: Create `checkSystemDep()` that runs `require_system` detection logic (exec.LookPath + version check)
  - For provisionable deps: Create `checkProvisionableDep()` that checks `mgr.GetState()` for installation status
  - Return appropriate status: "installed", "missing", "version_mismatch"

- [ ] Implement the main `runCheckDeps()` function
  - Load recipe via `loader.Get()`
  - Resolve direct deps via `actions.ResolveDependencies()`
  - Resolve transitive deps via `actions.ResolveTransitive()`
  - Iterate over all deps, classify, and check status
  - Track if any system dependency is missing for exit code

- [ ] Implement colorized output
  - Define ANSI color constants (green, red, yellow, reset)
  - Create output formatting functions: `printDepStatus()`, `printSummary()`
  - Use colors: green for installed, red for missing, yellow for version mismatch
  - Mark system-required vs provisionable clearly in output

- [ ] Implement JSON output mode
  - Define JSON output struct with tool name and dependencies array
  - Add `--json` flag handling
  - Output JSON when flag is set

- [ ] Handle exit codes
  - Exit 0 if all dependencies satisfied
  - Exit 1 (`ExitGeneral`) if any system dependency is missing
  - Provisionable missing deps don't cause non-zero exit (they can be auto-installed)

- [ ] Register command in main.go
  - Add `rootCmd.AddCommand(checkDepsCmd)` in `init()`

- [ ] Write unit tests
  - Test dependency classification (system-required vs provisionable)
  - Test status detection with mock state
  - Test exit code behavior
  - Test JSON output format

## Testing Strategy

- **Unit tests**: Test classification logic with mock recipes (system-required vs provisionable), test status detection with mock state manager
- **Integration tests**: Test with real recipes (docker.toml for system-required, nodejs.toml for provisionable)
- **Manual verification**:
  - `tsuku check-deps docker` shows system-required dependency
  - `tsuku check-deps nodejs` shows provisionable dependency
  - Exit code is 1 when docker is not installed
  - Output is properly colorized

## Risks and Mitigations

- **Dependency recipe not found**: Treat as "unknown" type with warning message rather than failing. The command should still report on known dependencies.
- **Performance with deep transitive chains**: Use the existing `MaxTransitiveDepth` limit (10) from resolver.go to prevent infinite loops.
- **Color output in non-TTY**: Check if stdout is a terminal before using colors (or let users disable with `--no-color` in future enhancement).

## Success Criteria

- [ ] `tsuku check-deps <recipe>` shows all dependencies (direct and transitive)
- [ ] Each dependency shows type (provisionable vs system-required)
- [ ] Each dependency shows status (installed, missing, version mismatch)
- [ ] Missing system dependencies cause exit code 1
- [ ] Output is colorized (green=installed, red=missing, yellow=version mismatch)
- [ ] `--json` flag outputs structured JSON
- [ ] All existing tests pass
- [ ] New unit tests cover core logic

## Open Questions

None. All prerequisite work (#560, #643) is complete and patterns are well-established.
