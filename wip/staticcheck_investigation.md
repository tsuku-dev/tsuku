# Staticcheck Recipe CI Failure Investigation

## Summary

The staticcheck recipe is failing in CI because the `go_install` action specifies an incorrect module path. The error occurs during plan decomposition when `go_build` generates a go.mod file with a non-existent module requirement.

## Current Configuration

**File:** `/home/dgazineu/dev/workspace/tsuku/tsuku-2/public/tsuku/internal/recipe/recipes/s/staticcheck.toml`

```toml
[metadata]
name = "staticcheck"
description = "Go static analysis tool"
homepage = "https://staticcheck.dev/"
version_format = "semver"

[version]
source = "goproxy"
module = "honnef.co/go/tools"

[[steps]]
action = "go_install"
module = "honnef.co/go/tools/cmd/staticcheck"
executables = ["staticcheck"]

[verify]
command = "staticcheck --version"
pattern = "{version}"
```

## Error Details

**CI Run:** `20511519816` - macOS: staticcheck integration test

**Error Output:**
```
Installation failed: step 1 (go_build) failed: go mod download failed: exit status 1
Output: # get https://proxy.golang.org/honnef.co/go/tools/cmd/staticcheck/@v/v0.6.1.mod
# get https://proxy.golang.org/honnef.co/go/tools/cmd/staticcheck/@v/v0.6.1.mod: 404 Not Found

...

go: honnef.co/go/tools/cmd/staticcheck@v0.6.1: reading honnef.co/go/tools/cmd/staticcheck/go.mod at revision cmd/staticcheck/v0.6.1: unknown revision cmd/staticcheck/v0.6.1
```

## Root Cause

The issue involves a **module path mismatch** between version detection and the build action:

### The Problem Flow

1. **Version detection (line 9)** correctly specifies: `module = "honnef.co/go/tools"`
   - This is the actual Go module in the honnef.co/go/tools repository
   - Located in `/go.mod` at the root of the repository

2. **Build action (line 13)** incorrectly specifies: `module = "honnef.co/go/tools/cmd/staticcheck"`
   - This is NOT a separate module in the Go registry
   - When `go_install` decomposes into `go_build`, it passes this incorrect module path

3. **During plan execution**, `go_build` (line 164 of go_build.go) creates:
   ```
   require honnef.co/go/tools/cmd/staticcheck v0.6.1
   ```

4. **When `go mod download` runs** (line 178 of go_build.go):
   - It tries to fetch `https://proxy.golang.org/honnef.co/go/tools/cmd/staticcheck/@v/v0.6.1.mod`
   - Returns 404 because this module path doesn't exist
   - Fails with "unknown revision cmd/staticcheck/v0.6.1"

### Why Other go_install Recipes Work

Recipes like dlv and gore use similar `/cmd/` paths but work differently:
- dlv recipe: `module = "github.com/go-delve/delve/cmd/dlv"` - Works because github.com/go-delve/delve is a single-module repo where the module root is at `/`
- gore recipe: `module = "github.com/x-motemen/gore/cmd/gore"` - Works because github.com/x-motemen/gore is a single-module repo

However, honnef.co/go/tools has:
- Module root: `honnef.co/go/tools` (only one go.mod file, at root)
- Binary location: `cmd/staticcheck/main.go` (subdirectory within the module)
- Import path: `honnef.co/go/tools/cmd/staticcheck` (used for `go install`)

The key difference: the `/cmd/staticcheck` path is NOT a separate module—it's a package within the module.

## Recommended Fix

The recipe should use the **module path** (not the import path) in the go_install action:

```toml
[[steps]]
action = "go_install"
module = "honnef.co/go/tools"
import_path = "honnef.co/go/tools/cmd/staticcheck"
executables = ["staticcheck"]
```

However, if `go_install` doesn't support a separate `import_path` field, the simplest fix is:

```toml
[[steps]]
action = "go_install"
module = "honnef.co/go/tools"
executables = ["staticcheck"]
```

And specify the full binary path in the build:
```bash
go install honnef.co/go/tools/cmd/staticcheck@v0.6.1
```

This works because `go install` automatically finds the binary in `cmd/staticcheck/` when given the module.

## Verification

Both commands work locally:
```bash
# Current incorrect recipe approach (fails in tsuku's decomposition)
go install honnef.co/go/tools/cmd/staticcheck@v0.6.1

# Correct approach via go get first
go get -d honnef.co/go/tools@v0.6.1
go install honnef.co/go/tools/cmd/staticcheck@v0.6.1
```

The issue is that tsuku's decomposition creates an invalid go.mod requirement.

## Architecture Context

The tsuku build system uses a two-phase approach:
1. **go_install (composite action)** - Called during recipe installation, decomposes into go_build
2. **Decompose phase** - go_install.Decompose() runs `go get <module>@<version>` to capture go.sum
3. **go_build (primitive action)** - Takes the module path and builds with locked dependencies
   - Creates a minimal go.mod with `require <module> <version>`
   - Runs `go mod download` to validate dependencies
   - Runs `go install <module>@<version>`

The problem: the module path passed to go_build must be the actual module, not a package path.

## References

- Staticcheck homepage: https://staticcheck.dev/
- Go tools module: https://pkg.go.dev/honnef.co/go/tools
- Repository: https://github.com/dominikh/go-tools
- Related code files:
  - `/internal/actions/go_install.go` (Decompose method, lines 272-402)
  - `/internal/actions/go_build.go` (module requirement generation, line 164)
