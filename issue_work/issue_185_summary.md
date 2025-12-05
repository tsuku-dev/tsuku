# Issue 185 Summary

## What Was Implemented

Added 9 new recipe files for popular Go CLI tools using the `go_install` action. These tools build from source with proper GOBIN/GOPATH isolation, supporting the Go ecosystem in tsuku.

## Changes Made

New files created in `internal/recipe/recipes/`:
- `c/cobra-cli.toml`: CLI scaffolding tool for Cobra applications
- `d/dlv.toml`: Delve Go debugger
- `g/godoc.toml`: Go documentation server
- `g/gofumpt.toml`: Stricter gofmt formatter
- `g/goimports.toml`: Import formatter
- `g/gopls.toml`: Official Go language server
- `g/gore.toml`: Go REPL
- `m/mockgen.toml`: Mock generator for Go interfaces
- `s/staticcheck.toml`: Go static analysis tool

## Key Decisions

- **Tool selection**: Chose tools that are pure Go (no cgo), commonly used, and don't have existing pre-built binary recipes
- **Version source**: All recipes use `source = "goproxy"` for version resolution from proxy.golang.org
- **Verify patterns**: Used tool-specific patterns based on each tool's `--version` or `-h` output format

## Trade-offs Accepted

- **air skipped**: Already has a github_archive recipe; no need for duplicate go_install version
- **Verify patterns may need refinement**: Some patterns use generic matches (e.g., `usage:` for help output) that may need adjustment based on actual tool behavior

## Test Coverage

- No new test code added (recipe files only)
- Existing recipe validation tests cover the new TOML files
- All 721 tests pass

## Known Limitations

- Tools require Go toolchain to be installed first (handled via `dependencies = ["go"]`)
- Compilation time varies by tool size (addressed in Go ecosystem design doc)

## Future Improvements

- Consider adding more Go tools as demand arises
- Integration tests could verify actual go_install execution (T53 in test-matrix covers gofumpt)
