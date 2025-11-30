# Issue #43: Add npm Builder for Node.js Package Recipes - Baseline

## Issue Summary
Implement NpmBuilder to generate recipes from npm package metadata.

## Acceptance Criteria
- [ ] `tsuku create prettier --from npm` generates valid recipe
- [ ] Generated recipe uses `npm_install` action
- [ ] Executable discovery uses `bin` field from package metadata
- [ ] Warning shown when executables inferred from package name
- [ ] Unit tests with mocked npm API responses
- [ ] `tsuku install prettier` successfully executes generated recipe

## Key Files to Reference
- `internal/builders/cargo.go` - CargoBuilder pattern
- `internal/builders/gem.go` - GemBuilder pattern
- `internal/builders/pypi.go` - PyPIBuilder pattern (most recent)
- `cmd/tsuku/create.go` - Builder registration
- `internal/version/npm.go` - Existing npm version provider

## npm Registry API
- Endpoint: `https://registry.npmjs.org/<package>`
- `bin` field in package.json directly exposes executables
- No need to fetch separate files (unlike PyPI/Gem)

## Branch
feature/43-npm-builder
