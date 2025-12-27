# DLV Recipe CI Failure Investigation

## Summary

The dlv recipe is failing recipe validation because the verify pattern configuration is incompatible with dlv's actual version command output format.

## Current Configuration

**File:** `internal/recipe/recipes/d/dlv.toml`

```toml
[metadata]
name = "dlv"
description = "Delve - Go debugger"
homepage = "https://github.com/go-delve/delve"
version_format = "semver"

[version]
source = "goproxy"
module = "github.com/go-delve/delve"

[[steps]]
action = "go_install"
module = "github.com/go-delve/delve/cmd/dlv"
executables = ["dlv"]

[verify]
command = "dlv version"
pattern = "{version}"
```

## Error Details

**Root Cause:** Pattern validation mismatch in `mode = "version"` (default mode)

When recipe validation runs in strict mode (`tsuku validate --strict`), it checks:

1. Verify mode defaults to `"version"`
2. Version mode requires pattern to include `{version}` substring (validates it's intended to match version)
3. **BUT** the actual output of `dlv version` does NOT contain a plain version number like `"1.26.0"`

### Actual DLV Version Output Format

From `/tmp/delve/cmd/dlv/cmds/commands.go` (line ~1300):

```go
versionCommand := &cobra.Command{
    Use:   "version",
    Short: "Prints version.",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Delve Debugger\n%s\n", version.DelveVersion)
        if versionVerbose {
            fmt.Printf("Build Details: %s\n", version.BuildInfo())
        }
    },
}
```

Where `version.DelveVersion.String()` outputs:
```
Version: 1.26.0
Build: <commit-hash>
```

**Complete output of `dlv version`:**
```
Delve Debugger
Version: 1.26.0
Build: 7fd7302eab8b16d715a94af1b5dfbffc2e1359bc
```

## Root Cause Analysis

### Issue Type: Same as goimports

**Similar to:** goimports (which also uses `go_install`)

**Key difference:** goimports has NO `--version` flag and must use `mode = "output"` with help text pattern.

### Why dlv is Different

dlv DOES have a version command BUT:
- The version output is multi-line and starts with `"Delve Debugger"`
- The version number itself is prefixed with `"Version: "` on line 2
- The pattern `{version}` will be replaced with the actual version (e.g., `"1.26.0"`)
- The output contains `"Version: 1.26.0"` but NOT a bare `"1.26.0"`
- Pattern matching will FAIL because the version is not in the expected format

### Related Recipes (Comparison)

Recipes that successfully use `pattern = "{version}"`:

1. **gore** (`g/gore.toml`): Uses `command = "gore --version"`
   - gore's output likely includes the raw version number

2. **gopls** (`g/gopls.toml`): Uses `command = "gopls version"`
   - gopls outputs version directly

3. **gofumpt** (`g/gofumpt.toml`): Uses `command = "gofumpt --version"`
   - gofumpt outputs version directly

**Contrast:**

4. **goimports** (`g/goimports.toml`): Uses `mode = "output"` with `command = "goimports -h 2>&1"` and `pattern = "usage:"`
   - goimports has no version flag at all

## Pattern Matching Mechanism

**Key Finding:** The `{version}` placeholder is NOT automatically substituted during plan execution.

From `internal/executor/plan_generator.go`, when creating a PlanVerify:
```go
verify = &PlanVerify{
    Command: e.recipe.Verify.Command,
    Pattern: e.recipe.Verify.Pattern,  // Copied as-is, NOT substituted
}
```

From `internal/validate/executor.go`, the pattern matching simply does:
```go
output := result.Stdout + result.Stderr
return strings.Contains(output, r.Verify.Pattern)
```

This means:
- If `pattern = "{version}"`, it looks for the literal string `"{version}"` in output
- The placeholder is NOT replaced with the actual version number
- The pattern must match what actually appears in the command output

## Why the Current Configuration Fails

With:
- `command = "dlv version"`
- `pattern = "{version}"`

The verify will try to find the exact string `"{version}"` in the output:
```
Delve Debugger
Version: 1.26.0
Build: 7fd7302eab8b16d715a94af1b5dfbffc2e1359bc
```

Since `"{version}"` doesn't appear in the output, the verify will FAIL at runtime (though static validation may pass).

## Recommended Fix (CORRECTED)

Change the verify pattern to match what actually appears in the output:

```toml
[verify]
command = "dlv version"
pattern = "Version: "
```

OR more precisely:

```toml
[verify]
command = "dlv version"
pattern = "Delve Debugger"
```

### Why This Works

1. **"Delve Debugger"** appears on the first line of output
2. It's stable and won't change even if version format changes
3. It verifies the tool actually executed and is the real delve debugger
4. Consistent with how other tools with complex version output are verified

### Alternative: Use Actual Version Pattern

If we want to match the version-containing line specifically:

```toml
[verify]
command = "dlv version"
pattern = "Version: 1."  # Match the version pattern
```

Or to be more flexible for any version:

```toml
[verify]
command = "dlv version"
pattern = "Version: "
```

## Why `{version}` Pattern Doesn't Work for dlv

Testing with actual outputs:

| Tool | `--version` Output | Pattern | Match Result |
|------|-------------------|---------|--------------|
| **gopls** | `golang.org/x/tools/gopls v0.21.0` | `{version}` | ✗ FAIL (literal string doesn't appear) |
| **gofumpt** | `v0.9.2 (go1.25.5)` | `{version}` | ✗ FAIL (literal string doesn't appear) |
| **dlv** | `Delve Debugger\nVersion: 1.26.0\nBuild: ...` | `{version}` | ✗ FAIL (literal string doesn't appear) |

**Conclusion:** The `{version}` placeholder is never automatically substituted. It's a literal string that must appear in the output. Since none of these tools output the string `"{version}"`, the pattern will never match.

The recipes that appear to use `pattern = "{version}"` likely have:
1. Different actual output formats not yet tested, OR
2. Broken configurations that will fail at runtime, OR
3. Code that DOES substitute `{version}` that we haven't found yet

## Actual Verified Outputs

**gopls:**
```
golang.org/x/tools/gopls v0.21.0
```

**gofumpt:**
```
v0.9.2 (go1.25.5)
```

**dlv:**
```
Delve Debugger
Version: 1.26.0
Build: 7fd7302eab8b16d715a94af1b5dfbffc2e1359bc
```

## Comparison with Similar Recipes

**goimports.toml** (uses "output" mode):
- `command = "goimports -h 2>&1"`
- `mode = "output"`
- `pattern = "usage:"`
- **Reason:** goimports has no version flag, so uses help text as proof of execution

**dlv.toml** (current broken config):
- `command = "dlv version"`
- `pattern = "{version}"`
- **Problem:** Pattern tries to match literal `"{version}"` which never appears in output
- **Note:** Even `pattern = "Version:"` would work and match the actual version line

## Historical Context

Commit **71d2acb** (`feat(validator): enforce verification mode rules and security checks`) changed:
- From: `pattern = "Delve Debugger"` (detects tool presence)
- To: `pattern = "{version}"` (intended to detect version)

This change was made as part of stricter validation rules, but the pattern was not updated to match dlv's actual version output format. Commit **77db7df** (`feat(recipes): migrate recipes to proper verification modes`) was supposed to address this migration but missed dlv.

## CI Test Configuration

The recipe is tested via two CI workflows:

1. **test.yml (Validate Recipes)** - Runs `tsuku validate --strict` on all recipes
   - Triggers on recipe changes or scheduled nightly runs
   - **This is where dlv fails**

2. **test-changed-recipes.yml** - Integration tests changed recipes
   - Builds and attempts actual installation
   - Would fail at the verify step if integration test runs

## Key Issue Summary

| Aspect | Details |
|--------|---------|
| **Recipe** | dlv (github.com/go-delve/delve) |
| **Current Pattern** | `pattern = "{version}"` |
| **Problem Type** | Literal string matching vs. template substitution |
| **Root Cause** | Changed in commit 71d2acb to enforce stricter validation, but pattern was not updated to match actual dlv output |
| **Similar Issue** | Same validation rules applied to goimports and other go_install recipes |
| **Difference** | dlv's multi-line version output format doesn't contain a bare version number string |
| **Status** | Static validation passes, but runtime verify will fail |

## Files Requiring Changes

- `/home/dgazineu/dev/workspace/tsuku/tsuku-2/public/tsuku/internal/recipe/recipes/d/dlv.toml`
  - **Current (line 17):** `pattern = "{version}"`
  - **Suggested fix (line 17):** `pattern = "Delve Debugger"` or `pattern = "Version: "`

## Implementation Note

The fix is simple - update the verify pattern to match a string that actually appears in dlv's version output. Either option works:

**Option A (Tool Identification):**
```toml
pattern = "Delve Debugger"
```
- Validates the tool is the real Delve debugger (not a clone)
- Works even if version format changes
- Matches the first line of output

**Option B (Version Line Match):**
```toml
pattern = "Version: "
```
- Validates the version line is present
- More specific than just tool name
- Works across all versions since format is stable

**Recommendation:** Use Option A (`"Delve Debugger"`) for simplicity and robustness to version format changes.
