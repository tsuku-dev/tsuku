# Issue 221 Baseline

## Environment
- Date: 2025-12-07
- Branch: feature/221-library-dependency-resolution
- Base commit: 29e77492128005e2eb3f87e1a9ba5272b2be4a54

## Test Results
- Total: 17 packages
- Passed: 17
- Failed: 0

## Build Status
Pass - no warnings

## Pre-existing Issues
None - all tests pass

## Context
This issue integrates the library dependency system into the installer:
- Detect `dependencies` in recipe metadata
- Install library dependencies to `$TSUKU_HOME/libs/`
- Track `used_by` in state.json
- Reuse existing library installations

Dependencies: #214, #215, #216 (all closed)
Design reference: docs/DESIGN-relocatable-library-deps.md
