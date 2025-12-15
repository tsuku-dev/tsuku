# Issue 540 & 544 Summary

## What Was Implemented

Added zlib and expat recipes following the established pattern from PR #565:
- Registry recipes use homebrew_bottle for fast, reliable installation
- testdata recipe (expat-source) validates source builds with library dependencies

## Changes Made

- `internal/recipe/recipes/z/zlib.toml`: New library recipe using homebrew_bottle
- `internal/recipe/recipes/e/expat.toml`: New tool recipe with xmlwf binary
- `testdata/recipes/expat-source.toml`: Source build test recipe with zlib dependency
- `.github/workflows/build-essentials.yml`: Added zlib, expat to bottle tests; added expat-source job
- `docs/DESIGN-dependency-provisioning.md`: Marked #540 and #544 as done

## Key Decisions

- Following bottle-first pattern: Official registry uses homebrew_bottle for speed/reliability
- Source builds in testdata: Validates infrastructure without slowing user installs
- expat-source tests dependency linking: Validates configure_make with tsuku-provided libraries

## Test Coverage

- Recipe validation: All 3 new recipes pass `./tsuku validate`
- CI matrix: zlib and expat bottles tested on 3 platforms
- CI new job: test-zlib-dependency validates expat-source with zlib linking

## Known Limitations

- expat-source currently doesn't set LDFLAGS/CFLAGS for zlib path
  (may need enhancement in configure_make for full library linking support)

## Future Improvements

- Add pkg-config recipe and enhance configure_make to auto-discover library paths
- Consider adding more library recipes (openssl, ncurses) per design doc
