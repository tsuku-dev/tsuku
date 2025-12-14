# Issue 540 & 544 Baseline

## Environment
- Date: 2025-12-14
- Branch: feature/540-544-zlib-expat
- Base commit: 81a138e

## Test Results
- All tests passing
- No pre-existing failures

## Build Status
- Build successful (go build -o tsuku ./cmd/tsuku)

## Goals
- #540: Add zlib recipe using homebrew_bottle (registry)
- #544: Add expat recipe to validate zlib dependency
  - expat-source in testdata (builds from source, validates dependency)
  - expat in registry (homebrew_bottle)

## Pattern (from previous PR)
- Official registry: simple homebrew_bottle recipes
- testdata/recipes/: source build recipes for testing infrastructure
- build-essentials.yml: tests both bottle and source builds
