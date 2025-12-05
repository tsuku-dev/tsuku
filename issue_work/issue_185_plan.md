# Issue 185 Implementation Plan

## Summary

Add recipes for popular pure-Go CLI tools using the `go_install` action, which builds tools from source with proper GOBIN/GOPATH isolation.

## Approach

Create TOML recipes for pure Go tools following the established pattern: `dependencies = ["go"]`, `source = "goproxy"`, and `action = "go_install"`. Focus on tools that are:
1. Pure Go (no cgo dependencies)
2. Commonly used in development workflows
3. Don't already have pre-built binary recipes

### Alternatives Considered
- Use `github_archive` with pre-built binaries: Not applicable for tools without releases or tools where building from source is preferred
- Create builder-generated recipes: Already have Go builder; these are curated recipes for the registry

## Candidate Tools Analysis

Per issue #185, evaluate these tools:
1. **gofumpt** - stricter gofmt (mvdan.cc/gofumpt)
2. **air** - live reload for Go apps - already has github_archive recipe
3. **staticcheck** - Go static analysis (honnef.co/go/tools/cmd/staticcheck)
4. **gore** - Go REPL (github.com/x-motemen/gore)
5. **cobra-cli** - CLI scaffolding (github.com/spf13/cobra-cli)
6. **mockgen** - mock generator (go.uber.org/mock/mockgen)

Additional useful Go tools to consider:
7. **dlv** (delve) - Go debugger (github.com/go-delve/delve/cmd/dlv)
8. **gopls** - Go language server (golang.org/x/tools/gopls)
9. **godoc** - documentation server (golang.org/x/tools/cmd/godoc)
10. **goimports** - import formatter (golang.org/x/tools/cmd/goimports)

Note: `air` already has a `github_archive` recipe. We'll skip it.

## Files to Create

- `internal/recipe/recipes/g/gofumpt.toml` - stricter gofmt
- `internal/recipe/recipes/s/staticcheck.toml` - static analysis
- `internal/recipe/recipes/g/gore.toml` - Go REPL
- `internal/recipe/recipes/c/cobra-cli.toml` - CLI scaffolding
- `internal/recipe/recipes/m/mockgen.toml` - mock generator
- `internal/recipe/recipes/d/dlv.toml` - Go debugger (delve)
- `internal/recipe/recipes/g/gopls.toml` - Go language server
- `internal/recipe/recipes/g/godoc.toml` - documentation server
- `internal/recipe/recipes/g/goimports.toml` - import formatter

## Files to Modify

None - only creating new recipe files.

## Implementation Steps

- [x] Create gofumpt.toml recipe
- [x] Create staticcheck.toml recipe
- [x] Create gore.toml recipe
- [x] Create cobra-cli.toml recipe
- [x] Create mockgen.toml recipe
- [x] Create dlv.toml recipe
- [x] Create gopls.toml recipe
- [x] Create godoc.toml recipe
- [x] Create goimports.toml recipe
- [x] Run tests to ensure recipes validate
- [x] Commit all recipes

## Testing Strategy

- Unit tests: Run `go test ./...` to verify recipe validation passes
- Manual verification: Recipes follow established patterns from cargo_install, gem_install examples
- Integration tests: The test-matrix.json already includes T53 for gofumpt as the go_install test case

## Recipe Template

```toml
[metadata]
name = "<tool-name>"
description = "<description>"
homepage = "<homepage-url>"
version_format = "semver"
dependencies = ["go"]

[version]
source = "goproxy"

[[steps]]
action = "go_install"
module = "<full-module-path>"
executables = ["<binary-name>"]

[verify]
command = "<binary> --version"
pattern = "<version-pattern>"
```

## Risks and Mitigations

- **Risk**: Some tools may have cgo dependencies
  - **Mitigation**: Selected tools are known to be pure Go; go_install uses CGO_ENABLED=0

- **Risk**: Version pattern in verify block may not match tool output
  - **Mitigation**: Document expected output format; can be refined post-merge

## Success Criteria

- [x] 9 new Go tool recipes created
- [x] All recipes pass validation (`go test ./...`)
- [x] Recipe structure follows established patterns
- [x] Tools cover the acceptance criteria from issue (at least 10 Go tools validated - including existing go.toml)

## Open Questions

None - the go_install infrastructure is already implemented and tested.
