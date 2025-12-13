# Issue 449 Implementation Plan

## Summary

Enhance the existing `cpan_install` action to be a proper ecosystem primitive with deterministic execution features, aligning with patterns from cargo_build, go_build, and npm_exec.

## Approach

The cpan_install.go already exists with basic functionality. This issue requires:
1. Registering it as a primitive in decomposable.go (not decomposable)
2. Adding deterministic execution features (SOURCE_DATE_EPOCH, perl_version validation)
3. Adding cpanfile support with --installdeps
4. Adding mirror/offline support
5. Following patterns from other ecosystem primitives

### Alternatives Considered
- **Full Carton integration**: Too complex for initial implementation; spec suggests using cpanm with --mirror-only for offline/deterministic execution
- **Keep as-is**: Missing required deterministic features and primitive registration

## Files to Modify
- `internal/actions/cpan_install.go` - Add deterministic features, cpanfile support, mirror, perl_version validation
- `internal/actions/cpan_install_test.go` - Add tests for new functionality
- `internal/actions/decomposable.go` - Register cpan_install as primitive

## Files to Create
- None (cpan_install.go already exists)

## Implementation Steps
- [ ] Register cpan_install as ecosystem primitive in decomposable.go
- [ ] Add SOURCE_DATE_EPOCH environment variable for reproducibility
- [ ] Add perl_version parameter with validation (like rust_version in cargo_build)
- [ ] Add cpanfile parameter with --installdeps support
- [ ] Add mirror parameter with --mirror and --mirror-only flags
- [ ] Add offline parameter for security (like cargo_build --offline)
- [ ] Use GetBool helper for boolean params (consistency with other primitives)
- [ ] Update tests for new parameters
- [ ] Run tests and verify all pass

## Testing Strategy
- Unit tests: Validate parameter handling, version validation, mirror flags
- Integration tests: Already exist, verify they still pass
- Manual verification: Not required (cpan_install already functional)

## Risks and Mitigations
- **Backward compatibility**: New parameters are optional with sensible defaults
- **Missing cpanm/perl**: Already handled with clear error messages

## Success Criteria
- [ ] cpan_install registered as primitive in decomposable.go
- [ ] SOURCE_DATE_EPOCH set for deterministic builds
- [ ] perl_version validation with helpful error messages
- [ ] cpanfile support with --installdeps
- [ ] mirror/offline support for security
- [ ] All tests pass
- [ ] Consistent patterns with cargo_build, go_build, npm_exec

## Open Questions
None - requirements are clear from issue and ecosystem spec.
