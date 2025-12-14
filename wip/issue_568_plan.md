# Issue 568/569 Implementation Plan

## Summary

Add NetworkValidator interface and implement RequiresNetwork() method on all actions to declare their network requirements for sandbox testing.

## Approach

Extend the existing BaseAction embedding pattern: add NetworkValidator interface with RequiresNetwork() method, add default false implementation to BaseAction, and override to true in actions that fetch dependencies at runtime.

### Alternatives Considered
- Separate static map (like ActionDependencies): Rejected per design - keeps metadata separate from action code
- Add to Action interface directly: Would require updating all actions; interface method is cleaner

## Files to Modify
- `internal/actions/action.go` - Add NetworkValidator interface, add RequiresNetwork() to BaseAction

Actions needing RequiresNetwork() = true:
- `internal/actions/cargo_build.go` - Fetches crates.io dependencies
- `internal/actions/cargo_install.go` - Fetches from crates.io
- `internal/actions/go_build.go` - Fetches Go modules
- `internal/actions/go_install.go` - Fetches from Go proxy
- `internal/actions/cpan_install.go` - Fetches from CPAN
- `internal/actions/npm_install.go` - Fetches npm packages
- `internal/actions/pip_install.go` - Fetches from PyPI
- `internal/actions/pipx_install.go` - Fetches from PyPI
- `internal/actions/gem_install.go` - Fetches from RubyGems
- `internal/actions/system_packages.go` - apt/yum/brew fetch packages
- `internal/actions/run_command.go` - Conservative default (arbitrary commands)
- `internal/actions/nix_install.go` - Fetches from nix cache
- `internal/actions/nix_realize.go` - Fetches from nix cache

## Files to Create
- `internal/actions/action_test.go` - Add tests for NetworkValidator (extend existing file)

## Implementation Steps
- [ ] Add NetworkValidator interface and BaseAction.RequiresNetwork() to action.go
- [ ] Add RequiresNetwork() = true to ecosystem actions (cargo, go, npm, pip, gem, cpan)
- [ ] Add RequiresNetwork() = true to system package managers (apt, yum, brew)
- [ ] Add RequiresNetwork() = true to nix actions and run_command
- [ ] Add unit tests verifying network requirements for all actions

## Testing Strategy
- Unit tests: Verify each action returns correct RequiresNetwork() value
- Table-driven test covering all registered actions
- Verify BaseAction default is false

## Risks and Mitigations
- Missing an action: Use table-driven test that iterates all registered actions
- Incorrect classification: Follow design document classification table

## Success Criteria
- [ ] NetworkValidator interface exists with RequiresNetwork() method
- [ ] BaseAction.RequiresNetwork() returns false
- [ ] 13 actions return true (ecosystem + system + nix + run_command)
- [ ] All other actions return false (via BaseAction default)
- [ ] Tests pass, lint passes, build passes
