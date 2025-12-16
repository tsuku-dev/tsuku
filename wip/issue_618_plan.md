# Implementation Plan: Issue #618

Wire Orchestrator into tsuku create command.

## Overview

The Orchestrator exists (`internal/builders/orchestrator.go`) but is never instantiated. The `create.go` command calls `session.Generate()` directly, bypassing sandbox testing. The `--skip-sandbox` flag is dead code.

## Changes

### 1. Update Orchestrator to handle RepairNotSupportedError gracefully

**File**: `internal/builders/orchestrator.go`

Currently, when `session.Repair()` returns an error, the Orchestrator fails the entire operation (line 168-171). For ecosystem builders that use `DeterministicSession`, `Repair()` returns `RepairNotSupportedError`. Instead of failing, the Orchestrator should:
- Detect `RepairNotSupportedError`
- Log a warning that repair is not supported for this builder type
- Return `ValidationFailedError` with the sandbox result (so the user sees what failed)

```go
// In the repair loop (line 168-171):
result, err = session.Repair(ctx, sandboxResult)
if err != nil {
    // Check if this builder doesn't support repair
    if errors.Is(err, &RepairNotSupportedError{}) {
        // Can't repair - return validation failed error
        return nil, &ValidationFailedError{
            SandboxResult:  sandboxResult,
            RepairAttempts: repairAttempts,
        }
    }
    return nil, fmt.Errorf("repair attempt %d failed: %w", repairAttempts, err)
}
```

### 2. Wire Orchestrator into create.go

**File**: `cmd/tsuku/create.go`

Replace the direct `session.Generate()` call with Orchestrator usage:

a) Create sandbox executor and orchestrator (around line 290):
```go
// Create sandbox executor (if not skipping)
var sandboxExec *sandbox.Executor
if !skipSandbox {
    detector := validate.NewRuntimeDetector()
    sandboxExec = sandbox.NewExecutor(detector,
        sandbox.WithDownloadCacheDir(cfg.DownloadCacheDir),
        sandbox.WithLogger(logger))
}

// Create orchestrator
orchestrator := builders.NewOrchestrator(
    builders.WithSandboxExecutor(sandboxExec),
    builders.WithOrchestratorConfig(builders.OrchestratorConfig{
        SkipSandbox: skipSandbox,
        MaxRepairs:  builders.DefaultMaxRepairs,
    }),
)
```

b) Replace session creation and Generate() call (lines 297-326) with Orchestrator.Create():
```go
// Use orchestrator for full generate → sandbox → repair cycle
orchResult, err := orchestrator.Create(ctx, builder, buildReq, sessionOpts)
if err != nil {
    // Handle ValidationFailedError specially
    var valErr *builders.ValidationFailedError
    if errors.As(err, &valErr) {
        fmt.Fprintf(os.Stderr, "Error: recipe validation failed\n")
        fmt.Fprintf(os.Stderr, "Exit code: %d\n", valErr.SandboxResult.ExitCode)
        if valErr.SandboxResult.Stderr != "" {
            fmt.Fprintf(os.Stderr, "Error output:\n%s\n", valErr.SandboxResult.Stderr)
        }
        exitWithCode(ExitGeneral)
    }
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    exitWithCode(ExitGeneral)
}

result := orchResult.BuildResult
```

c) Remove the dead code:
```go
// Remove these lines:
forceInit := false
_ = forceInit
_ = skipSandbox
```

d) Update imports to include:
```go
"github.com/tsukumogami/tsuku/internal/sandbox"
"github.com/tsukumogami/tsuku/internal/validate"
```

### 3. Add Is() method to RepairNotSupportedError for errors.Is() support

**File**: `internal/builders/builder.go`

```go
func (e *RepairNotSupportedError) Is(target error) bool {
    _, ok := target.(*RepairNotSupportedError)
    return ok
}
```

## Testing

1. Run existing tests to ensure no regressions
2. Manual test: `./tsuku create rg --from crates.io` (ecosystem builder, no sandbox)
3. Manual test: `./tsuku create jq --from homebrew:jq --skip-sandbox` (LLM builder, skip sandbox)
4. Manual test without Docker: should warn and continue with `--skip-sandbox`

## Success Criteria

1. `tsuku create foo --from homebrew:bar` runs sandbox testing when Docker available
2. `tsuku create foo --from homebrew:bar --skip-sandbox` skips sandbox testing
3. Sandbox failure triggers repair cycle for LLM-based builders
4. Graceful degradation for ecosystem builders (validation fails if sandbox fails, no retry)
5. Missing Docker runtime warns but doesn't fail with `--skip-sandbox`
