# Issue 559 Implementation Plan

## Summary

Create production git recipe using Homebrew bottles and test git-source recipe that builds from source. This validates the complete toolchain with git's complex dependencies: curl, openssl, zlib, expat.

## Approach

Following the established pattern from PR #661 (readline/sqlite):

1. **Production git.toml** - Homebrew bottle for fast user installs
2. **Test git-source.toml** - Build from source to validate configure_make + dependency resolution
3. **Test infrastructure** - Verification functions and optional integration test
4. **CI integration** - Add git-source to build-essentials.yml
5. **Design doc update** - Mark #559 as done

This follows the bash/bash-source and sqlite/sqlite-source pattern where production uses bottles, tests validate build toolchain.

### Alternatives Considered

- **Build git from source as production recipe**: Rejected - too slow for users, bottles are faster
- **Skip test recipe**: Rejected - need to validate configure_make with complex multi-dependency builds

## Files to Create

- `internal/recipe/recipes/g/git.toml` - Production recipe (homebrew bottle)
- `testdata/recipes/git-source.toml` - Test recipe (source build)

## Files to Modify

- `docs/DESIGN-dependency-provisioning.md` - Update mermaid diagram (#559 done)
- `.github/workflows/build-essentials.yml` - Add git-source test job
- `test/scripts/verify-tool.sh` - Add verify_git function

## Implementation Steps

- [ ] Create git.toml using homebrew action with curl dependency
- [ ] Create git-source.toml building from source with all dependencies
  - Dependencies: curl, openssl, zlib, expat
  - Use configure_make action
  - Set RPATH for git binary to find libraries
- [ ] Add verify_git function to verify-tool.sh
  - Test git --version
  - Test git clone of a small public repo
- [ ] Add git-source to CI test matrix in build-essentials.yml
- [ ] Update design doc mermaid diagram to mark #559 as done

## Testing Strategy

**Integration tests**:
1. **CI matrix test** - git-source installs on all 3 platforms
2. **Functionality test** - git --version works
3. **Clone test** - git clone works (validates curl integration)
4. **Dependency chain test** - Validates git → curl → openssl/zlib

**Manual verification**:
- Run tsuku install git locally
- Test git clone of a repository
- Verify ldd/otool shows tsuku-provided libraries

## Risks and Mitigations

**Risk**: Git has many dependencies and complex build
- **Mitigation**: Follow curl recipe pattern which already handles openssl + zlib

**Risk**: Git clone requires network access in tests
- **Mitigation**: Use a small, stable public repo for testing

**Risk**: RPATH setup may be complex with multiple libraries
- **Mitigation**: Follow sqlite-source pattern with multiple lib paths in RPATH

## Success Criteria

- [ ] git recipe uses homebrew bottles
- [ ] git-source builds from source with all dependencies
- [ ] git --version works
- [ ] git clone successfully clones a repository
- [ ] CI validates on all 3 platforms
- [ ] Design doc updated

## Open Questions

None - all patterns established and proven.
