# Mockgen Recipe CI Failure Investigation

## Current Configuration

**Recipe File**: `internal/recipe/recipes/m/mockgen.toml`

```toml
[metadata]
name = "mockgen"
description = "Mock generator for Go interfaces"
homepage = "https://github.com/uber-go/mock"
version_format = "semver"

[version]
source = "goproxy"
module = "go.uber.org/mock"

[[steps]]
action = "go_install"
module = "go.uber.org/mock/mockgen"
executables = ["mockgen"]

[verify]
command = "mockgen --version"
pattern = "{version}"
```

## Error Details

The verification step uses `mockgen --version` to extract and verify the installed version. However, this command is failing in CI because mockgen uses a single-dash flag format.

### Timeline

- **PR #190**: Added `module` field support to version section for goproxy source (commit 12b5638)
- **Commit 12b5638**: First introduction of mockgen recipe with `go.uber.org/mock` as version module and `go.uber.org/mock/mockgen` as install path
- **Current Status**: Recipe exists but verification fails in test-changed-recipes workflow

## Root Cause

According to the [official uber-go/mock repository documentation](https://github.com/uber-go/mock), the correct command to verify mockgen installation is:

```bash
mockgen -version
```

Not:

```bash
mockgen --version
```

The recipe incorrectly uses the double-dash flag `--version`, which mockgen does not recognize. This causes the verification command to fail during CI testing in the `test-changed-recipes.yml` workflow.

## Verification

The uber-go/mock README explicitly states:

> "To ensure it was installed correctly, use: `mockgen -version`"

This is confirmed by inspecting the mockgen source code, which accepts the single-dash `-version` flag for version output.

## Recommended Fix

Change the verify command in `internal/recipe/recipes/m/mockgen.toml`:

**Current (incorrect)**:
```toml
[verify]
command = "mockgen --version"
pattern = "{version}"
```

**Should be (correct)**:
```toml
[verify]
command = "mockgen -version"
pattern = "{version}"
```

### Impact

- This fix will allow the mockgen recipe to pass verification in CI
- The `test-changed-recipes.yml` workflow will successfully verify mockgen installation on both Linux and macOS runners
- The recipe will be consistent with the official upstream documentation

## Related Recipes

Other Go tool recipes using `go_install` action were added in the same commit (PR #190) but they use different verification commands:
- `gofumpt --version` (double-dash)
- `staticcheck --version` (double-dash)
- `gore --version` (double-dash)

These should be validated as well to ensure they support double-dash format, or similarly updated if they also use single-dash format upstream.

## Testing Strategy

1. Update `mockgen --version` → `mockgen -version` in the recipe
2. Run `go build -o tsuku ./cmd/tsuku`
3. Test with: `./tsuku install --force mockgen`
4. Verify with: `mockgen -version`
