# Issue 686 Baseline

## Environment
- Date: 2025-12-27
- Branch: docs/platform-tuple-support
- Base commit: e746f5df038b0e0eb83f136c43508168f7d5a30a

## Test Results
- Total: All packages tested
- Passed: All tests passing
- Failed: 0

Test output:
```
ok  	github.com/tsukumogami/tsuku	9.852s
ok  	github.com/tsukumogami/tsuku/cmd/tsuku	0.034s
ok  	github.com/tsukumogami/tsuku/internal/actions	1.518s
ok  	github.com/tsukumogami/tsuku/internal/builders	1.489s
ok  	github.com/tsukumogami/tsuku/internal/executor	6.358s
ok  	github.com/tsukumogami/tsuku/internal/recipe	(cached)
ok  	github.com/tsukumogami/tsuku/internal/sandbox	1.079s
ok  	github.com/tsukumogami/tsuku/internal/validate	0.458s
(other packages cached)
```

## Build Status
- Status: PASS
- Command: `go build -o tsuku ./cmd/tsuku`
- No warnings

## Coverage
Not measured in baseline (will compare after implementation)

## Pre-existing Issues
None - all tests passing, build succeeds

## Scope
This issue implements platform tuple support for install_guide field only:
- Extend install_guide to accept os/arch format keys (e.g., "darwin/arm64")
- Hierarchical fallback: exact tuple → OS key → fallback key
- Update validation to ensure platform coverage
- Maintain backwards compatibility with existing OS-only keys

Step-level when clause support is tracked separately in issue #690.
