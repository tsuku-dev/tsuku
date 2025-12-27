# Implementation Plan: Issue #692 - Support file paths in tsuku eval command

## Summary

Add support for file path arguments to `tsuku eval`, allowing direct evaluation of recipe files without registry lookup. This aligns `eval` behavior with `validate` and enables efficient recipe development workflows.

## Current State Analysis

### eval command (`cmd/tsuku/eval.go`)
- Uses global `loader.Get(toolName)` to fetch recipes from registry only
- Has OS/Arch flag support via `--os` and `--arch` flags (lines 31-32, 63-64)
- Validates OS/Arch flags against whitelists (lines 69-90, 107-115)
- Passes OS/Arch to `PlanConfig` (lines 163-164)
- **Observation**: OS/Arch flag handling is already fully implemented and functional

### validate command (`cmd/tsuku/validate.go`)
- Takes file path directly as argument (line 31)
- Uses `recipe.ValidateFile(filePath)` which:
  1. Reads file with `os.ReadFile`
  2. Parses TOML with `toml.Unmarshal`
  3. Returns `ValidationResult` with `Recipe` pointer

### Recipe loading options
1. **`recipe.ValidateFile(path)`** - Returns `*ValidationResult` with embedded `*Recipe`
2. **`loader.parseBytes(data)`** - Private method on Loader
3. **Direct TOML parsing** - Use `toml.Unmarshal` + `validate()`

## Approach

### Option A: Add `ParseFile` function to recipe package (Recommended)
Create a new exported function `recipe.ParseFile(path string) (*Recipe, error)` that:
1. Reads the file
2. Parses TOML
3. Validates the recipe
4. Returns the parsed recipe

**Pros:**
- Clean separation of concerns
- Reusable across commands
- Follows existing patterns (ValidateFile does similar work)

**Cons:**
- New function to maintain

### Option B: Use ValidateFile and extract Recipe
Use existing `recipe.ValidateFile(path)` and access `result.Recipe`.

**Pros:**
- No new code in recipe package
- Reuses existing validation

**Cons:**
- Semantically awkward (validate vs parse)
- Extra validation overhead (we just need parsing + basic validation)

### Option C: Inline parsing in eval.go
Copy parsing logic directly into eval command.

**Pros:**
- Self-contained change

**Cons:**
- Code duplication
- Maintenance burden

**Decision:** Option A - Add `recipe.ParseFile` function

## Detection Logic

Distinguish file paths from registry names:
```go
func isFilePath(arg string) bool {
    return strings.Contains(arg, "/") || strings.HasSuffix(arg, ".toml")
}
```

This matches the issue specification and handles:
- Absolute paths: `/tmp/test-recipe.toml`
- Relative paths: `./my-recipe.toml`, `recipes/tool.toml`
- TOML files without path separator: `recipe.toml`

## Files to Modify

### 1. `internal/recipe/loader.go`
Add exported `ParseFile` function:
```go
// ParseFile parses a recipe from a file path.
// Unlike ValidateFile, this returns the parsed recipe directly for use in commands.
func ParseFile(path string) (*Recipe, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    var r Recipe
    if err := toml.Unmarshal(data, &r); err != nil {
        return nil, fmt.Errorf("failed to parse TOML: %w", err)
    }

    if err := validate(&r); err != nil {
        return nil, fmt.Errorf("recipe validation failed: %w", err)
    }

    return &r, nil
}
```

### 2. `cmd/tsuku/eval.go`
Modify `runEval` to detect file paths and load accordingly:

```go
func runEval(cmd *cobra.Command, args []string) {
    arg := args[0]

    // Validate platform flags early (existing code)
    if err := ValidateOS(evalOS); err != nil { ... }
    if err := ValidateArch(evalArch); err != nil { ... }

    var r *recipe.Recipe
    var recipeSource string
    var reqVersion string

    if isFilePath(arg) {
        // File path mode: load from file
        var err error
        r, err = recipe.ParseFile(arg)
        if err != nil {
            printError(err)
            exitWithCode(ExitGeneral)
        }
        recipeSource = arg  // Use file path as source
        // Note: version specification not supported with file paths
    } else {
        // Registry mode: existing behavior
        toolName := arg
        if strings.Contains(toolName, "@") {
            parts := strings.SplitN(toolName, "@", 2)
            toolName = parts[0]
            reqVersion = parts[1]
        }
        if reqVersion == "latest" {
            reqVersion = ""
        }

        var err error
        r, err = loader.Get(toolName)
        if err != nil { ... }
        recipeSource = "registry"
    }

    // Check platform support (existing)
    if !r.SupportsPlatformRuntime() { ... }

    // Create executor (existing, but use r directly)
    ...
}

func isFilePath(arg string) bool {
    return strings.Contains(arg, "/") || strings.HasSuffix(arg, ".toml")
}
```

### 3. `cmd/tsuku/eval.go` - Update command help
Update `Use` and `Long` descriptions:
```go
Use:   "eval <tool>[@version] | <path-to-recipe.toml>",
Long: `Generate a deterministic installation plan for a tool and output it as JSON.

...existing text...

File paths are detected by containing '/' or ending with '.toml':
  tsuku eval ./my-recipe.toml
  tsuku eval /path/to/recipe.toml
  tsuku eval recipes/tool.toml`,
```

### 4. `internal/recipe/loader_test.go`
Add tests for `ParseFile`:
- Valid recipe file
- Non-existent file
- Invalid TOML syntax
- Invalid recipe (missing required fields)

### 5. `cmd/tsuku/eval_test.go`
Add tests for:
- `isFilePath` detection logic
- Integration with file-based recipes (if feasible without mocking)

## Implementation Steps

1. **Add `recipe.ParseFile` function**
   - Implement in `internal/recipe/loader.go`
   - Add unit tests in `internal/recipe/loader_test.go`

2. **Add `isFilePath` helper function**
   - Implement in `cmd/tsuku/eval.go`
   - Add unit tests in `cmd/tsuku/eval_test.go`

3. **Modify `runEval` to handle file paths**
   - Early detection of file path vs registry name
   - Load recipe from file or registry
   - Set appropriate `RecipeSource` in PlanConfig

4. **Update command documentation**
   - Update `Use` field
   - Update `Long` description with examples

5. **Verify OS/Arch handling works correctly**
   - Test with file path + `--os darwin --arch arm64`
   - Confirm platform filtering applies to file-loaded recipes

6. **Run tests and validation**
   - `go vet ./...`
   - `go test ./...`
   - `go build ./cmd/tsuku`
   - Manual testing with sample recipe file

## Testing Strategy

### Unit Tests
1. **`TestParseFile_ValidRecipe`**: Parse a valid TOML recipe file
2. **`TestParseFile_FileNotFound`**: Error on missing file
3. **`TestParseFile_InvalidTOML`**: Error on malformed TOML
4. **`TestParseFile_MissingRequiredFields`**: Error on invalid recipe
5. **`TestIsFilePath`**: Detection logic for paths vs names

### Integration Tests
1. Create temporary recipe file
2. Run `tsuku eval /tmp/test.toml`
3. Verify JSON output contains expected fields
4. Test with `--os` and `--arch` flags to confirm platform override works

### Manual Testing
```bash
# Create test recipe
cat > /tmp/test-recipe.toml <<EOF
[metadata]
name = "test-tool"
description = "Test recipe"

[version]
source = "manual"

[[steps]]
action = "download"
url = "https://example.com/tool"
dest = "tool"

[verify]
command = "tool --version"
EOF

# Test file path detection
./tsuku eval /tmp/test-recipe.toml
./tsuku eval ./recipes/gh.toml --os darwin --arch arm64
./tsuku eval recipes/serve.toml

# Test registry mode still works
./tsuku eval gh
./tsuku eval gh@2.40.0
```

## Risks and Mitigations

### Risk 1: Breaking existing behavior
**Mitigation**: File path detection is strict (requires `/` or `.toml`). Registry names like `gh`, `kubectl`, `aws-cli` will never match.

### Risk 2: Ambiguous input like `mytool.toml` (no path separator)
**Mitigation**: The `.toml` suffix is explicit enough. Users expect `.toml` files to be treated as files.

### Risk 3: Security concerns with arbitrary file paths
**Mitigation**: Same exposure as `validate` command. Recipe parsing is sandboxed (no code execution during parse). Plan generation may trigger downloads but respects existing security model.

### Risk 4: Version specification with file paths
**Mitigation**: Document that `@version` is not supported with file paths. The recipe file defines the version configuration.

## Success Criteria

1. `tsuku eval /path/to/recipe.toml` generates valid JSON plan
2. `tsuku eval ./recipe.toml --os darwin --arch arm64` respects platform flags
3. `tsuku eval gh` (registry mode) continues to work unchanged
4. All existing tests pass
5. New tests cover file path parsing and detection logic
6. Error messages are clear when file is not found or invalid

## Estimated Effort

- Implementation: ~30-45 minutes
- Testing: ~20-30 minutes
- Total: ~1 hour
