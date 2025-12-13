# Issue 445 Summary

## What Was Implemented

Implemented the `npm_exec` ecosystem primitive action for deterministic npm/Node.js build execution. The action uses lockfile enforcement, isolated caching, and reproducible build flags to achieve deterministic builds.

## Changes Made

- `internal/actions/npm_exec.go`: New action implementing deterministic npm execution
  - Parameter validation for source_dir, command, and optional settings
  - Node.js version validation with semver constraints (>=, >, exact, .x patterns)
  - npm ci with security hardening (--ignore-scripts, --no-audit, --no-fund)
  - SOURCE_DATE_EPOCH for reproducible timestamps
  - Isolated npm cache directory

- `internal/actions/npm_exec_test.go`: Comprehensive unit tests
  - Parameter validation tests
  - Lockfile enforcement tests
  - Version parsing and comparison tests

- `internal/actions/action.go`: Registered NpmExecAction in init()

- `internal/actions/decomposable.go`: Added npm_exec to primitives registry

- `internal/actions/decomposable_test.go`: Updated primitive tests for new count

- `internal/actions/dependencies.go`: Added npm_exec with nodejs dependencies

## Key Decisions

- **use_lockfile defaults to true**: Enforces determinism by default
- **ignore_scripts defaults to true**: Security-first approach to prevent arbitrary code execution
- **Version constraint support**: Supports semver patterns (>=18.0.0, 18.x, exact versions)

## Trade-offs Accepted

- **Native addons not handled**: npm_exec does not explicitly handle native addon compilation, which remains non-deterministic across platforms

## Test Coverage

- New tests added: 11 test cases covering parameter validation, version parsing, and lockfile behavior

## Known Limitations

- Node.js must be pre-installed (use nodejs recipe as dependency)
- Native addon packages may produce non-deterministic results
- No npm version constraint validation (only node version)
