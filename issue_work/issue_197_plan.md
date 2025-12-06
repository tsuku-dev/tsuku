# Issue 197 Implementation Plan

## Summary

Create version string transformation functions with input validation. Add a new `transform.go` file to the existing `internal/version` package.

## Approach

Implement transforms as pure functions with regex-based extraction. Return the original string on transform failure (with error for caller to handle). Use allowlist validation for security.

### Alternatives Considered

- Separate package: Not chosen - transforms are version-related and belong with version providers
- Parse into semver struct: Overcomplicated - we only need string extraction

## Files to Modify

None

## Files to Create

- `internal/version/transform.go` - Transform functions and validation
- `internal/version/transform_test.go` - Unit tests

## Implementation Steps

- [ ] Create transform.go with ValidateVersionString function
- [ ] Implement TransformVersion with raw/unknown format handling
- [ ] Implement semver transform (extract X.Y.Z)
- [ ] Implement semver_full transform (extract X.Y.Z[-pre][+build])
- [ ] Implement strip_v transform (remove leading v)
- [ ] Add comprehensive unit tests for all transforms
- [ ] Add edge case tests for validation

## Testing Strategy

- Unit tests: Each transform function with various inputs
- Edge cases: Empty string, invalid chars, max length, malformed versions
- Real-world examples: biome@2.3.8, v1.2.3, go1.21.0, 2.4.0-0

## Risks and Mitigations

- Risk: Regex complexity for semver extraction
  - Mitigation: Use well-tested semver regex patterns
- Risk: Performance with complex regex
  - Mitigation: Pre-compile regex at package init

## Success Criteria

- [ ] ValidateVersionString rejects invalid characters and overlength strings
- [ ] TransformVersion handles all four formats correctly
- [ ] Unknown formats fall back to raw (return original)
- [ ] All unit tests pass
- [ ] No regressions in existing tests

## Open Questions

None
