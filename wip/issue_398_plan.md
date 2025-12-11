# Issue 398 Implementation Plan: Split install.go into focused modules

## Goal

Split `cmd/tsuku/install.go` (603 lines) into focused modules to reduce merge conflicts and improve maintainability.

## Implementation Steps

### Step 1: Create install_deps.go

Extract dependency handling functions to `cmd/tsuku/install_deps.go`:

- `runInstallWithTelemetry()` - entry point wrapping installWithDependencies
- `installWithDependencies()` - core recursive installation logic (~290 lines)
- `ensurePackageManagersForRecipe()` - package manager bootstrap (~50 lines)
- `findDependencyBinPath()` - locate dependency bin directories (~25 lines)
- `resolveRuntimeDeps()` - resolve runtime dependencies (~25 lines)
- `mapKeys()` - helper function (~8 lines)

### Step 2: Create install_lib.go

Extract library installation logic to `cmd/tsuku/install_lib.go`:

- `installLibrary()` - handles library recipe installation (~65 lines)

### Step 3: Update install.go

Keep CLI concerns in `install.go`:

- Package declaration and imports
- `installDryRun`, `installForce` flag variables
- `installCmd` command definition
- `init()` with flag definitions
- `isInteractive()` - check if running interactively
- `confirmInstall()` - prompt for user confirmation
- `runDryRun()` - dry-run execution

### Step 4: Adjust imports

Each new file needs appropriate imports:
- `install_deps.go`: All current imports except `bufio` (only used in confirmInstall)
- `install_lib.go`: `fmt`, `config`, `executor`, `install`, `recipe`, `telemetry`
- `install.go`: Keep `bufio`, `cobra`, remove imports only used by extracted code

### Step 5: Verify

- Run `go build ./cmd/tsuku` to verify compilation
- Run `go test ./cmd/tsuku` to verify tests pass
- Run `go vet ./...` to check for issues
- Run `gofmt -w cmd/tsuku/` to format all files

## File Size Estimates

| File | Lines |
|------|-------|
| `install.go` | ~120 |
| `install_deps.go` | ~400 |
| `install_lib.go` | ~80 |

## Notes

- All files remain in `package main` - no import changes needed for callers
- Functions maintain their current signatures
- No functional changes - pure refactoring
