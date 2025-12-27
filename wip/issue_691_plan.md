# Issue 691 Implementation Plan

## Approach

Following the design in `docs/DESIGN-go-install-inference.md`, implement go_install version inference by:
1. Adding `go_install` to the actionInference map
2. Adding InferredGoProxyStrategy to the provider factory
3. Updating recipes to use inference
4. Updating documentation

## Files to Modify

| File | Change |
|------|--------|
| `internal/version/redundancy.go` | Add `go_install` to actionInference map |
| `internal/version/redundancy_test.go` | Update test for go_install inference |
| `internal/version/provider_factory.go` | Add InferredGoProxyStrategy |
| `internal/version/provider_factory_test.go` | Add tests for new strategy |
| `internal/recipe/recipes/g/gofumpt.toml` | Remove [version] entirely |
| `internal/recipe/recipes/g/gopls.toml` | Remove [version] entirely |
| `internal/recipe/recipes/c/cobra-cli.toml` | Remove [version] entirely |
| `internal/recipe/recipes/d/dlv.toml` | Remove source, keep module |
| `internal/recipe/recipes/g/goimports.toml` | Remove source, keep module |
| `internal/recipe/recipes/s/staticcheck.toml` | Remove source, keep module |
| `internal/recipe/recipes/g/gore.toml` | Remove source, keep module |
| `CONTRIBUTING.md` | Document go_install inference pattern |

## Implementation Steps

1. Add `"go_install": "goproxy"` to actionInference map in redundancy.go
2. Update redundancy_test.go to expect redundancy detection for go_install with goproxy
3. Add InferredGoProxyStrategy to provider_factory.go
4. Add tests for InferredGoProxyStrategy
5. Update simple recipes (gofumpt, gopls, cobra-cli) - remove [version]
6. Update complex recipes (dlv, goimports, staticcheck, gore) - remove source only
7. Update CONTRIBUTING.md documentation
8. Run tests and validate

## Testing Strategy

- Run existing tests to ensure no regressions
- Verify `go test ./...` passes
- Validate recipes still work with `./tsuku install --dry-run`
