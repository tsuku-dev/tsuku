# Issue 394 Summary

## What Was Implemented

Created tracking infrastructure for reducing merge conflicts by analyzing high-churn files and creating focused refactoring issues. Added code organization guidelines to CONTRIBUTING.md.

## Changes Made

- `wip/issue_394_baseline.md`: Established baseline test state
- `wip/issue_394_plan.md`: Created analysis and implementation plan
- `CONTRIBUTING.md`: Added "Code Organization" section with file size guidelines, split criteria, and refactoring patterns

## Issues Created

| Issue | File | Priority | Lines | 2mo Commits |
|-------|------|----------|-------|-------------|
| #397 | `internal/version/resolver.go` | HIGH | 1,269 | 13 |
| #398 | `cmd/tsuku/install.go` | MEDIUM | 603 | 17 |
| #399 | `internal/install/state.go` | MEDIUM | 613 | 9 |

## Key Decisions

- **Skip github_release.go**: Despite being 1,110 lines, it only had 5 commits in 2 months and has high internal cohesion due to LLM conversation flow complexity
- **Functional boundaries over LOC**: Proposed splits based on responsibility domains, not arbitrary line counts
- **Documentation over enforcement**: Added guidelines to CONTRIBUTING.md rather than automated tools

## Trade-offs Accepted

- **Guidelines not rules**: File size guidelines are soft limits (200-400 target, 600 max) rather than hard enforcement
- **Incremental refactoring**: Each file gets its own issue to allow gradual progress without large risky PRs

## Test Coverage

No code changes requiring tests - this is a documentation/issue creation task.

## Known Limitations

- Guidelines require manual adherence; no automated enforcement
- Churn metrics are based on 2-month window which may not capture seasonal patterns

## Future Improvements

- Consider adding git hooks to warn about large file changes
- Add metrics dashboard to track file sizes over time
