# goimports Recipe CI Failure Investigation

## Problem Statement

The goimports recipe fails in CI with error: `"unknown revision cmd/goimports/v0.40.0"` when attempting to install the tool. This indicates a version resolution failure when trying to fetch the module.

## Current Configuration

**File:** `internal/recipe/recipes/g/goimports.toml`

```toml
[metadata]
name = "goimports"
description = "Updates Go import lines and formats code"
homepage = "https://pkg.go.dev/golang.org/x/tools/cmd/goimports"
version_format = "semver"

[version]
source = "goproxy"
module = "golang.org/x/tools"

[[steps]]
action = "go_install"
module = "golang.org/x/tools/cmd/goimports"
executables = ["goimports"]

[verify]
command = "goimports -h 2>&1"
mode = "output"
pattern = "usage:"
exit_code = 2
reason = "goimports has no --version flag, verify via help output"
```

## Root Cause Analysis

### Issue 1: Version Resolution Module is Correct

The `[version]` section correctly specifies:
```toml
module = "golang.org/x/tools"
```

This is essential because:
- The goproxy version provider queries `proxy.golang.org/{module}/@latest`
- `golang.org/x/tools` is the parent module that has versions (e.g., v0.40.0)
- The subpackage `golang.org/x/tools/cmd/goimports` does NOT have separate versions in the Go proxy

**Verified:** Running `curl -s "https://proxy.golang.org/golang.org/x/tools/@v/list"` returns versions including `v0.40.0`. Running `curl -s "https://proxy.golang.org/golang.org/x/tools/cmd/goimports/@v/list"` returns `"not found: module golang.org/x/tools/cmd/goimports: no matching versions"`

### Issue 2: Installation Action Uses Correct Module

The `[[steps]]` section correctly specifies:
```toml
action = "go_install"
module = "golang.org/x/tools/cmd/goimports"
```

This is correct because:
- The `go_install` action installs the specific binary from the subpackage path
- The executable lives at `golang.org/x/tools/cmd/goimports`
- `go install golang.org/x/tools/cmd/goimports@v0.40.0` works correctly because it installs the subpackage at the parent module's version

### How Version Resolution Works

The **GoProxySourceStrategy** (in `internal/version/provider_factory.go`, lines 427-465) follows this priority:

1. **Check for explicit `[version.module]`** (line 452-454)
   - If specified, use it for version lookup
   - In goimports recipe: uses `golang.org/x/tools` for version resolution ✓

2. **Fall back to go_install module** (line 456-463)
   - Only if no explicit version module is set
   - Not applicable here since version.module is set

This two-tier approach is documented in the test (line 490-523):
```go
// Recipe where version module differs from install module
// This is the case for tools like staticcheck where:
// - Version lookup: honnef.co/go/tools
// - Install path: honnef.co/go/tools/cmd/staticcheck
```

The goimports recipe correctly uses this pattern (though version and install modules are the same parent).

## Likely Error Source

The error message `"unknown revision cmd/goimports/v0.40.0"` suggests:
- Something is attempting to resolve versions using `golang.org/x/tools/cmd/goimports` instead of `golang.org/x/tools`
- This happens during the `go install` step when trying to install at a specific version

### Possible Cause: Missing Version Field in go_install Action

Looking at similar recipes:
- **gopls.toml:** Uses `source = "goproxy"` with NO `module` field in `[version]`
- **gofumpt.toml:** Uses `source = "goproxy"` with NO `module` field in `[version]`
- **staticcheck.toml:** Uses `source = "goproxy"` WITH `module = "honnef.co/go/tools"` in `[version]`
- **mockgen.toml:** Uses `source = "goproxy"` WITH `module = "go.uber.org/mock"` in `[version]`

For gopls and gofumpt, the version source works WITHOUT an explicit module field because:
- The version provider falls back to using the go_install module path
- `golang.org/x/tools/gopls` and `mvdan.cc/gofumpt` are the actual module paths (not subpackages)

For goimports, the problem is:
- go_install module is `golang.org/x/tools/cmd/goimports` (a subpackage)
- But the actual versioned module is `golang.org/x/tools` (the parent)
- The explicit `module = "golang.org/x/tools"` in `[version]` section is correct

## Configuration Analysis

The goimports recipe has the correct configuration:

| Component | Value | Correctness |
|-----------|-------|------------|
| Version source | `goproxy` | ✓ Correct - Go module proxy |
| Version module | `golang.org/x/tools` | ✓ Correct - Parent module with versions |
| Install action | `go_install` | ✓ Correct - Go binary installation |
| Install module | `golang.org/x/tools/cmd/goimports` | ✓ Correct - Actual binary location |
| Verify command | `goimports -h` | ✓ Correct - Handles non-zero exit codes |

## Verification Results

### Local Testing

Tested locally with `./tsuku install --force goimports` on commit `af9620e`.

**Result:** FAILURE - Issue is real and reproducible locally

Error output:
```
go: golang.org/x/tools/cmd/goimports@v0.40.0: reading golang.org/x/tools/cmd/goimports/go.mod at revision cmd/goimports/v0.40.0: unknown revision cmd/goimports/v0.40.0
```

### Root Cause Identified

The problem is in the **go_install action's Decompose method** (in `internal/actions/go_install.go`, line 361):

```go
target := module + "@" + version  // Line 361
getCmd := exec.CommandContext(ctx.Context, goPath, "get", target)
```

**What happens:**
1. Recipe specifies `module = "golang.org/x/tools/cmd/goimports"` in go_install step
2. Version resolver correctly resolves to `v0.40.0` from `golang.org/x/tools` (version.module)
3. Decompose receives the go_install module: `golang.org/x/tools/cmd/goimports`
4. Decompose runs: `go get golang.org/x/tools/cmd/goimports@v0.40.0`
5. Go looks for git revision `cmd/goimports/v0.40.0` → does not exist
6. Error: "unknown revision cmd/goimports/v0.40.0"

**Why this happens:**
- `golang.org/x/tools/cmd/goimports` is a **subpackage**, not a module with its own versions
- It doesn't have separate version tags in git (like `cmd/goimports/v0.40.0`)
- The actual versions come from the parent module: `golang.org/x/tools`

**Affected recipes confirmed:**
- goimports (fails locally: `unknown revision cmd/goimports/v0.40.0`)
- staticcheck (fails locally with identical error pattern: `unknown revision cmd/staticcheck/v0.6.1`)
- Any other recipe using `[version.module]` set to a parent module while go_install uses a subpackage

### The Bug

The Decompose method in GoInstallAction doesn't use the recipe's `[version.module]` field. It only has access to the go_install module parameter.

**Current behavior:** Uses `go_install.module` for decomposition
**Required behavior:** Use `version.module` (if set) for go get, and keep go_install.module for the actual installation

The EvalContext does have access to the full Recipe (line 66 of `internal/actions/decomposable.go`):
```go
Recipe *recipe.Recipe
```

So the fix is available: Use `ctx.Recipe.Version.Module` if present, otherwise fall back to the module from step params.

## The Fix

**Location:** `internal/actions/go_install.go`, Decompose method, around line 361

**Current code:**
```go
target := module + "@" + version
getCmd := exec.CommandContext(ctx.Context, goPath, "get", target)
```

**Fixed code:**
```go
// Use version module if recipe specifies one (for subpackages)
moduleForVersioning := module
if ctx.Recipe != nil && ctx.Recipe.Version.Module != "" {
    moduleForVersioning = ctx.Recipe.Version.Module
}
target := moduleForVersioning + "@" + version
getCmd := exec.CommandContext(ctx.Context, goPath, "get", target)
```

This allows recipes to specify different modules for version resolution vs. installation, which is essential for Go subpackages that don't have their own version tags.

## Conclusion

**Issue Status:** CONFIRMED - Real bug in go_install Decompose method

**Recipe configuration:** Correct ✓
- `[version.module] = "golang.org/x/tools"` (correct parent module)
- `go_install.module = "golang.org/x/tools/cmd/goimports"` (correct subpackage)

**Code issue:** The Decompose method must be fixed to use `version.module` when present

**Impact:** Affects all Go tool recipes that use subpackages with different versioning modules (e.g., goimports, staticcheck)

**Fix Required:** Modify Decompose to respect recipe's version.module field during `go get` execution
