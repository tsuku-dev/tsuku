# Issue 691 Summary

## What Was Implemented

Added version inference support for go_install action, enabling go_install recipes to omit explicit `source = "goproxy"` when the version source can be inferred. This brings go_install to parity with other ecosystem installers (cargo_install, pipx_install, npm_install, etc.) that already support version inference.

## Changes Made

- `internal/version/redundancy.go`: Added `go_install` to actionInference map
- `internal/version/redundancy_test.go`: Updated test to expect redundancy detection for go_install with goproxy
- `internal/version/provider_factory.go`: Added InferredGoProxyStrategy implementation
- `internal/version/provider_factory_test.go`: Added tests for InferredGoProxyStrategy
- `internal/recipe/recipes/g/gofumpt.toml`: Removed [version] section (simple case)
- `internal/recipe/recipes/g/gopls.toml`: Removed [version] section (simple case)
- `internal/recipe/recipes/c/cobra-cli.toml`: Removed [version] section (simple case)
- `internal/recipe/recipes/d/dlv.toml`: Removed source, kept module (complex case)
- `internal/recipe/recipes/g/goimports.toml`: Removed source, kept module (complex case)
- `internal/recipe/recipes/s/staticcheck.toml`: Removed source, kept module (complex case)
- `internal/recipe/recipes/g/gore.toml`: Removed source, kept module (complex case)
- `CONTRIBUTING.md`: Updated documentation to reflect new inference pattern

## Key Decisions

- **Leverage existing infrastructure**: Used existing `Recipe.Version.Module` field instead of adding a new `version_module` step parameter. The architecture review identified this as the simpler approach.
- **Two patterns for go_install**: Simple cases (matching paths) need no [version], complex cases (differing paths) need only `module = "..."` in [version]

## Trade-offs Accepted

- **Complex recipes still need `[version] module`**: For recipes where install path differs from versioning module (dlv, goimports, etc.), explicit module specification is required. This is non-redundant information so it's acceptable.

## Test Coverage

- New tests added: 5 (Priority, CanHandle, Create variations)
- Coverage change: No regression, tests cover all new code paths

## Known Limitations

- Recipes with differing install/version paths must still specify `[version] module = "..."`
- No automatic detection of when module differs from install path (by design - explicit is better)

## Future Improvements

- Could add validation warning when version inference fails to help recipe authors debug issues
