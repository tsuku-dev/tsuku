# Issue 129 Summary

## What Was Implemented

MetaCPAN version provider for resolving CPAN distribution versions. This enables `tsuku versions <distribution>` for Perl/CPAN tools, following the pattern established by RubyGems, PyPI, npm, and crates.io providers.

## Changes Made

- `internal/version/metacpan.go`: API interaction logic
  - `isValidMetaCPANDistribution()` validates distribution names (rejects module names with `::`)
  - `normalizeModuleToDistribution()` converts module names to distribution format
  - `ResolveMetaCPAN()` queries `GET /release/{distribution}` for latest version
  - `ListMetaCPANVersions()` queries `POST /release/_search` for version history

- `internal/version/provider_metacpan.go`: Provider struct
  - Implements `VersionLister` interface (ListVersions, ResolveLatest, ResolveVersion, SourceDescription)
  - Supports fuzzy version matching (e.g., "3.7" matches "3.7.0")

- `internal/version/provider_factory.go`: Strategy registration
  - `MetaCPANSourceStrategy` for explicit `source="metacpan"` recipes
  - `InferredMetaCPANStrategy` for recipes with `cpan_install` action

- `internal/version/resolver.go`: Registry URL injection
  - Added `metacpanRegistryURL` field for test injection
  - Added `NewWithMetaCPANRegistry()` constructor

- `internal/version/metacpan_test.go`: Comprehensive test suite
  - Distribution name validation tests
  - Mock server tests for API responses
  - Error handling tests (404, 429, invalid content-type, HTTPS enforcement)
  - Strategy matching tests

## Key Decisions

- **Separate metacpan.go and provider_metacpan.go files**: Follows existing pattern (rubygems.go + provider_rubygems.go)
- **POST /_search for version listing**: Elasticsearch query enables filtering and sorting; more flexible than paginated GET
- **10MB response size limit**: Matches RubyGems limit; sufficient for any reasonable distribution

## Trade-offs Accepted

- **No automatic module-to-distribution conversion**: Users must provide distribution names (App-Ack, not App::Ack). Error message suggests conversion.

## Test Coverage

- New tests added: 17 test functions covering validation, API calls, error handling, and strategy matching
- All tests pass with mock TLS server

## Known Limitations

- Does not automatically discover executable names from distributions (deferred to cpan_install action)
- Module names with `::` are rejected (user must convert to distribution format)

## Future Improvements

- Could add caching for version lists (low priority - API is fast)
- Integration test in issue #144 will verify end-to-end functionality
